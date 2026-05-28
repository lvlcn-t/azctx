package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/config"
)

// contextItem implements list.Item for a context entry.
type contextItem struct {
	name         string
	tenant       string
	tenantID     string
	credential   string
	credType     string
	subscription string
	current      bool
}

func (i *contextItem) Title() string {
	prefix := "  "
	if i.current {
		prefix = "* "
	}
	return prefix + i.name
}

func (i *contextItem) Description() string {
	parts := []string{i.tenant}
	if i.subscription != "" {
		parts = append(parts, i.subscription)
	}
	return strings.Join(parts, " | ")
}

func (i *contextItem) FilterValue() string { return i.name }

// buildItems creates list items from a config.
func buildItems(cfg *config.Config) []list.Item {
	items := make([]list.Item, 0, len(cfg.Contexts))
	for _, ctx := range cfg.Contexts {
		item := &contextItem{
			name:         ctx.Name,
			tenant:       ctx.Details.Tenant,
			credential:   ctx.Details.Credential,
			subscription: ctx.Details.Subscription,
			current:      ctx.Name == cfg.CurrentContext,
		}
		if t, ok := cfg.TenantByName(ctx.Details.Tenant); ok {
			item.tenantID = t.Details.ID
		}
		if c, ok := cfg.CredentialByName(ctx.Details.Credential); ok {
			item.credType = string(c.Details.Type)
		}
		items = append(items, item)
	}
	return items
}
