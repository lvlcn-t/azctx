// Package keyvault resolves keyvault:// URI references to their secret
// or certificate values using Azure SDK and ambient credentials.
package keyvault

import (
	"context"
	"fmt"
	"strings"
)

const uriScheme = "keyvault://"

const (
	// ObjectTypeSecrets is the Key Vault object type for secrets.
	ObjectTypeSecrets = "secrets"
	// ObjectTypeCertificates is the Key Vault object type for certificates.
	ObjectTypeCertificates = "certificates"
)

// maxURIParts is the maximum number of path segments in a keyvault URI
// (vault/type/name/version).
const maxURIParts = 4

// Reference represents a parsed keyvault:// URI.
type Reference struct {
	VaultName  string
	ObjectType string // "secrets" or "certificates"
	ObjectName string
	Version    string // optional
}

// VaultURL returns the full vault URL for a reference.
func (r Reference) VaultURL() string {
	return fmt.Sprintf("https://%s.vault.azure.net", r.VaultName)
}

// IsReference reports whether a value is a keyvault:// URI.
func IsReference(value string) bool {
	return strings.HasPrefix(value, uriScheme)
}

// Parse parses a keyvault:// URI into its components.
// Format: keyvault://<vault>/<secrets|certificates>/<name>[/<version>]
func Parse(uri string) (Reference, error) {
	if !IsReference(uri) {
		return Reference{}, fmt.Errorf("not a keyvault URI: %q", uri)
	}

	path := strings.TrimPrefix(uri, uriScheme)
	parts := strings.Split(path, "/")

	if len(parts) < 3 || len(parts) > maxURIParts {
		return Reference{}, fmt.Errorf(
			"invalid keyvault URI %q: expected keyvault://<vault>/<secrets|certificates>/<name>[/<version>]",
			uri,
		)
	}

	ref := Reference{
		VaultName:  parts[0],
		ObjectType: parts[1],
		ObjectName: parts[2],
	}

	if len(parts) == maxURIParts {
		ref.Version = parts[3]
	}

	switch ref.ObjectType {
	case ObjectTypeSecrets, ObjectTypeCertificates:
	default:
		return Reference{}, fmt.Errorf(
			"invalid keyvault URI %q: object type must be \"secrets\" or \"certificates\", got %q",
			uri, ref.ObjectType,
		)
	}

	return ref, nil
}

// Client abstracts Key Vault secret/certificate retrieval.
//
//go:generate go tool moq -out client_moq.go . Client
type Client interface {
	GetSecret(ctx context.Context, ref Reference) (string, error)
}

// Resolver resolves keyvault:// URIs to their values.
type Resolver struct {
	client Client
}

// NewResolver creates a Resolver with the given Client.
func NewResolver(client Client) *Resolver {
	return &Resolver{client: client}
}

// Resolve resolves a keyvault:// URI to its string value.
func (r *Resolver) Resolve(ctx context.Context, uri string) (string, error) {
	ref, err := Parse(uri)
	if err != nil {
		return "", err
	}

	return r.client.GetSecret(ctx, ref)
}

// ResolveCertificateBytes parses a keyvault:// URI and returns the
// certificate PEM bytes.
func (r *Resolver) ResolveCertificateBytes(ctx context.Context, uri string) ([]byte, error) {
	ref, err := Parse(uri)
	if err != nil {
		return nil, err
	}

	val, err := r.client.GetSecret(ctx, ref)
	if err != nil {
		return nil, err
	}

	return []byte(val), nil
}
