package tui

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/keyvault"
)

// contextItem represents a context entry in the TUI.
type contextItem struct {
	*config.ResolvedContext
	current bool
}

func (i *contextItem) Title() string {
	marker := normalMarkerStyle.Render("○")
	if i.current {
		marker = currentMarkerStyle.Render("●")
	}
	return marker + " " + i.Name
}

func (i *contextItem) Description() string {
	desc := i.Name
	if i.Subscription != "" {
		desc += " | " + i.Subscription
	}
	return desc
}

func (i *contextItem) FilterValue() string { return i.Name }

func (i *contextItem) detailTitle() string { return "Context: " + i.Name }

func (i *contextItem) detailRows() []detailRow {
	return []detailRow{
		{"Tenant", i.Tenant.Name},
		{"Tenant ID", i.Tenant.Details.ID},
		{"Credential", i.Credential.Name},
		{"Credential Type", i.Credential.Details.Type.String()},
		{"Subscription", i.Subscription},
		{"Current", fmt.Sprintf("%t", i.current)},
	}
}

// tenantItem represents a tenant entry in the TUI.
type tenantItem config.Tenant

func (i *tenantItem) Title() string       { return i.Name }
func (i *tenantItem) Description() string { return i.Details.ID }
func (i *tenantItem) FilterValue() string { return i.Name }

func (i *tenantItem) detailTitle() string { return "Tenant: " + i.Name }

func (i *tenantItem) detailRows() []detailRow {
	return []detailRow{
		{"Name", i.Name},
		{"ID", i.Details.ID},
	}
}

// credentialItem represents a credential entry in the TUI.
type credentialItem config.Credential

func (i *credentialItem) Title() string { return i.Name }
func (i *credentialItem) Description() string {
	desc := i.Details.Type.String()
	if i.Details.Azure.ClientID != "" {
		desc += " | " + i.Details.Azure.ClientID
	}
	return desc
}
func (i *credentialItem) FilterValue() string { return i.Name }

func (i *credentialItem) detailTitle() string { return "Credential: " + i.Name }

func (i *credentialItem) detailRows() []detailRow {
	rows := []detailRow{
		{"Name", i.Name},
		{"Type", i.Details.Type.String()},
	}
	switch i.Details.Type {
	case config.CredentialTypeServicePrincipal:
		rows = append(rows, i.servicePrincipalRows()...)
	case config.CredentialTypeManagedIdentity:
		if i.Details.Azure.ClientID != "" {
			rows = append(rows, detailRow{"Client ID", i.Details.Azure.ClientID}) //nolint:goconst // not worth extracting
		}
	case config.CredentialTypeWorkloadIdentity:
		rows = append(rows, i.workloadIdentityRows()...)
	}

	return rows
}

func (i *credentialItem) servicePrincipalRows() []detailRow {
	rows := []detailRow{
		{"Client ID", i.Details.Azure.ClientID},
	}

	switch {
	case i.Details.Azure.ClientSecret != "":
		secret := strings.Repeat("*", len(i.Details.Azure.ClientSecret))
		if i.Details.Azure.ClientSecret != "" && keyvault.IsReference(i.Details.Azure.ClientSecret) {
			secret = i.Details.Azure.ClientSecret + " (key vault reference)"
		}
		rows = append(rows, detailRow{"Client Secret", secret})
	case i.Details.Azure.ClientCertificatePath != "":
		path := i.Details.Azure.ClientCertificatePath
		if keyvault.IsReference(i.Details.Azure.ClientCertificatePath) {
			path += " (key vault reference)"
		}
		rows = append(rows, detailRow{"Client Certificate Path", path})
	}

	return rows
}

func (i *credentialItem) workloadIdentityRows() []detailRow {
	rows := []detailRow{
		{"Client ID", i.Details.Azure.ClientID},
	}

	switch i.Details.Token.Source {
	case config.TokenSourceFile:
		rows = append(rows, detailRow{"Token File", i.Details.Token.File.Path})
	case config.TokenSourceOAuth2:
		rows = append(rows,
			detailRow{"OAuth Issuer", i.Details.Token.OAuth2.Issuer},
			detailRow{"OAuth Client ID", i.Details.Token.OAuth2.ClientID},
			detailRow{"OAuth Scopes", fmt.Sprintf("[%s]", strings.Join(i.Details.Token.OAuth2.Scopes, " "))},
			detailRow{"OAuth Redirect URI", cmp.Or(i.Details.Token.OAuth2.RedirectURI, "http://localhost:<random>")},
			detailRow{"OAuth PKCE", cmp.Or(i.Details.Token.OAuth2.PKCE.String(), "auto")},
		)
	}

	return rows
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
func buildContextItems(store *config.Store) []list.Item {
	cfg := store.Config
	items := make([]list.Item, 0, len(cfg.Contexts))
	for _, ctx := range cfg.Contexts {
		item := contextItem{
			ResolvedContext: &config.ResolvedContext{
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
			item.ResolvedContext = &resolved
		}
		items = append(items, &item)
	}
	return items
}

// buildTenantItems creates list items from tenants in the config.
func buildTenantItems(cfg *config.Config) []list.Item {
	items := make([]list.Item, 0, len(cfg.Tenants))
	for _, t := range cfg.Tenants {
		i := tenantItem(t)
		items = append(items, &i)
	}
	return items
}

// buildCredentialItems creates list items from credentials in the config.
func buildCredentialItems(cfg *config.Config) []list.Item {
	items := make([]list.Item, 0, len(cfg.Credentials))
	for _, c := range cfg.Credentials {
		i := credentialItem(c)
		items = append(items, &i)
	}
	return items
}
