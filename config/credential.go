package config

import (
	"cmp"
	"errors"
	"fmt"
	"net/url"
)

// Credential represents a named Azure credential definition.
type Credential struct {
	Name       string            `yaml:"name" json:"name"`
	Credential CredentialDetails `yaml:"credential" json:"credential"`
}

// CredentialType identifies the kind of Azure credential.
type CredentialType string

// NewCredentialType parses a raw string into a CredentialType, validating that it's supported.
func NewCredentialType(raw string) (CredentialType, error) {
	v := CredentialType(raw)
	if v == "" {
		return "", ErrCredentialTypeRequired
	}

	switch v {
	case CredentialTypeUser, CredentialTypeServicePrincipal,
		CredentialTypeManagedIdentity, CredentialTypeWorkloadIdentity:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported credential type %q", raw)
	}
}

const (
	// CredentialTypeUser represents interactive user auth.
	CredentialTypeUser CredentialType = "user"
	// CredentialTypeServicePrincipal represents service principal auth.
	CredentialTypeServicePrincipal CredentialType = "service-principal"
	// CredentialTypeManagedIdentity represents managed identity auth.
	CredentialTypeManagedIdentity CredentialType = "managed-identity"
	// CredentialTypeWorkloadIdentity represents workload identity federation auth.
	CredentialTypeWorkloadIdentity CredentialType = "workload-identity"
)

// ErrCredentialTypeRequired indicates that a credential type was not provided.
var ErrCredentialTypeRequired = errors.New("credential type is required")

// CredentialDetails represents the details of an Azure credential.
type CredentialDetails struct {
	Type  CredentialType  `yaml:"type" json:"type"`
	Azure AzureCredential `yaml:"azure" json:"azure"`
	Token TokenDetails    `yaml:"token" json:"token"`
}

// AzureCredential represents Azure auth details for a credential.
type AzureCredential struct {
	ClientID              string `yaml:"client-id,omitempty" json:"clientId,omitempty"`
	ClientSecret          string `yaml:"client-secret,omitempty" json:"clientSecret,omitempty"`
	ClientCertificatePath string `yaml:"client-certificate-path,omitempty" json:"clientCertificatePath,omitempty"`
}

// TokenSource identifies the source of auth tokens for a credential.
type TokenSource string

const (
	// TokenSourceFile indicates tokens are sourced from a file.
	TokenSourceFile TokenSource = "file"
	// TokenSourceOAuth2 indicates tokens are sourced via an OAuth2 flow.
	TokenSourceOAuth2 TokenSource = "oauth2"
)

// TokenDetails represents auth token retrieval details for a credential.
type TokenDetails struct {
	Source TokenSource   `yaml:"source" json:"source"`
	OAuth2 *OAuth2Source `yaml:"oauth2,omitempty" json:"oauth2,omitempty"`
	File   *FileSource   `yaml:"file,omitempty" json:"file,omitempty"`
}

// FileSource represents a file-based token source.
type FileSource struct {
	Path string `yaml:"path" json:"path"`
}

// PKCE controls whether Proof Key for Code Exchange is used in the
// authorization code flow. The zero value is treated as [PKCEAuto].
type PKCE string

const (
	// PKCEAuto enables PKCE with S256 (default when omitted).
	PKCEAuto PKCE = "auto"
	// PKCEEnabled explicitly enables PKCE with S256.
	PKCEEnabled PKCE = "enabled"
	// PKCEDisabled explicitly disables PKCE.
	PKCEDisabled PKCE = "disabled"
)

// UnmarshalText implements [encoding.TextUnmarshaler]. It validates the
// raw value and normalizes empty strings to [PKCEAuto].
func (p *PKCE) UnmarshalText(b []byte) error {
	v := PKCE(b)
	switch v {
	case "", PKCEAuto, PKCEEnabled, PKCEDisabled:
		*p = cmp.Or(v, PKCEAuto)
		return nil
	default:
		return fmt.Errorf("invalid pkce value %q", v)
	}
}

// IsEnabled reports whether PKCE should be used. Returns true for all
// values except [PKCEDisabled].
func (p PKCE) IsEnabled() bool {
	return p != PKCEDisabled
}

