package az

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/afero"
)

const azInstallURL = "https://aka.ms/install-azure-cli"

// errCLIUnavailable indicates that the Azure CLI is not installed.
var errCLIUnavailable = errors.New("az CLI not found")

//go:generate go tool moq -out client_moq.go . CLI
type CLI interface {
	Login(ctx context.Context, credential *config.Credential, tenantID string) error
	SetSubscription(ctx context.Context, subscriptionID string) error
}

type client struct{}

func NewClient() (CLI, error) {
	if err := ensureInstalled(); err != nil {
		return nil, err
	}
	return &client{}, nil
}

// Login authenticates Azure CLI for the given credential and tenant.
func (c *client) Login(ctx context.Context, credential *config.Credential, tenantID string) error {
	if err := credential.Validate(); err != nil {
		return fmt.Errorf("invalid credential %q: %w", credential.Name, err)
	}

	switch credential.Type {
	case config.CredentialTypeServicePrincipal:
		return c.loginServicePrincipal(ctx, credential, tenantID)
	case config.CredentialTypeUser:
		return withLoginExperienceOff(func() error {
			return az(ctx, "login", "--tenant", tenantID, "--output", "none")
		})
	case config.CredentialTypeManagedIdentity:
		args := []string{"login", "--identity"}
		if credential.ClientID != "" {
			args = append(args, "--client-id", credential.ClientID)
		}

		return az(ctx, args...)
	case config.CredentialTypeOIDC:
		token, err := afero.ReadFile(fsys, credential.FederatedTokenFile)
		if err != nil {
			return fmt.Errorf("read federated token file %q: %w", credential.FederatedTokenFile, err)
		}

		trimmedToken := strings.TrimSpace(string(token))
		if trimmedToken == "" {
			return fmt.Errorf("federated token file %q is empty", credential.FederatedTokenFile)
		}

		return az(
			ctx,
			"login",
			"--service-principal",
			"--username", credential.ClientID,
			"--tenant", tenantID,
			"--federated-token", trimmedToken,
		)
	default:
		return fmt.Errorf("unsupported credential type %q", credential.Type)
	}
}

// SetSubscription sets the active subscription with a context.
func (c *client) SetSubscription(ctx context.Context, subscriptionID string) error {
	if strings.TrimSpace(subscriptionID) == "" {
		return nil
	}

	return az(ctx, "account", "set", "--subscription", subscriptionID)
}

func (c *client) loginServicePrincipal(ctx context.Context, credential *config.Credential, tenantID string) error {
	args := []string{
		"login",
		"--service-principal",
		"--username", credential.ClientID,
		"--tenant", tenantID,
	}

	if credential.ClientSecret != "" {
		args = append(args, "--password", credential.ClientSecret)
	} else {
		args = append(args, "--certificate", credential.ClientCertificatePath)
	}

	return az(ctx, args...)
}

// ensureInstalled validates that az is available in PATH.
func ensureInstalled() error {
	if _, err := exec.LookPath("az"); err != nil {
		return fmt.Errorf("%w in PATH. install it from %s", errCLIUnavailable, azInstallURL)
	}

	return nil
}

// az executes one Azure CLI command.
func az(ctx context.Context, args ...string) error {
	command := exec.CommandContext(ctx, "az", args...)

	var stderr bytes.Buffer
	command.Stdout = os.Stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if stderrText != "" {
			return fmt.Errorf("az %s failed: %w: %s", strings.Join(args, " "), err, stderrText)
		}

		return fmt.Errorf("az %s failed: %w", strings.Join(args, " "), err)
	}

	return nil
}
