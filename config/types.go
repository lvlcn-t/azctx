package config

import (
	"errors"
	"fmt"
)

// CredentialType identifies the kind of Azure credential.
type CredentialType string

const (
	// CredentialTypeServicePrincipal represents service principal auth.
	CredentialTypeServicePrincipal CredentialType = "service-principal"
	// CredentialTypeUser represents interactive user auth.
	CredentialTypeUser CredentialType = "user"
	// CredentialTypeManagedIdentity represents managed identity auth.
	CredentialTypeManagedIdentity CredentialType = "managed-identity"
	// CredentialTypeOIDC represents workload identity federation auth.
	CredentialTypeOIDC CredentialType = "oidc"
)

// ErrCredentialTypeRequired indicates that a credential type was not provided.
var ErrCredentialTypeRequired = errors.New("credential type is required")

// Config is the root azctx configuration document.
type Config struct {
	Tenants        []Tenant     `yaml:"tenants" json:"tenants"`
	Credentials    []Credential `yaml:"credentials" json:"credentials"`
	Contexts       []Context    `yaml:"contexts" json:"contexts"`
	CurrentContext string       `yaml:"current-context,omitempty" json:"current-context,omitempty"`
}

// Tenant represents a named Azure tenant definition.
type Tenant struct {
	Name string `yaml:"name" json:"name"`
	ID   string `yaml:"id" json:"id"`
}

// Credential represents a named Azure credential definition.
type Credential struct {
	Name                  string         `yaml:"name" json:"name"`
	Type                  CredentialType `yaml:"type" json:"type"`
	ClientID              string         `yaml:"client-id,omitempty" json:"clientId,omitempty"`
	ClientSecret          string         `yaml:"client-secret,omitempty" json:"clientSecret,omitempty"`
	ClientCertificatePath string         `yaml:"client-certificate-path,omitempty" json:"clientCertificatePath,omitempty"`
	FederatedTokenFile    string         `yaml:"federated-token-file,omitempty" json:"federatedTokenFile,omitempty"`
}

// Context represents a named azctx context entry.
type Context struct {
	Name         string `yaml:"name" json:"name"`
	Tenant       string `yaml:"tenant" json:"tenant"`
	Credential   string `yaml:"credential" json:"credential"`
	Subscription string `yaml:"subscription,omitempty" json:"subscription,omitempty"`
}

// ParseCredentialType parses and validates a credential type string.
func ParseCredentialType(raw string) (CredentialType, error) {
	v := CredentialType(raw)
	if v == "" {
		return "", ErrCredentialTypeRequired
	}

	switch v {
	case CredentialTypeServicePrincipal, CredentialTypeUser,
		CredentialTypeManagedIdentity, CredentialTypeOIDC:
		return v, nil
	default:
		return "", fmt.Errorf("unsupported credential type %q", raw)
	}
}

// TenantByName returns a tenant by name.
func (cfg *Config) TenantByName(name string) (Tenant, bool) {
	if cfg == nil {
		return Tenant{}, false
	}

	for _, tenant := range cfg.Tenants {
		if tenant.Name == name {
			return tenant, true
		}
	}

	return Tenant{}, false
}

// CredentialByName returns a credential by name.
func (cfg *Config) CredentialByName(name string) (Credential, bool) {
	if cfg == nil {
		return Credential{}, false
	}

	for _, credential := range cfg.Credentials {
		if credential.Name == name {
			return credential, true
		}
	}

	return Credential{}, false
}

// ContextByName returns a context by name.
func (cfg *Config) ContextByName(name string) (Context, bool) {
	if cfg == nil {
		return Context{}, false
	}

	for _, context := range cfg.Contexts {
		if context.Name == name {
			return context, true
		}
	}

	return Context{}, false
}

// UpsertTenant creates or replaces a tenant by name.
func (cfg *Config) UpsertTenant(tenant Tenant) {
	for index, current := range cfg.Tenants {
		if current.Name == tenant.Name {
			cfg.Tenants[index] = tenant
			return
		}
	}

	cfg.Tenants = append(cfg.Tenants, tenant)
}

// UpsertCredential creates or replaces a credential by name.
func (cfg *Config) UpsertCredential(credential *Credential) {
	if cfg == nil || credential == nil {
		return
	}

	for index, current := range cfg.Credentials {
		if current.Name == credential.Name {
			cfg.Credentials[index] = *credential
			return
		}
	}

	cfg.Credentials = append(cfg.Credentials, *credential)
}

