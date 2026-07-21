package tabs

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/keyvault"
	"github.com/lvlcn-t/azctx/tui/details"
	"github.com/lvlcn-t/azctx/tui/form"
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
		{Label: labelName, Value: i.Name},
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

// credentialForm builds the create or edit form for a credential. All fields
// are shown; only the ones relevant to the chosen type are used at submit time,
// where the domain validates the result. On edit the name is locked read-only.
func credentialForm(intent formIntent, item details.Item) form.Model {
	var c config.Credential
	title := "New credential"
	readonly := false
	if cred, ok := item.(*CredentialItem); ok && intent == intentEdit {
		c = cred.Credential
		title = "Edit credential"
		readonly = true
	}

	azure := c.Details.Azure
	oauth := config.OAuth2Source{}
	tokenFile := ""
	if c.Details.Token.OAuth2 != nil {
		oauth = *c.Details.Token.OAuth2
	}
	if c.Details.Token.File != nil {
		tokenFile = c.Details.Token.File.Path
	}
	return form.New(title, []form.Field{
		{Key: fieldName, Label: labelName, Placeholder: "my-credential", Value: c.Name, Required: true, ReadOnly: readonly},
		{
			Key: fieldType, Label: "Type", Value: c.Details.Type.String(), Required: true,
			Placeholder: "user | service-principal | managed-identity | workload-identity",
			Validate:    credentialTypeValidator,
		},
		{Key: fieldClientID, Label: "Client ID", Value: azure.ClientID, Placeholder: "for sp/mi/wif"},
		{Key: fieldClientSecret, Label: "Client Secret", Value: azure.ClientSecret, Placeholder: "for service-principal"},
		{Key: fieldCertPath, Label: "Cert Path", Value: azure.ClientCertificatePath, Placeholder: "for service-principal"},
		{Key: fieldTokenSource, Label: "Token Source", Value: c.Details.Token.Source.String(), Placeholder: "file | oauth2 (for wif)"},
		{Key: fieldTokenFile, Label: "Token File", Value: tokenFile, Placeholder: "for wif file source"},
		{Key: fieldIssuer, Label: "OIDC Issuer", Value: oauth.Issuer, Placeholder: "for wif oauth2"},
		{Key: fieldOIDCClientID, Label: "OIDC Client ID", Value: oauth.ClientID, Placeholder: "for wif oauth2"},
		{Key: fieldRedirectURI, Label: "Redirect URI", Value: oauth.RedirectURI, Placeholder: "optional"},
		{Key: fieldScopes, Label: "Scopes", Value: strings.Join(oauth.Scopes, ","), Placeholder: "comma-separated"},
	})
}

// credentialTypeValidator rejects an unsupported credential type.
func credentialTypeValidator(value string) error {
	_, err := config.NewCredentialType(value)
	return err
}

// credentialFromValues assembles a config.Credential from submitted form values.
// Only the fields relevant to the chosen type are populated; the domain
// validates the result.
func credentialFromValues(values map[string]string) *config.Credential {
	credType, _ := config.NewCredentialType(values[fieldType])

	d := config.CredentialDetails{
		Type: credType,
		Azure: config.AzureCredential{
			ClientID:              values[fieldClientID],
			ClientSecret:          values[fieldClientSecret],
			ClientCertificatePath: values[fieldCertPath],
		},
	}

	if credType == config.CredentialTypeWorkloadIdentity {
		d.Token = tokenFromValues(values)
	}

	return &config.Credential{Name: values[fieldName], Details: d}
}

// tokenFromValues builds the token details for a workload-identity credential.
func tokenFromValues(values map[string]string) config.TokenDetails {
	source := config.TokenSource(values[fieldTokenSource])
	token := config.TokenDetails{Source: source}

	switch source {
	case config.TokenSourceFile:
		token.File = &config.FileSource{Path: values[fieldTokenFile]}
	case config.TokenSourceOAuth2:
		token.OAuth2 = &config.OAuth2Source{
			Issuer:      values[fieldIssuer],
			ClientID:    values[fieldOIDCClientID],
			RedirectURI: values[fieldRedirectURI],
			Scopes:      splitScopes(values[fieldScopes]),
		}
	}

	return token
}

// splitScopes parses a comma-separated scope list, trimming blanks.
func splitScopes(raw string) []string {
	if raw == "" {
		return nil
	}

	var scopes []string
	for _, s := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			scopes = append(scopes, trimmed)
		}
	}
	return scopes
}
