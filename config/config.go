package config

import (
	"errors"
	"fmt"

	"github.com/lvlcn-t/azctx/semver"
)

const (
	APIVersion = "azctx.lvlcn-t.dev/v1alpha1"
	Kind       = "Config"
)

// Config is the root azctx configuration document.
type Config struct { //nolint:govet // field alignment is intentional
	APIVersion     semver.Version `yaml:"apiVersion" json:"apiVersion"`
	Kind           string         `yaml:"kind" json:"kind"`
	Tenants        []Tenant       `yaml:"tenants" json:"tenants"`
	Credentials    []Credential   `yaml:"credentials" json:"credentials"`
	Contexts       []Context      `yaml:"contexts" json:"contexts"`
	CurrentContext string         `yaml:"current-context,omitempty" json:"current-context,omitempty"`
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

// DeleteTenant removes a tenant by name.
func (cfg *Config) DeleteTenant(name string) bool {
	for index, tenant := range cfg.Tenants {
		if tenant.Name == name {
			cfg.Tenants = append(cfg.Tenants[:index], cfg.Tenants[index+1:]...)
			return true
		}
	}

	return false
}

// DeleteCredential removes a credential by name.
func (cfg *Config) DeleteCredential(name string) bool {
	for index, credential := range cfg.Credentials {
		if credential.Name == name {
			cfg.Credentials = append(cfg.Credentials[:index], cfg.Credentials[index+1:]...)
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

// RenameTenant renames an existing tenant. It does not update contexts that
// reference the tenant; use RetargetTenant for that.
func (cfg *Config) RenameTenant(oldName, newName string) bool {
	for index, tenant := range cfg.Tenants {
		if tenant.Name == oldName {
			cfg.Tenants[index].Name = newName
			return true
		}
	}

	return false
}

// RenameCredential renames an existing credential. It does not update contexts
// that reference the credential; use RetargetCredential for that.
func (cfg *Config) RenameCredential(oldName, newName string) bool {
	for index, credential := range cfg.Credentials {
		if credential.Name == oldName {
			cfg.Credentials[index].Name = newName
			return true
		}
	}

	return false
}

// RetargetTenant rewrites every context that references oldName to reference
// newName, returning the number of contexts updated.
func (cfg *Config) RetargetTenant(oldName, newName string) int {
	updated := 0
	for index := range cfg.Contexts {
		if cfg.Contexts[index].Details.Tenant == oldName {
			cfg.Contexts[index].Details.Tenant = newName
			updated++
		}
	}

	return updated
}

// RetargetCredential rewrites every context that references oldName to
// reference newName, returning the number of contexts updated.
func (cfg *Config) RetargetCredential(oldName, newName string) int {
	updated := 0
	for index := range cfg.Contexts {
		if cfg.Contexts[index].Details.Credential == oldName {
			cfg.Contexts[index].Details.Credential = newName
			updated++
		}
	}

	return updated
}

// ContextsReferencingTenant returns the names of contexts referencing the
// tenant, in declaration order.
func (cfg *Config) ContextsReferencingTenant(name string) []string {
	var names []string
	for _, context := range cfg.Contexts {
		if context.Details.Tenant == name {
			names = append(names, context.Name)
		}
	}

	return names
}

// ContextsReferencingCredential returns the names of contexts referencing the
// credential, in declaration order.
func (cfg *Config) ContextsReferencingCredential(name string) []string {
	var names []string
	for _, context := range cfg.Contexts {
		if context.Details.Credential == name {
			names = append(names, context.Name)
		}
	}

	return names
}

// Merge merges another config into this config.
func (cfg *Config) Merge(next *Config) error {
	if cfg == nil || next == nil {
		return nil
	}

	if !cfg.APIVersion.Compatible(next.APIVersion) {
		return fmt.Errorf(
			"cannot merge configs with incompatible API versions: %q vs %q",
			cfg.APIVersion,
			next.APIVersion,
		)
	}

	if cfg.CurrentContext == "" && next.CurrentContext != "" {
		cfg.CurrentContext = next.CurrentContext
	}

	cfg.mergeTenants(next.Tenants)
	cfg.mergeCredentials(next.Credentials)
	cfg.mergeContexts(next.Contexts)
	return nil
}

// ValidateContextReferences validates a context and its references.
func (cfg *Config) ValidateContextReferences(context Context) error {
	if context.Name == "" {
		return errors.New("context name is required")
	}

	if context.Details.Tenant == "" {
		return errors.New("context tenant is required")
	}

	if context.Details.Credential == "" {
		return errors.New("context credential is required")
	}

	if _, found := cfg.TenantByName(context.Details.Tenant); !found {
		return fmt.Errorf("tenant %q does not exist", context.Details.Tenant)
	}

	if _, found := cfg.CredentialByName(context.Details.Credential); !found {
		return fmt.Errorf("credential %q does not exist", context.Details.Credential)
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
