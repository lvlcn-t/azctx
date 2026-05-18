package cmd

import (
	"fmt"
	"strings"

	"github.com/lvlcn-t/azctx/config"
)

// contextView is the output projection for a single azctx context.
type contextView struct {
	Name           string                `json:"name"`
	Tenant         string                `json:"tenant"`
	TenantID       string                `json:"tenantId,omitempty"`
	Credential     string                `json:"credential"`
	CredentialType config.CredentialType `json:"credentialType,omitempty"`
	Subscription   string                `json:"subscription,omitempty"`
	Current        bool                  `json:"current"`
}

// buildContextView builds a display model from a context entry and related refs.
func buildContextView(cfg *config.Config, context config.Context, currentContext string) contextView {
	if cfg == nil {
		cfg = &config.Config{}
	}

	view := contextView{
		Name:         context.Name,
		Tenant:       context.Context.Tenant,
		Credential:   context.Context.Credential,
		Subscription: context.Context.Subscription,
		Current:      context.Name == currentContext,
	}

	if tenant, found := cfg.TenantByName(context.Context.Tenant); found {
		view.TenantID = tenant.Tenant.ID
	}

	if credential, found := cfg.CredentialByName(context.Context.Credential); found {
		view.CredentialType = credential.Credential.Type
	}

	return view
}

// contextViewText formats one context view for text output.
func contextViewText(view *contextView) string {
	if view == nil {
		return ""
	}

	lines := []string{
		fmt.Sprintf("name: %s", view.Name),
		fmt.Sprintf("tenant: %s", view.Tenant),
		fmt.Sprintf("tenant-id: %s", emptyIfUnset(view.TenantID)),
		fmt.Sprintf("credential: %s", view.Credential),
		fmt.Sprintf("credential-type: %s", emptyIfUnset(string(view.CredentialType))),
		fmt.Sprintf("subscription: %s", emptyIfUnset(view.Subscription)),
	}

	return strings.Join(lines, "\n")
}

// contextTableRow formats one context view as a table row.
func contextTableRow(view *contextView) []string {
	if view == nil {
		return []string{"", "", "", "", "", "", ""}
	}

	currentMarker := ""
	if view.Current {
		currentMarker = "*"
	}

	return []string{
		currentMarker,
		view.Name,
		view.Tenant,
		emptyIfUnset(view.TenantID),
		view.Credential,
		emptyIfUnset(string(view.CredentialType)),
		emptyIfUnset(view.Subscription),
	}
}

// emptyIfUnset renders empty values with a visible placeholder.
func emptyIfUnset(value string) string {
	if strings.TrimSpace(value) == "" {
		return "<unset>"
	}

	return value
}
