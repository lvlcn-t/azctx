package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/contexts"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/form"
	"github.com/lvlcn-t/azctx/tui/state"
	"github.com/lvlcn-t/azctx/tui/styles"
)

var (
	_ list.Item        = (*ContextItem)(nil)
	_ list.DefaultItem = (*ContextItem)(nil)
	_ details.Item     = (*ContextItem)(nil)
	_ entry            = (*ContextItem)(nil)
	_ activatable      = (*ContextItem)(nil)
)

type ContextItem struct {
	config.ResolvedContext
	current bool
}

func contextItems(store *config.Store) []list.Item {
	cfg := store.Config
	items := make([]list.Item, 0, len(cfg.Contexts))
	for _, ctx := range cfg.Contexts {
		item := &ContextItem{
			ResolvedContext: config.ResolvedContext{
				Name:                 ctx.Name,
				Tenant:               config.Tenant{Name: ctx.Details.Tenant},
				Credential:           config.Credential{Name: ctx.Details.Credential},
				Subscription:         ctx.Details.Subscription,
				AllowNoSubscriptions: ctx.Details.AllowNoSubscriptions,
			},
			current: ctx.Name == cfg.CurrentContext,
		}
		resolved, err := store.Resolve(ctx.Name)
		if err == nil {
			item.ResolvedContext = resolved
		}
		items = append(items, item)
	}
	return items
}

func (i *ContextItem) Title() string {
	marker := styles.NormalMarkerStyle.Render("○")
	if i.current {
		marker = styles.CurrentMarkerStyle.Render("●")
	}
	return marker + " " + i.Name
}

func (i *ContextItem) Description() string {
	desc := i.Name
	if i.Subscription != "" {
		desc += " | " + i.Subscription
	}
	return desc
}

func (i *ContextItem) FilterValue() string {
	return strings.Join([]string{i.Name, i.Subscription, i.Tenant.Name, i.Credential.Name}, " ")
}

func (i *ContextItem) Details() details.View {
	return details.View{
		Title: "Context: " + i.Name,
		Rows: []details.Row{
			{Label: "Tenant", Value: i.Tenant.Name},
			{Label: "Tenant ID", Value: i.Tenant.Details.ID},
			{Label: "Credential", Value: i.Credential.Name},
			{Label: "Credential Type", Value: i.Credential.Details.Type.String()},
			{Label: "Subscription", Value: i.Subscription},
			{Label: "Current", Value: fmt.Sprintf("%t", i.current)},
		},
	}
}

func (i *ContextItem) name() string { return i.Name }
func (i *ContextItem) blank() entry { return &ContextItem{} }

// activate makes this context the current selection and quits the TUI.
func (i *ContextItem) activate(s *state.UI) tea.Cmd {
	s.SelectContext(i.Name)
	return s.Quit()
}

// form builds the create or edit form for a context. On edit the name is
// pre-filled and locked; tenant and credential must reference existing entries,
// whose names are shown in the field placeholders.
func (i *ContextItem) form(intent formIntent, store *config.Store) form.Model {
	name, tenant, credential, subscription := "", "", "", ""
	title := "New context"
	readonly := false
	if intent == intentEdit {
		name = i.Name
		tenant = i.Tenant.Name
		credential = i.Credential.Name
		subscription = i.Subscription
		title = "Edit context"
		readonly = true
	}

	return form.New(title, []form.Field{
		{Key: fieldName, Label: labelName, Placeholder: "my-context", Value: name, Required: true, ReadOnly: readonly},
		{
			Key: fieldTenant, Label: "Tenant", Value: tenant, Required: true,
			Placeholder: strings.Join(tenantNames(store), ", "),
			Validate:    existsValidator(tenantNames(store)),
		},
		{
			Key: fieldCredential, Label: "Credential", Value: credential, Required: true,
			Placeholder: strings.Join(credentialNames(store), ", "),
			Validate:    existsValidator(credentialNames(store)),
		},
		{Key: fieldSubscription, Label: "Subscription", Value: subscription, Placeholder: "optional"},
	})
}

func (i *ContextItem) save(m *contexts.Manager, store *config.Store, sub submission) (string, error) {
	name := sub.values[fieldName]
	next := config.Context{
		Name: name,
		Details: config.ContextDetails{
			Tenant:       sub.values[fieldTenant],
			Credential:   sub.values[fieldCredential],
			Subscription: sub.values[fieldSubscription],
		},
	}

	switch sub.intent {
	case intentCreate:
		return "created context " + name, m.CreateContext(store, next)
	case intentEdit:
		// The form always carries the subscription value, so treat it as changed.
		return "updated context " + name, m.UpdateContext(store, next, true)
	case intentRename:
		return "renamed context " + i.Name + " to " + name, m.RenameContext(store, i.Name, name)
	default:
		return "", nil
	}
}

func (i *ContextItem) remove(m *contexts.Manager, store *config.Store) (string, error) {
	result, err := m.DeleteContext(store, i.Name)
	status := "deleted context " + i.Name
	if result.WasActive {
		status += " (warning: removed the active context; use a context to select a new one)"
	}
	return status, err
}

// tenantNames returns the configured tenant names.
func tenantNames(store *config.Store) []string {
	names := make([]string, 0, len(store.Config.Tenants))
	for _, t := range store.Config.Tenants {
		names = append(names, t.Name)
	}
	return names
}

// credentialNames returns the configured credential names.
func credentialNames(store *config.Store) []string {
	names := make([]string, 0, len(store.Config.Credentials))
	for _, c := range store.Config.Credentials {
		names = append(names, c.Name)
	}
	return names
}

// existsValidator rejects a value that is not one of the allowed names,
// returning an error that wraps errReferenceUnknown.
func existsValidator(allowed []string) func(string) error {
	return func(value string) error {
		for _, name := range allowed {
			if name == value {
				return nil
			}
		}
		return fmt.Errorf("%w: %q", errReferenceUnknown, value)
	}
}
