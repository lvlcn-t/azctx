package tabs

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/tui/details"
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
	return i.Name
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
