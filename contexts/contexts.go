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

// CreateContext adds a new context entry. It fails with ErrContextExists if a
// context with the same name already exists.
func (m *Manager) CreateContext(store *config.Store, next config.Context) error {
	if next.Name == "" {
		return ErrContextNameRequired
	}

	if _, exists := store.Config.ContextByName(next.Name); exists {
		return fmt.Errorf("%w: %q", ErrContextExists, next.Name)
	}

	return m.writeContext(store, next)
}

// UpdateContext changes an existing context's references. It never changes the
// name and fails with ErrContextNotFound if the context does not exist. next is
// merged onto the existing entry: empty tenant and credential are preserved, and
// the subscription is only replaced when subscriptionChanged is true.
func (m *Manager) UpdateContext(store *config.Store, next config.Context, subscriptionChanged bool) error {
	if next.Name == "" {
		return ErrContextNameRequired
	}

	if _, exists := store.Config.ContextByName(next.Name); !exists {
		return fmt.Errorf("%w: %q", ErrContextNotFound, next.Name)
	}

	merged, _ := merge(&store.Config, next, subscriptionChanged)
	return m.writeContext(store, merged)
}

// SetContext creates the context if it does not exist, otherwise updates it. It
// reports whether an entry already existed. Prefer CreateContext/UpdateContext.
func (m *Manager) SetContext(store *config.Store, next config.Context, subscriptionChanged bool) (bool, error) {
	if _, existed := store.Config.ContextByName(next.Name); existed {
		return true, m.UpdateContext(store, next, subscriptionChanged)
	}
	return false, m.CreateContext(store, next)
}

// writeContext validates a context's references and persists it.
func (m *Manager) writeContext(store *config.Store, ctx config.Context) error {
	if err := store.Config.ValidateContextReferences(ctx); err != nil {
		return err
	}

	tenant, _ := store.Config.TenantByName(ctx.Details.Tenant)
	if tenant.Details.ID == "" {
		return fmt.Errorf("tenant %q is missing id", tenant.Name)
	}

	cred, _ := store.Config.CredentialByName(ctx.Details.Credential)
	if err := cred.Validate(); err != nil {
		return err
	}

	path := store.PathForContext(ctx.Name)
	cfg := store.FileConfig(path)
	cfg.UpsertContext(ctx)

	return m.Writer.Write(path, &cfg)
}

// CreateTenant adds a new tenant entry. It fails with ErrTenantExists if a
// tenant with the same name already exists.
func (m *Manager) CreateTenant(store *config.Store, name, id string) error {
	if err := validateTenant(name, id); err != nil {
		return err
	}

	if _, exists := store.Config.TenantByName(name); exists {
		return fmt.Errorf("%w: %q", ErrTenantExists, name)
	}

	return m.writeTenant(store, name, id)
}

// UpdateTenant changes an existing tenant's id. It never changes the name and
// fails with ErrTenantNotFound if the tenant does not exist. Use RenameTenant
// to change a tenant's name.
func (m *Manager) UpdateTenant(store *config.Store, name, id string) error {
	if err := validateTenant(name, id); err != nil {
		return err
	}

	if _, exists := store.Config.TenantByName(name); !exists {
		return fmt.Errorf("%w: %q", ErrTenantNotFound, name)
	}

	return m.writeTenant(store, name, id)
}

// RenameTenant renames a tenant and cascades the new name to every context that
// referenced it, across all affected config files.
func (m *Manager) RenameTenant(store *config.Store, oldName, newName string) (RenameResult, error) { //nolint:dupl // tenant/credential renames intentionally parallel; sharing more would hurt clarity
	return m.renameEntry(store, &renameTarget{
		oldName:   oldName,
		newName:   newName,
		emptyErr:  ErrTenantNameRequired,
		missErr:   ErrTenantNotFound,
		existsErr: ErrTenantExists,
		exists:    func(name string) bool { _, ok := store.Config.TenantByName(name); return ok },
		path:      store.PathForTenant(oldName),
		affected:  store.Config.ContextsReferencingTenant(oldName),
		rename:    func(c *config.Config) { c.RenameTenant(oldName, newName) },
		retarget:  func(c *config.Config) { c.RetargetTenant(oldName, newName) },
	})
}

// writeTenant upserts the tenant into its file and persists it.
func (m *Manager) writeTenant(store *config.Store, name, id string) error {
	path := store.PathForTenant(name)
	cfg := store.FileConfig(path)
	cfg.UpsertTenant(config.Tenant{Name: name, Details: config.TenantDetails{ID: id}})
	return m.Writer.Write(path, &cfg)
}

