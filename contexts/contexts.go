// Package contexts manages CRUD on azctx config entries (contexts, tenants, and
// credentials) and reports the current context. It is a leaf package: it must
// never import cmd or tui.
package contexts

import (
	"cmp"
	"fmt"

	"github.com/lvlcn-t/azctx/config"
)

// Manager creates, updates, renames, and deletes azctx config entries.
type Manager struct {
	// Writer persists config changes to disk.
	Writer config.Writer
}

// New returns a Manager wired with a production writer.
func New() *Manager {
	return &Manager{Writer: config.NewWriter()}
}

// SetContext creates or updates a context entry. When the context already
// exists in the merged config, next is merged onto the existing entry: empty
// tenant and credential are preserved, and the subscription is only replaced
// when subscriptionChanged is true. It reports whether an entry already existed.
func (m *Manager) SetContext(store *config.Store, next config.Context, subscriptionChanged bool) (bool, error) {
	if next.Name == "" {
		return false, ErrContextNameRequired
	}

	merged, existed := merge(&store.Config, next, subscriptionChanged)

	if err := store.Config.ValidateContextReferences(merged); err != nil {
		return existed, err
	}

	tenant, _ := store.Config.TenantByName(merged.Details.Tenant)
	if tenant.Details.ID == "" {
		return existed, fmt.Errorf("tenant %q is missing id", tenant.Name)
	}

	cred, _ := store.Config.CredentialByName(merged.Details.Credential)
	if err := cred.Validate(); err != nil {
		return existed, err
	}

	path := store.PathForContext(next.Name)
	cfg := store.FileConfig(path)
	cfg.UpsertContext(merged)

	return existed, m.Writer.Write(path, &cfg)
}

// SetTenant creates or updates a tenant entry. It reports whether an entry
// already existed.
func (m *Manager) SetTenant(store *config.Store, name, id string) (bool, error) {
	if name == "" {
		return false, ErrTenantNameRequired
	}

	if id == "" {
		return false, ErrTenantIDRequired
	}

	_, existed := store.Config.TenantByName(name)

	path := store.PathForTenant(name)
	cfg := store.FileConfig(path)
	cfg.UpsertTenant(config.Tenant{Name: name, Details: config.TenantDetails{ID: id}})

	return existed, m.Writer.Write(path, &cfg)
}

// SetCredential validates and then creates or updates a credential entry. It
// reports whether an entry already existed.
func (m *Manager) SetCredential(store *config.Store, cred *config.Credential) (bool, error) {
	if cred.Name == "" {
		return false, ErrCredentialNameRequired
	}

	if err := cred.Validate(); err != nil {
		return false, err
	}

	_, existed := store.Config.CredentialByName(cred.Name)

	path := store.PathForCredential(cred.Name)
	cfg := store.FileConfig(path)
	cfg.UpsertCredential(cred)

	return existed, m.Writer.Write(path, &cfg)
}

// RenameContext renames an existing context. When the renamed context was the
// active one and current-context lives in a different file, current-context is
// updated there too.
func (m *Manager) RenameContext(store *config.Store, oldName, newName string) error {
	if _, found := store.Config.ContextByName(oldName); !found {
		return fmt.Errorf("cannot rename context %q, it does not exist", oldName)
	}

	if _, found := store.Config.ContextByName(newName); found {
		return fmt.Errorf("cannot rename context %q, context %q already exists", oldName, newName)
	}

	path := store.PathForContext(oldName)
	cfg := store.FileConfig(path)
	if renamed := cfg.RenameContext(oldName, newName); !renamed {
		return fmt.Errorf("cannot rename context %q, it does not exist in %q", oldName, path)
	}

	if cfg.CurrentContext == oldName {
		cfg.CurrentContext = newName
	}

	if err := m.Writer.Write(path, &cfg); err != nil {
		return err
	}

	if store.Config.CurrentContext == oldName && path != store.PathForCurrentContext() {
		currentPath := store.PathForCurrentContext()
		currentConfig := store.FileConfig(currentPath)
		currentConfig.CurrentContext = newName
		if err := m.Writer.Write(currentPath, &currentConfig); err != nil {
			return err
		}
	}

	return nil
}