// UpsertContext creates or replaces a context by name.
func (cfg *Config) UpsertContext(context Context) {
	for index, current := range cfg.Contexts {
		if current.Name == context.Name {
			cfg.Contexts[index] = context
			return
		}
	}

	cfg.Contexts = append(cfg.Contexts, context)
}

// DeleteContext removes a context by name.
func (cfg *Config) DeleteContext(name string) bool {
	for index, context := range cfg.Contexts {
		if context.Name == name {
			cfg.Contexts = append(cfg.Contexts[:index], cfg.Contexts[index+1:]...)
			return true
		}
	}

	return false
}

// RenameContext renames an existing context.
func (cfg *Config) RenameContext(oldName, newName string) bool {
	for index, context := range cfg.Contexts {
		if context.Name == oldName {
			cfg.Contexts[index].Name = newName
			return true
		}
	}

	return false
}

// Merge merges another config into this config.
func (cfg *Config) Merge(next *Config) {
	if cfg == nil || next == nil {
		return
	}

	if cfg.CurrentContext == "" && next.CurrentContext != "" {
		cfg.CurrentContext = next.CurrentContext
	}

	cfg.mergeTenants(next.Tenants)
	cfg.mergeCredentials(next.Credentials)
	cfg.mergeContexts(next.Contexts)
}

// ValidateContextReferences validates a context and its references.
func (cfg *Config) ValidateContextReferences(context Context) error {
	if context.Name == "" {
		return errors.New("context name is required")
	}

	if context.Tenant == "" {
		return errors.New("context tenant is required")
	}

	if context.Credential == "" {
		return errors.New("context credential is required")
	}

	if _, found := cfg.TenantByName(context.Tenant); !found {
		return fmt.Errorf("tenant %q does not exist", context.Tenant)
	}

	if _, found := cfg.CredentialByName(context.Credential); !found {
		return fmt.Errorf("credential %q does not exist", context.Credential)
	}

	return nil
}

// Validate validates credential data for the selected credential type.
func (credential *Credential) Validate() error {
	if credential == nil {
		return errors.New("credential is required")
	}

	if credential.Name == "" {
		return errors.New("credential name is required")
	}

	if credential.Type == "" {
		return ErrCredentialTypeRequired
	}

	if _, err := ParseCredentialType(string(credential.Type)); err != nil {
		return err
	}

	switch credential.Type {
	case CredentialTypeServicePrincipal:
		if credential.ClientID == "" {
			return errors.New("service-principal credential requires client-id")
		}

		hasSecret := credential.ClientSecret != ""
		hasCertificatePath := credential.ClientCertificatePath != ""
		if !hasSecret && !hasCertificatePath {
			return errors.New("service-principal credential requires client-secret or client-certificate-path")
		}
	case CredentialTypeUser:
		return nil
	case CredentialTypeManagedIdentity:
		return nil
	case CredentialTypeOIDC:
		if credential.ClientID == "" {
			return errors.New("oidc credential requires client-id")
		}

		if credential.FederatedTokenFile == "" {
			return errors.New("oidc credential requires federated-token-file")
		}
	default:
		return fmt.Errorf("unsupported credential type %q", credential.Type)
	}

	return nil
}

// mergeTenants merges tenant entries with first-wins semantics.
func (cfg *Config) mergeTenants(tenants []Tenant) {
	known := make(map[string]struct{}, len(cfg.Tenants))
	for _, tenant := range cfg.Tenants {
		known[tenant.Name] = struct{}{}
	}

	for _, tenant := range tenants {
		if _, exists := known[tenant.Name]; exists {
			continue
		}

		cfg.Tenants = append(cfg.Tenants, tenant)
		known[tenant.Name] = struct{}{}
	}
}

// mergeCredentials merges credential entries with first-wins semantics.
func (cfg *Config) mergeCredentials(credentials []Credential) {
	known := make(map[string]struct{}, len(cfg.Credentials))
	for _, credential := range cfg.Credentials {
		known[credential.Name] = struct{}{}
	}

	for _, credential := range credentials {
		if _, exists := known[credential.Name]; exists {
			continue
		}

		cfg.Credentials = append(cfg.Credentials, credential)
		known[credential.Name] = struct{}{}
	}
}

// mergeContexts merges context entries with first-wins semantics.
func (cfg *Config) mergeContexts(contexts []Context) {
	known := make(map[string]struct{}, len(cfg.Contexts))
	for _, context := range cfg.Contexts {
		known[context.Name] = struct{}{}
	}

	for _, context := range contexts {
		if _, exists := known[context.Name]; exists {
			continue
		}

		cfg.Contexts = append(cfg.Contexts, context)
		known[context.Name] = struct{}{}
	}
}