// SetTenant creates the tenant if it does not exist, otherwise updates it. It
// reports whether an entry already existed. Prefer CreateTenant/UpdateTenant.
func (m *Manager) SetTenant(store *config.Store, name, id string) (bool, error) {
	if _, existed := store.Config.TenantByName(name); existed {
		return true, m.UpdateTenant(store, name, id)
	}
	return false, m.CreateTenant(store, name, id)
}

// validateTenant checks the required tenant fields.
func validateTenant(name, id string) error {
	if name == "" {
		return ErrTenantNameRequired
	}
	if id == "" {
		return ErrTenantIDRequired
	}
	return nil
}

// CreateCredential adds a new credential entry after validation. It fails with
// ErrCredentialExists if a credential with the same name already exists.
func (m *Manager) CreateCredential(store *config.Store, cred *config.Credential) error {
	if err := validateCredential(cred); err != nil {
		return err
	}

	if _, exists := store.Config.CredentialByName(cred.Name); exists {
		return fmt.Errorf("%w: %q", ErrCredentialExists, cred.Name)
	}

	return m.writeCredential(store, cred)
}

// UpdateCredential changes an existing credential's details. It never changes
// the name and fails with ErrCredentialNotFound if the credential does not
// exist. Use RenameCredential to change a credential's name.
func (m *Manager) UpdateCredential(store *config.Store, cred *config.Credential) error {
	if err := validateCredential(cred); err != nil {
		return err
	}

	if _, exists := store.Config.CredentialByName(cred.Name); !exists {
		return fmt.Errorf("%w: %q", ErrCredentialNotFound, cred.Name)
	}

	return m.writeCredential(store, cred)
}

// RenameCredential renames a credential and cascades the new name to every
// context that referenced it, across all affected config files.
func (m *Manager) RenameCredential(store *config.Store, oldName, newName string) (RenameResult, error) { //nolint:dupl // tenant/credential renames intentionally parallel; sharing more would hurt clarity
	return m.renameEntry(store, &renameTarget{
		oldName:   oldName,
		newName:   newName,
		emptyErr:  ErrCredentialNameRequired,
		missErr:   ErrCredentialNotFound,
		existsErr: ErrCredentialExists,
		exists:    func(name string) bool { _, ok := store.Config.CredentialByName(name); return ok },
		path:      store.PathForCredential(oldName),
		affected:  store.Config.ContextsReferencingCredential(oldName),
		rename:    func(c *config.Config) { c.RenameCredential(oldName, newName) },
		retarget:  func(c *config.Config) { c.RetargetCredential(oldName, newName) },
	})
}

// renameTarget describes an entity rename with reference cascade.
type renameTarget struct {
	emptyErr  error
	missErr   error
	existsErr error
	exists    func(name string) bool
	rename    func(*config.Config)
	retarget  func(*config.Config)
	oldName   string
	newName   string
	path      string
	affected  []string
}

// renameEntry validates and applies a tenant or credential rename, cascading
// references to affected contexts.
func (m *Manager) renameEntry(store *config.Store, t *renameTarget) (RenameResult, error) {
	if t.newName == "" {
		return RenameResult{}, t.emptyErr
	}

	if !t.exists(t.oldName) {
		return RenameResult{}, fmt.Errorf("%w: %q", t.missErr, t.oldName)
	}

	if t.exists(t.newName) {
		return RenameResult{}, fmt.Errorf("%w: %q", t.existsErr, t.newName)
	}

	if err := m.renameAndRetarget(store, t); err != nil {
		return RenameResult{}, err
	}

	return RenameResult{Path: t.path, UpdatedContexts: t.affected}, nil
}

// writeCredential upserts the credential into its file and persists it.
func (m *Manager) writeCredential(store *config.Store, cred *config.Credential) error {
	path := store.PathForCredential(cred.Name)
	cfg := store.FileConfig(path)
	cfg.UpsertCredential(cred)
	return m.Writer.Write(path, &cfg)
}

// SetCredential creates the credential if it does not exist, otherwise updates
// it. It reports whether an entry already existed. Prefer
// CreateCredential/UpdateCredential.
func (m *Manager) SetCredential(store *config.Store, cred *config.Credential) (bool, error) {
	if _, existed := store.Config.CredentialByName(cred.Name); existed {
		return true, m.UpdateCredential(store, cred)
	}
	return false, m.CreateCredential(store, cred)
}

// validateCredential checks the required credential fields.
func validateCredential(cred *config.Credential) error {
	if cred.Name == "" {
		return ErrCredentialNameRequired
	}
	return cred.Validate()
}

// renameAndRetarget renames an entry in its own file and rewrites references in
// every file that owns an affected context. Each file is written exactly once:
// a file that both defines the entry and holds referencing contexts gets both
// mutations applied before it is persisted.
func (m *Manager) renameAndRetarget(store *config.Store, t *renameTarget) error {
	// Collect every file that needs a write.
	paths := map[string]struct{}{t.path: {}}
	for _, name := range t.affected {
		paths[store.PathForContext(name)] = struct{}{}
	}

	for path := range paths {
		cfg := store.FileConfig(path)
		if path == t.path {
			t.rename(&cfg)
		}
		t.retarget(&cfg)

		if err := m.Writer.Write(path, &cfg); err != nil {
			return err
		}
	}

	return nil
}

