package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/form"
	"github.com/lvlcn-t/azctx/tui/styles"
)

var (
	_ list.Item        = (*ContextItem)(nil)
	_ list.DefaultItem = (*ContextItem)(nil)
	_ details.Item     = (*ContextItem)(nil)
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

// contextForm builds the create or edit form for a context. On edit the name is
// pre-filled and locked; tenant and credential must reference existing entries.
// The tenant and credential help text lists the available names.
func contextForm(intent formIntent, store *config.Store, item details.Item) form.Model {
	var name, tenant, credential, subscription string
	title := "New context"
	readonly := false
	if ctx, ok := item.(*ContextItem); ok && intent == intentEdit {
		name = ctx.Name
		tenant = ctx.Tenant.Name
		credential = ctx.Credential.Name
		subscription = ctx.Subscription
		title = "Edit context"
		readonly = true
	}

	return form.New(title, []form.Field{
		{Key: fieldName, Label: labelName, Placeholder: "my-context", Value: name, Required: true, ReadOnly: readonly},
		{
			Key: fieldTenant, Label: "Tenant", Value: tenant, Required: true,
			Placeholder: strings.Join(tenantNames(store), ", "),
			Validate:    existsValidator("tenant", tenantNames(store)),
		},
		{
			Key: fieldCredential, Label: "Credential", Value: credential, Required: true,
			Placeholder: strings.Join(credentialNames(store), ", "),
			Validate:    existsValidator("credential", credentialNames(store)),
		},
		{Key: fieldSubscription, Label: "Subscription", Value: subscription, Placeholder: "optional"},
	})
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

// existsValidator rejects a value that is not one of the allowed names.
func existsValidator(kind string, allowed []string) func(string) error {
	return func(value string) error {
		for _, name := range allowed {
			if name == value {
				return nil
			}
		}
		return fmt.Errorf("%s %q does not exist", kind, value)
	}
}