// DeleteResult reports the outcome of a delete call.
type DeleteResult struct {
	// Path is the config file the entry was removed from.
	Path string
	// WasActive reports whether the deleted context was the current one. It is
	// always false for tenants and credentials.
	WasActive bool
}

// DeleteContext removes a context entry and reports where it lived and whether
// it was the active context.
func (m *Manager) DeleteContext(store *config.Store, name string) (DeleteResult, error) {
	if _, found := store.Config.ContextByName(name); !found {
		return DeleteResult{}, fmt.Errorf("context %q not found", name)
	}

	path := store.PathForContext(name)
	cfg := store.FileConfig(path)
	if deleted := cfg.DeleteContext(name); !deleted {
		return DeleteResult{}, fmt.Errorf("context %q not found in %q", name, path)
	}

	if err := m.Writer.Write(path, &cfg); err != nil {
		return DeleteResult{}, err
	}

	return DeleteResult{Path: path, WasActive: store.Config.CurrentContext == name}, nil
}

// DeleteTenant removes a tenant entry and reports where it lived. Deleting a
// tenant that is still referenced by a context leaves that context dangling;
// callers should warn the user.
func (m *Manager) DeleteTenant(store *config.Store, name string) (DeleteResult, error) {
	_, found := store.Config.TenantByName(name)
	return m.deleteEntry(store, entryTarget{
		kind:    "tenant",
		name:    name,
		found:   found,
		path:    store.PathForTenant(name),
		deleteX: func(cfg *config.Config) bool { return cfg.DeleteTenant(name) },
	})
}

// DeleteCredential removes a credential entry and reports where it lived.
// Deleting a credential that is still referenced by a context leaves that
// context dangling; callers should warn the user.
func (m *Manager) DeleteCredential(store *config.Store, name string) (DeleteResult, error) {
	_, found := store.Config.CredentialByName(name)
	return m.deleteEntry(store, entryTarget{
		kind:    "credential",
		name:    name,
		found:   found,
		path:    store.PathForCredential(name),
		deleteX: func(cfg *config.Config) bool { return cfg.DeleteCredential(name) },
	})
}

// entryTarget describes a non-context entry to delete.
type entryTarget struct {
	deleteX func(cfg *config.Config) bool
	kind    string
	name    string
	path    string
	found   bool
}

// deleteEntry removes a tenant or credential entry. Contexts use DeleteContext
// directly because they additionally report whether the active context changed.
func (m *Manager) deleteEntry(store *config.Store, t entryTarget) (DeleteResult, error) {
	if !t.found {
		return DeleteResult{}, fmt.Errorf("%s %q not found", t.kind, t.name)
	}

	cfg := store.FileConfig(t.path)
	if deleted := t.deleteX(&cfg); !deleted {
		return DeleteResult{}, fmt.Errorf("%s %q not found in %q", t.kind, t.name, t.path)
	}

	if err := m.Writer.Write(t.path, &cfg); err != nil {
		return DeleteResult{}, err
	}

	return DeleteResult{Path: t.path}, nil
}

// merge resolves the effective context payload for an upsert. For a new context
// it returns next unchanged. For an existing one it overlays only the fields
// that were provided: non-empty tenant and credential, and the subscription
// only when subscriptionChanged is true.
func merge(cfg *config.Config, next config.Context, subscriptionChanged bool) (config.Context, bool) {
	existing, ok := cfg.ContextByName(next.Name)
	if !ok {
		return next, false
	}

	merged := existing
	merged.Details.Tenant = cmp.Or(next.Details.Tenant, existing.Details.Tenant)
	merged.Details.Credential = cmp.Or(next.Details.Credential, existing.Details.Credential)
	if subscriptionChanged {
		merged.Details.Subscription = next.Details.Subscription
	}

	return merged, true
}