// RenameContext renames an existing context. When the renamed context was the
// active one and current-context lives in a different file, current-context is
// updated there too.
func (m *Manager) RenameContext(store *config.Store, oldName, newName string) error {
	if _, found := store.Config.ContextByName(oldName); !found {
		return fmt.Errorf("%w: %q", ErrContextNotFound, oldName)
	}

	if _, found := store.Config.ContextByName(newName); found {
		return fmt.Errorf("%w: %q", ErrContextExists, newName)
	}

	path := store.PathForContext(oldName)
	cfg := store.FileConfig(path)
	if renamed := cfg.RenameContext(oldName, newName); !renamed {
		return fmt.Errorf("%w: %q in %q", ErrContextNotFound, oldName, path)
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
	// OrphanedContexts lists contexts that referenced a deleted tenant or
	// credential and are now dangling. It is empty for context deletes.
	OrphanedContexts []string
	// WasActive reports whether the deleted context was the current one. It is
	// always false for tenants and credentials.
	WasActive bool
}

// RenameResult reports the outcome of a rename call.
type RenameResult struct {
	// Path is the config file the renamed entry lives in.
	Path string
	// UpdatedContexts lists contexts whose tenant or credential reference was
	// cascaded to the new name. It is empty for context renames.
	UpdatedContexts []string
}

// DeleteContext removes a context entry and reports where it lived and whether
// it was the active context.
func (m *Manager) DeleteContext(store *config.Store, name string) (DeleteResult, error) {
	if _, found := store.Config.ContextByName(name); !found {
		return DeleteResult{}, fmt.Errorf("%w: %q", ErrContextNotFound, name)
	}

	path := store.PathForContext(name)
	cfg := store.FileConfig(path)
	if deleted := cfg.DeleteContext(name); !deleted {
		return DeleteResult{}, fmt.Errorf("%w: %q in %q", ErrContextNotFound, name, path)
	}

	if err := m.Writer.Write(path, &cfg); err != nil {
		return DeleteResult{}, err
	}

	return DeleteResult{Path: path, WasActive: store.Config.CurrentContext == name}, nil
}

// DeleteTenant removes a tenant entry and reports where it lived and which
// contexts it leaves dangling. Deleting a tenant referenced by a context is
// allowed; callers should warn the user using OrphanedContexts.
func (m *Manager) DeleteTenant(store *config.Store, name string) (DeleteResult, error) {
	_, found := store.Config.TenantByName(name)
	return m.deleteEntry(store, &entryTarget{
		notFound: ErrTenantNotFound,
		name:     name,
		found:    found,
		path:     store.PathForTenant(name),
		remove:   func(cfg *config.Config) bool { return cfg.DeleteTenant(name) },
		orphans:  store.Config.ContextsReferencingTenant(name),
	})
}

// DeleteCredential removes a credential entry and reports where it lived and
// which contexts it leaves dangling. Deleting a credential referenced by a
// context is allowed; callers should warn the user using OrphanedContexts.
func (m *Manager) DeleteCredential(store *config.Store, name string) (DeleteResult, error) {
	_, found := store.Config.CredentialByName(name)
	return m.deleteEntry(store, &entryTarget{
		notFound: ErrCredentialNotFound,
		name:     name,
		found:    found,
		path:     store.PathForCredential(name),
		remove:   func(cfg *config.Config) bool { return cfg.DeleteCredential(name) },
		orphans:  store.Config.ContextsReferencingCredential(name),
	})
}

// entryTarget describes a non-context entry to delete.
type entryTarget struct {
	notFound error
	remove   func(cfg *config.Config) bool
	name     string
	path     string
	orphans  []string
	found    bool
}

// deleteEntry removes a tenant or credential entry. Contexts use DeleteContext
// directly because they additionally report whether the active context changed.
func (m *Manager) deleteEntry(store *config.Store, t *entryTarget) (DeleteResult, error) {
	if !t.found {
		return DeleteResult{}, fmt.Errorf("%w: %q", t.notFound, t.name)
	}

	cfg := store.FileConfig(t.path)
	if removed := t.remove(&cfg); !removed {
		return DeleteResult{}, fmt.Errorf("%w: %q in %q", t.notFound, t.name, t.path)
	}

	if err := m.Writer.Write(t.path, &cfg); err != nil {
		return DeleteResult{}, err
	}

	return DeleteResult{Path: t.path, OrphanedContexts: t.orphans}, nil
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
