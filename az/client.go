package az

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/keyvault"
)

const (
	azInstallURL = "https://aka.ms/install-azure-cli"

	flagLogin                = "login"
	flagServicePrincipal     = "--service-principal"
	flagUsername             = "--username"
	flagTenant               = "--tenant"
	flagPassword             = "--password"
	flagFederatedToken       = "--federated-token"
	flagSubscription         = "--subscription"
	flagDisableSubDiscovery  = "--skip-subscription-discovery"
	flagAllowNoSubscriptions = "--allow-no-subscriptions"
)

var (
	// errCLIUnavailable indicates that the Azure CLI is not installed.
	errCLIUnavailable = errors.New("az CLI not found")
	// errTempCertMayRemain indicates that a temporary certificate file may not have been removed due to an os error.
	errTempCertMayRemain = errors.New("temporary cert file may remain on filesystem")
)

//go:generate go tool moq -out client_moq.go . CLI
type CLI interface {
	// WithCredential adds the given credential to the client for use in subsequent Login calls.
	WithCredential(credential *config.Credential) CLI
	// WithTenant adds the given tenant ID to the client for use in subsequent Login calls.
	WithTenant(tenantID string) CLI
	// WithSubscription adds the given subscription ID to the client for use in subsequent Login calls.
	WithSubscription(subscriptionID string) CLI
	// AllowNoSubscriptions configures the client to allow logging in without a subscription.
	AllowNoSubscriptions(allow bool) CLI
	// WithFederatedToken sets a pre-acquired federated token for workload identity login.
	WithFederatedToken(token string) CLI
	// Login authenticates Azure CLI for the given credential, tenant and optional subscription.
	Login(ctx context.Context) error
}

type client struct {
	azVersion            version
	credential           *config.Credential
	tenantID             string
	subscriptionID       string
	federatedToken       string
	allowNoSubscriptions bool
	kvResolver           *keyvault.Resolver
}

func NewClient(ctx context.Context) (CLI, error) {
	if err := ensureInstalled(); err != nil {
		return nil, err
	}

	v, err := azVersion(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to get az CLI version: %v\n", err)
		v = "0.0.0"
	}

	return &client{azVersion: v}, nil
}

func (c *client) WithCredential(credential *config.Credential) CLI {
	c.credential = credential
	return c
}

func (c *client) WithTenant(tenantID string) CLI {
	c.tenantID = tenantID
	return c
}

func (c *client) WithSubscription(subscriptionID string) CLI {
	c.subscriptionID = subscriptionID
	return c
}

func (c *client) AllowNoSubscriptions(allow bool) CLI {
	c.allowNoSubscriptions = allow
	return c
}

func (c *client) WithFederatedToken(token string) CLI {
	c.federatedToken = token
	return c
}

// Login authenticates Azure CLI for the given credential, tenant and optional subscription.
func (c *client) Login(ctx context.Context) error {
	if c.credential == nil {
		return errors.New("credential is required")
	}

	if c.tenantID == "" {
		return errors.New("tenant ID is required")
	}

	if err := c.credential.Validate(); err != nil {
		return fmt.Errorf("invalid credential %q: %w", c.credential.Name, err)
	}

	if err := c.login(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrLogin, err)
	}

	if !c.azVersion.supportsScopedLogin() && c.subscriptionID != "" {
		return az(ctx, "account", "set", flagSubscription, c.subscriptionID)
	}

	return nil
}

func (c *client) login(ctx context.Context) error {
	args := []string{flagLogin}
	args = appendIf(c.allowNoSubscriptions, args, flagAllowNoSubscriptions)

	switch c.credential.Details.Type {
	case config.CredentialTypeServicePrincipal:
		return c.loginServicePrincipal(ctx)
	case config.CredentialTypeUser:
		return withLoginExperienceOff(func() error {
			args = append(args, flagTenant, c.tenantID, "--output", "none")
			args = c.appendScopedLoginArgs(args)
			return az(ctx, args...)
		})
	case config.CredentialTypeManagedIdentity:
		args = append(args, "--identity")
		args = appendIf(c.credential.Details.Azure.ClientID != "", args, "--client-id", c.credential.Details.Azure.ClientID)
		args = c.appendScopedLoginArgs(args)
		return az(ctx, args...)
	case config.CredentialTypeWorkloadIdentity:
		return c.loginWithWorkloadIdentity(ctx)
	default:
		return fmt.Errorf("unsupported credential type %q", c.credential.Details.Type)
	}
}

// loginServicePrincipal performs az login with a service principal credential.
func (c *client) loginServicePrincipal(ctx context.Context) error {
	args := []string{
		flagLogin,
		flagServicePrincipal,
		flagUsername, c.credential.Details.Azure.ClientID,
		flagTenant, c.tenantID,
	}
	args = c.appendScopedLoginArgs(args)

	if c.credential.Details.Azure.ClientSecret != "" {
		return c.loginWithSecret(ctx, args)
	}

	return c.loginWithCert(ctx, args)
}

// loginWithSecret performs az login with a client secret, resolving from Key Vault if needed.
func (c *client) loginWithSecret(ctx context.Context, args []string) error {
	secret := c.credential.Details.Azure.ClientSecret
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

	return az(ctx, append(args, flagPassword, secret)...)
}

// loginWithCert performs az login with a client certificate, resolving from Key Vault if needed.
func (c *client) loginWithCert(ctx context.Context, args []string) error {
	certPath := c.credential.Details.Azure.ClientCertificatePath
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

// loginWithWorkloadIdentity performs az login with an OIDC federated token credential.
func (c *client) loginWithWorkloadIdentity(ctx context.Context) error {
	if c.federatedToken == "" {
		return errors.New("federated token is required for workload identity login")
	}

	args := []string{
		flagLogin,
		flagServicePrincipal,
		flagUsername, c.credential.Details.Azure.ClientID,
		flagTenant, c.tenantID,
		flagFederatedToken, c.federatedToken,
	}
	args = c.appendScopedLoginArgs(args)

	return az(ctx, args...)
}

func (c *client) appendScopedLoginArgs(args []string) []string {
	if !c.azVersion.supportsScopedLogin() {
		return args
	}

	args = appendIf(c.subscriptionID != "", args, flagSubscription, c.subscriptionID)
	args = append(args, flagDisableSubDiscovery)

	return args
}

// appendIf appends elems to slice if condition is true, otherwise returns slice unchanged.
func appendIf[T any](condition bool, slice []T, elems ...T) []T {
	if condition {
		return append(slice, elems...)
	}
	return slice
}
