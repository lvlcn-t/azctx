package az

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/keyvault"
	"github.com/spf13/afero"
)

const azInstallURL = "https://aka.ms/install-azure-cli"

// errCLIUnavailable indicates that the Azure CLI is not installed.
var (
	errCLIUnavailable    = errors.New("az CLI not found")
	errTempCertMayRemain = errors.New("temporary cert file may remain on filesystem")
)

//go:generate go tool moq -out client_moq.go . CLI
type CLI interface {
	Login(ctx context.Context, credential *config.Credential, tenantID string) error
	SetSubscription(ctx context.Context, subscriptionID string) error
}

type client struct {
	kvResolver *keyvault.Resolver
}

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
		return c.loginWithSecret(ctx, args, credential.ClientSecret)
	}

	return c.loginWithCert(ctx, args, credential.ClientCertificatePath)
}

func (c *client) loginWithSecret(ctx context.Context, args []string, secret string) error {
	if keyvault.IsReference(secret) {
		kv, err := c.resolver()
		if err != nil {
			return fmt.Errorf("resolving client-secret from Key Vault: %w", err)
		}

		resolved, err := kv.Resolve(ctx, secret)
		if err != nil {
			return fmt.Errorf("resolving client-secret from Key Vault: %w", err)
		}

		secret = resolved
	}

	return az(ctx, append(args, "--password", secret)...)
}

func (c *client) loginWithCert(ctx context.Context, args []string, certPath string) error {
	if keyvault.IsReference(certPath) {
		resolver, err := c.resolver()
		if err != nil {
			return fmt.Errorf("resolving client-certificate from Key Vault: %w", err)
		}

		pem, err := resolver.ResolveCertificateBytes(ctx, certPath)
		if err != nil {
			return fmt.Errorf("resolving client-certificate from Key Vault: %w", err)
		}

		tmpFile, err := writeTempCert(pem)
		if err != nil {
			return fmt.Errorf("writing temporary certificate file: %w", err)
		}

		var once sync.Once
		defer func() {
			once.Do(func() {
				if err := removeTempCert(tmpFile); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to remove temp cert file %q: %v\n", tmpFile, err)
				}
			})
		}()
		go func() {
			// Since the root command handles signal cancellations and propagates the context,
			// this goroutine will not leak because the context will be canceled on interrupt,
			// triggering cleanup of the temp cert file.
			<-ctx.Done()
			once.Do(func() {
				if err := removeTempCert(tmpFile); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to remove temp cert file %q: %v\n", tmpFile, err)
				}
			})
		}()

		certPath = tmpFile
	}

	return az(ctx, append(args, "--certificate", certPath)...)
}

// resolver returns the keyvault resolver, creating it lazily on first use.
func (c *client) resolver() (*keyvault.Resolver, error) {
	if c.kvResolver != nil {
		return c.kvResolver, nil
	}

	kvClient, err := keyvault.NewAzureClient()
	if err != nil {
		return nil, err
	}

	c.kvResolver = keyvault.NewResolver(kvClient)

	return c.kvResolver, nil
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

// writeTempCert writes PEM bytes to a temporary file with restricted
// permissions and returns the file path.
func writeTempCert(pem []byte) (path string, err error) {
	f, err := afero.TempFile(fsys, "", "azctx-cert-*.pem")
	if err != nil {
		return "", err
	}

	name := f.Name()

	closed := false
	closeFile := func() error {
		if closed {
			return nil
		}

		// We're tracking the closed state before calling Close() because Close
		// must be treated as a one-shot operation. The io.Closer contract says
		// behavior after the first Close is undefined unless the implementation
		// documents otherwise:
		// https://pkg.go.dev/io#Closer
		//
		// On POSIX-like systems, close errors may be reported after the file
		// descriptor has already been released, so retrying Close can be unsafe:
		// the descriptor number may have been reused and a retry could close
		// something unrelated. Linux documents this explicitly:
		// https://man7.org/linux/man-pages/man2/close.2.html
		closed = true
		if cErr := f.Close(); cErr != nil {
			return fmt.Errorf("closing temp cert file %q: %w", name, cErr)
		}

		return nil
	}

	cleanup := func(cause error) error {
		cErr := closeFile()
		return errors.Join(cause, cErr, removeTempCert(name))
	}

	keep := false
	defer func() {
		if !keep {
			err = cleanup(err)
			path = ""
		}
	}()

	const certFileMode fs.FileMode = 0o600
	if err = fsys.Chmod(name, certFileMode); err != nil {
		return "", fmt.Errorf("setting permissions on temp cert file %q: %w", name, err)
	}

	if _, err = f.Write(pem); err != nil {
		return "", fmt.Errorf("writing to temp cert file %q: %w", name, err)
	}

	if err = closeFile(); err != nil {
		return "", err
	}

	keep = true
	return name, nil
}

// removeTempCert attempts to remove the temporary certificate file at the given path.
func removeTempCert(path string) error {
	if err := fsys.Remove(path); err != nil {
		return fmt.Errorf("%w: failed to remove temp cert file %q: %w",
			errTempCertMayRemain,
			path,
			err,
		)
	}

	return nil
}
