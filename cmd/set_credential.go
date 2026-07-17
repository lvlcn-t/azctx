package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

type setCredentialCmd struct {
	writer config.Writer
	loader config.Loader
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

	cmd.Flags().String("type", "", "Credential type: service-principal|user|managed-identity|workload-identity")
	cmd.Flags().String("client-id", "", "Client ID")
	cmd.Flags().String("client-secret", "", "Client secret")
	cmd.Flags().String("client-certificate-path", "", "Client certificate path")
	cmd.Flags().String("federated-token-file", "", "Path to federated token file")
	cmd.Flags().String("issuer", "", "OIDC issuer URL")
	cmd.Flags().String("oidc-client-id", "", "OIDC client ID")
	cmd.Flags().String("redirect-uri", "", "OIDC redirect URI")
	cmd.Flags().StringSlice("scopes", []string{}, "OIDC scopes")

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

	input, err := readCredentialFlags(cmd)
	if err != nil {
		return err
	}

	source, err := input.resolveTokenSource()
	if err != nil {
		return err
	}

	nextCredential := config.Credential{
		Name: credName,
		Details: config.CredentialDetails{
			Type: input.credType,
			Azure: config.AzureCredential{
				ClientID:              input.clientID,
				ClientSecret:          input.clientSecret,
				ClientCertificatePath: input.certPath,
			},
			Token: config.TokenDetails{
				Source: source,
				File: &config.FileSource{
					Path: input.fedTokenFile,
				},
				OAuth2: &config.OAuth2Source{
					Issuer:      input.issuer,
					ClientID:    input.oidcClientID,
					RedirectURI: input.redirectURI,
					Scopes:      input.scopes,
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

func readCredentialFlags(cmd *cobra.Command) (credentialFlagsInput, error) {
	var input credentialFlagsInput

	typeRaw, err := cmd.Flags().GetString("type")
	if err != nil {
		return input, fmt.Errorf("read type flag %w", err)
	}

	input.credType, err = config.NewCredentialType(typeRaw)
	if err != nil {
		return input, err
	}

	input.clientID, err = cmd.Flags().GetString("client-id")
	if err != nil {
		return input, fmt.Errorf("read client-id flag %w", err)
	}

	input.clientSecret, err = cmd.Flags().GetString("client-secret")
	if err != nil {
		return input, fmt.Errorf("read client-secret flag %w", err)
	}

	input.certPath, err = cmd.Flags().GetString("client-certificate-path")
	if err != nil {
		return input, fmt.Errorf("read client-certificate-path flag %w", err)
	}

	input.fedTokenFile, err = cmd.Flags().GetString("federated-token-file")
	if err != nil {
		return input, fmt.Errorf("read federated-token-file flag %w", err)
	}

	input.issuer, err = cmd.Flags().GetString("issuer")
	if err != nil {
		return input, fmt.Errorf("read issuer flag %w", err)
	}

	input.oidcClientID, err = cmd.Flags().GetString("oidc-client-id")
	if err != nil {
		return input, fmt.Errorf("read client-id flag %w", err)
	}

	input.redirectURI, err = cmd.Flags().GetString("redirect-uri")
	if err != nil {
		return input, fmt.Errorf("read redirect-uri flag %w", err)
	}

	input.scopes, err = cmd.Flags().GetStringSlice("scopes")
	if err != nil {
		return input, fmt.Errorf("read scopes flag: %w", err)
	}

	return input, nil
}

func (f *credentialFlagsInput) resolveTokenSource() (config.TokenSource, error) {
	if f.credType != config.CredentialTypeWorkloadIdentity {
		return config.TokenSourceFile, nil
	}

	hasFile := f.fedTokenFile != ""
	hasOIDC := f.hasOIDCParams()

	switch {
	case hasFile && hasOIDC:
		return "", fmt.Errorf("cannot specify both federated-token-file and OIDC parameters for workload identity credential")
	case hasOIDC:
		return config.TokenSourceOAuth2, nil
	case hasFile:
		return config.TokenSourceFile, nil
	default:
		return "", fmt.Errorf("must specify either federated-token-file or all OIDC parameters for workload identity credential")
	}
}

func (f *credentialFlagsInput) hasOIDCParams() bool {
	return f.issuer != "" || f.oidcClientID != "" ||
		f.redirectURI != "" || len(f.scopes) > 0
}