func (p PKCE) String() string {
	return string(p)
}

// OAuth2Source represents an OAuth2-based token source configuration for
// the authorization code flow with OIDC discovery.
type OAuth2Source struct {
	Issuer      string   `yaml:"issuer" json:"issuer"`
	ClientID    string   `yaml:"client-id" json:"clientId"`
	Scopes      []string `yaml:"scopes" json:"scopes"`
	RedirectURI string   `yaml:"redirect-uri,omitempty" json:"redirectUri,omitempty"`
	PKCE        PKCE     `yaml:"pkce,omitempty" json:"pkce,omitempty"`
}

// Validate validates credential data for the selected credential type.
func (credential *Credential) Validate() error {
	var errs []error
	if credential == nil {
		return errors.New("credential is required")
	}

	if credential.Name == "" {
		errs = append(errs, errors.New("credential name is required"))
	}

	if credential.Credential.Type == "" {
		errs = append(errs, ErrCredentialTypeRequired)
	}

	if _, err := NewCredentialType(string(credential.Credential.Type)); err != nil {
		errs = append(errs, fmt.Errorf("invalid credential type: %w", err))
	}

	switch credential.Credential.Type {
	case CredentialTypeUser:
		return errors.Join(errs...)
	case CredentialTypeServicePrincipal:
		errs = append(errs, credential.validateServicePrincipal())
	case CredentialTypeManagedIdentity:
		return errors.Join(errs...)
	case CredentialTypeWorkloadIdentity:
		if credential.Credential.Azure.ClientID == "" {
			errs = append(errs, errors.New("workload identity credential requires client-id"))
		}

		if err := credential.Credential.Token.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid token configuration for workload identity credential: %w", err))
		}
	default:
		errs = append(errs, fmt.Errorf("unsupported credential type %q", credential.Credential.Type))
	}

	return errors.Join(errs...)
}

func (token *TokenDetails) Validate() error {
	var errs []error
	if token.Source == "" {
		return errors.New("token source is required")
	}

	switch token.Source {
	case TokenSourceFile:
		if token.File == nil || token.File.Path == "" {
			errs = append(errs, errors.New("file token source requires file path"))
		}
	case TokenSourceOAuth2:
		if token.OAuth2 == nil {
			errs = append(errs, errors.New("oauth2 token source requires oauth2 configuration"))
		}
		u, err := url.Parse(token.OAuth2.Issuer)
		if err != nil || !u.IsAbs() {
			errs = append(errs, errors.New("oauth2 token source requires valid issuer URL"))
		}
		if token.OAuth2.ClientID == "" {
			errs = append(errs, errors.New("oauth2 token source requires client-id"))
		}
		if len(token.OAuth2.Scopes) == 0 {
			errs = append(errs, errors.New("oauth2 token source requires at least one scope"))
		}
	default:
		return fmt.Errorf("unsupported token source %q", token.Source)
	}
	return errors.Join(errs...)
}

// validateServicePrincipal validates service-principal specific fields.
func (credential *Credential) validateServicePrincipal() error {
	var errs []error
	if credential.Credential.Azure.ClientID == "" {
		errs = append(errs, errors.New("service-principal credential requires client-id"))
	}

	hasSecret := credential.Credential.Azure.ClientSecret != ""
	hasCertificatePath := credential.Credential.Azure.ClientCertificatePath != ""

	if !hasSecret && !hasCertificatePath {
		errs = append(errs, errors.New("service-principal credential requires client-secret or client-certificate-path"))
	}

	if hasSecret && isKeyVaultRef(credential.Credential.Azure.ClientSecret) {
		if err := validateKeyVaultURI(credential.Credential.Azure.ClientSecret); err != nil {
			errs = append(errs, fmt.Errorf("invalid client-secret Key Vault reference: %w", err))
		}
	}

	if hasCertificatePath && isKeyVaultRef(credential.Credential.Azure.ClientCertificatePath) {
		if err := validateKeyVaultURI(credential.Credential.Azure.ClientCertificatePath); err != nil {
			errs = append(errs, fmt.Errorf("invalid client-certificate-path Key Vault reference: %w", err))
		}
	}

	return errors.Join(errs...)
}
