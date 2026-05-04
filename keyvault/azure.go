package keyvault

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// ErrNoCredentials indicates that no ambient Azure credentials are available
// to authenticate to Key Vault.
var ErrNoCredentials = errors.New(
	"cannot resolve keyvault:// reference: no Azure credentials available. " +
		"Log in with \"azctx use\" using a non-Key Vault credential first",
)

// azureClient implements Client using the Azure SDK.
type azureClient struct {
	cred *azidentity.DefaultAzureCredential
}

// NewAzureClient creates a Client backed by DefaultAzureCredential.
func NewAzureClient() (Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoCredentials, err)
	}

	return &azureClient{cred: cred}, nil
}

// GetSecret fetches a secret value from Key Vault.
func (c *azureClient) GetSecret(ctx context.Context, ref Reference) (string, error) {
	client, err := azsecrets.NewClient(ref.VaultURL(), c.cred, nil)
	if err != nil {
		return "", fmt.Errorf("creating Key Vault secrets client: %w", err)
	}

	resp, err := client.GetSecret(ctx, ref.ObjectName, ref.Version, nil)
	if err != nil {
		return "", fmt.Errorf("fetching %s %q from vault %q: %w",
			ref.ObjectType, ref.ObjectName, ref.VaultName, err)
	}

	if resp.Value == nil {
		return "", fmt.Errorf("%s %q in vault %q has nil value",
			ref.ObjectType, ref.ObjectName, ref.VaultName)
	}

	return *resp.Value, nil
}
