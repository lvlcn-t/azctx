package tabs

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/keyvault"
	"github.com/lvlcn-t/azctx/tui/details"
)

var (
	_ list.Item        = (*CredentialItem)(nil)
	_ list.DefaultItem = (*CredentialItem)(nil)
	_ details.Item     = (*CredentialItem)(nil)
)

type CredentialItem struct{ config.Credential }

func credentialItems(s *config.Store) []list.Item {
	items := make([]list.Item, 0, len(s.Config.Credentials))
	for _, c := range s.Config.Credentials {
		i := CredentialItem{c}
		items = append(items, &i)
	}
	return items
}

func (i *CredentialItem) Title() string { return i.Name }
func (i *CredentialItem) Description() string {
	desc := i.Credential.Details.Type.String()
	if i.Credential.Details.Azure.ClientID != "" {
		desc += " | " + i.Credential.Details.Azure.ClientID
	}
	return desc
}

func (i *CredentialItem) FilterValue() string {
	typ := i.Credential.Details.Type
	d := []string{i.Name, typ.String()}
	switch typ {
	case config.CredentialTypeServicePrincipal, config.CredentialTypeManagedIdentity:
		d = append(d, i.Credential.Details.Azure.ClientID)
	case config.CredentialTypeWorkloadIdentity:
		d = append(d, []string{
			i.Credential.Details.Azure.ClientID,
			i.Credential.Details.Token.Source.String(),
			i.Credential.Details.Token.OAuth2.Issuer,
			i.Credential.Details.Token.OAuth2.ClientID,
			strings.Join(i.Credential.Details.Token.OAuth2.Scopes, " "),
		}...)
	}
	return strings.Join(d, " ")
}

func (i *CredentialItem) Details() details.View {
	return details.View{
		Title: "Credential: " + i.Name,
		Rows:  i.rows(),
	}
}

func (i *CredentialItem) rows() []details.Row {
	rows := []details.Row{
		{Label: "Name", Value: i.Name},
		{Label: "Type", Value: i.Credential.Details.Type.String()},
	}
	switch i.Credential.Details.Type {
	case config.CredentialTypeServicePrincipal:
		rows = append(rows, i.servicePrincipalRows()...)
	case config.CredentialTypeManagedIdentity:
		if i.Credential.Details.Azure.ClientID != "" {
			rows = append(rows, details.Row{Label: "Client ID", Value: i.Credential.Details.Azure.ClientID}) //nolint:goconst // not worth extracting
		}
	case config.CredentialTypeWorkloadIdentity:
		rows = append(rows, i.workloadIdentityRows()...)
	}

	return rows
}

func (i *CredentialItem) servicePrincipalRows() []details.Row {
	rows := []details.Row{
		{Label: "Client ID", Value: i.Credential.Details.Azure.ClientID},
	}

	switch {
	case i.Credential.Details.Azure.ClientSecret != "":
		secret := strings.Repeat("*", len(i.Credential.Details.Azure.ClientSecret))
		if i.Credential.Details.Azure.ClientSecret != "" && keyvault.IsReference(i.Credential.Details.Azure.ClientSecret) {
			secret = i.Credential.Details.Azure.ClientSecret + " (key vault reference)"
		}
		rows = append(rows, details.Row{Label: "Client Secret", Value: secret})
	case i.Credential.Details.Azure.ClientCertificatePath != "":
		path := i.Credential.Details.Azure.ClientCertificatePath
		if keyvault.IsReference(i.Credential.Details.Azure.ClientCertificatePath) {
			path += " (key vault reference)"
		}
		rows = append(rows, details.Row{Label: "Client Certificate Path", Value: path})
	}

	return rows
}

func (i *CredentialItem) workloadIdentityRows() []details.Row {
	rows := []details.Row{
		{Label: "Client ID", Value: i.Credential.Details.Azure.ClientID},
	}

	switch i.Credential.Details.Token.Source {
	case config.TokenSourceFile:
		rows = append(rows, details.Row{Label: "Token File", Value: i.Credential.Details.Token.File.Path})
	case config.TokenSourceOAuth2:
		rows = append(
			rows,
			details.Row{Label: "OAuth Issuer", Value: i.Credential.Details.Token.OAuth2.Issuer},
			details.Row{Label: "OAuth Client ID", Value: i.Credential.Details.Token.OAuth2.ClientID},
			details.Row{Label: "OAuth Scopes", Value: fmt.Sprintf("[%s]", strings.Join(i.Credential.Details.Token.OAuth2.Scopes, " "))},
			details.Row{Label: "OAuth Redirect URI", Value: cmp.Or(i.Credential.Details.Token.OAuth2.RedirectURI, "http://localhost:<random>")},
			details.Row{Label: "OAuth PKCE", Value: cmp.Or(i.Credential.Details.Token.OAuth2.PKCE.String(), "auto")},
		)
	}

	return rows
}
