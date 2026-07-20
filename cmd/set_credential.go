package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

type setCredentialCmd struct {
	writer config.Writer
	loader config.Loader
	flags  credentialFlagsInput
}

type credentialFlagsInput struct {
	credType     config.CredentialType
	clientID     string
	clientSecret string
	certPath     string
	fedTokenFile string
	issuer       string
	oidcClientID string
	redirectURI  string
	scopes       []string
}

// newSetCredentialCmd creates or updates a credential entry in config.
func newSetCredentialCmd() *cobra.Command {
	command := &setCredentialCmd{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
	}

	cmd := &cobra.Command{
		Use:   "set-credential NAME",
		Short: "Set a credential entry in azctx config",
		Long:  "Set a credential entry in azctx config.",
		Example: `  azctx set-credential ci-sp --type service-principal \
    --client-id 11111111-1111-1111-1111-111111111111 \
    --client-secret super-secret`,
		RunE:              command.run,
		DisableAutoGenTag: true,
		Args:              cobra.ExactArgs(1),
	}

	flags := &command.flags

	// TODO: Consider handling the conversion outside of the flag parsing.
	cmd.Flags().StringVar((*string)(&flags.credType), "type", "", "Credential type: service-principal|user|managed-identity|workload-identity")
	cmd.Flags().StringVar(&flags.clientID, "client-id", "", "Client ID")
	cmd.Flags().StringVar(&flags.clientSecret, "client-secret", "", "Client secret")
	cmd.Flags().StringVar(&flags.certPath, "client-certificate-path", "", "Client certificate path")
	cmd.Flags().StringVar(&flags.fedTokenFile, "federated-token-file", "", "Path to federated token file")
	cmd.Flags().StringVar(&flags.issuer, "issuer", "", "OIDC issuer URL")
	cmd.Flags().StringVar(&flags.oidcClientID, "oidc-client-id", "", "OIDC client ID")
	cmd.Flags().StringVar(&flags.redirectURI, "redirect-uri", "", "OIDC redirect URI")
	cmd.Flags().StringSliceVar(&flags.scopes, "scopes", []string{}, "OIDC scopes")

	if err := cmd.MarkFlagRequired("type"); err != nil {
		panic(fmt.Errorf("mark type flag required: %w", err))
	}

	return cmd
}

// run executes the set-credential command.
func (c *setCredentialCmd) run(cmd *cobra.Command, args []string) error {
	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	credName := args[0]
	if credName == "" {
		return fmt.Errorf("credential name must not be empty")
	}

	var source config.TokenSource
	if c.flags.issuer != "" && c.flags.oidcClientID != "" && c.flags.redirectURI != "" && len(c.flags.scopes) > 0 {
		source = config.TokenSourceOAuth2
	}

	nextCredential := config.Credential{
		Name: credName,
		Details: config.CredentialDetails{
			Type: c.flags.credType,
			Azure: config.AzureCredential{
				ClientID:              c.flags.clientID,
				ClientSecret:          c.flags.clientSecret,
				ClientCertificatePath: c.flags.certPath,
			},
			Token: config.TokenDetails{
				Source: source,
				File: &config.FileSource{
					Path: c.flags.fedTokenFile,
				},
				OAuth2: &config.OAuth2Source{
					Issuer:      c.flags.issuer,
					ClientID:    c.flags.oidcClientID,
					RedirectURI: c.flags.redirectURI,
					Scopes:      c.flags.scopes,
				},
			},
		},
	}

	if err = nextCredential.Validate(); err != nil {
		return err
	}

	wasExisting := false
	if _, found := store.Config.CredentialByName(credName); found {
		wasExisting = true
	}

	path := store.PathForCredential(credName)
	cfg := store.FileConfig(path)
	cfg.UpsertCredential(&nextCredential)
	if err = c.writer.Write(path, &cfg); err != nil {
		return err
	}

	if wasExisting {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Credential %q modified.\n", credName)
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Credential %q created.\n", credName)
	return err
}
