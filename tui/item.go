package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/config"
)

// contextItem represents a context entry in the TUI.
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
	marker := normalMarkerStyle.Render("○")
	if i.current {
		marker = currentMarkerStyle.Render("●")
	}
	return marker + " " + i.name
}

func (i *contextItem) Description() string {
	desc := i.tenant
	if i.subscription != "" {
		desc += " | " + i.subscription
	}
	return desc
}

func (i *contextItem) FilterValue() string { return i.name }

func (i *contextItem) detailTitle() string { return "Context: " + i.name }

func (i *contextItem) detailRows() []detailRow {
	return []detailRow{
		{"Tenant", i.tenant},
		{"Tenant ID", i.tenantID},
		{"Credential", i.credential},
		{"Credential Type", i.credType},
		{"Subscription", i.subscription},
		{"Current", fmt.Sprintf("%t", i.current)},
	}
}

// tenantItem represents a tenant entry in the TUI.
type tenantItem struct {
	name string
	id   string
}

func (i *tenantItem) Title() string       { return i.name }
func (i *tenantItem) Description() string { return i.id }
func (i *tenantItem) FilterValue() string { return i.name }

func (i *tenantItem) detailTitle() string { return "Tenant: " + i.name }

func (i *tenantItem) detailRows() []detailRow {
	return []detailRow{
		{"Name", i.name},
		{"ID", i.id},
	}
}

// credentialItem represents a credential entry in the TUI.
type credentialItem struct {
	name     string
	credType string
	clientID string
}

func (i *credentialItem) Title() string       { return i.name }
func (i *credentialItem) Description() string { return i.credType + " | " + i.clientID }
func (i *credentialItem) FilterValue() string { return i.name }

func (i *credentialItem) detailTitle() string { return "Credential: " + i.name }

func (i *credentialItem) detailRows() []detailRow {
	return []detailRow{
		{"Name", i.name},
		{"Type", i.credType},
		{"Client ID", i.clientID},
	}
}

// viewerContent is the interface for items that can be shown in the detail viewer.
type viewerContent interface {
	detailTitle() string
	detailRows() []detailRow
}

// detailRow is a label-value pair for the detail viewer.
type detailRow struct {
	label string
	value string
}

// buildContextItems creates list items from contexts in the config.
func buildContextItems(cfg *config.Config) []list.Item {
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

// buildTenantItems creates list items from tenants in the config.
func buildTenantItems(cfg *config.Config) []list.Item {
	items := make([]list.Item, 0, len(cfg.Tenants))
	for _, t := range cfg.Tenants {
		items = append(items, &tenantItem{
			name: t.Name,
			id:   t.Details.ID,
		})
	}
	return items
}

// buildCredentialItems creates list items from credentials in the config.
func buildCredentialItems(cfg *config.Config) []list.Item {
	items := make([]list.Item, 0, len(cfg.Credentials))
	for _, c := range cfg.Credentials {
		items = append(items, &credentialItem{
			name:     c.Name,
			credType: c.Details.Type.String(),
			clientID: c.Details.Azure.ClientID,
		})
	}
	return items
}
