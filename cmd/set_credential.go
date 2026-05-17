package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

type setCredentialCmd struct {
	loader config.Loader
	writer config.Writer
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

	cmd.Flags().String("type", "", "Credential type: service-principal|user|managed-identity|oidc")
	cmd.Flags().String("client-id", "", "Client ID")
	cmd.Flags().String("client-secret", "", "Client secret")
	cmd.Flags().String("client-certificate-path", "", "Client certificate path")
	cmd.Flags().String("federated-token-file", "", "Path to federated token file")

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

	typeRaw, err := cmd.Flags().GetString("type")
	if err != nil {
		return fmt.Errorf("read type flag: %w", err)
	}

	credType, err := config.NewCredentialType(typeRaw)
	if err != nil {
		return err
	}

	clientID, err := cmd.Flags().GetString("client-id")
	if err != nil {
		return fmt.Errorf("read client-id flag: %w", err)
	}

	clientSecret, err := cmd.Flags().GetString("client-secret")
	if err != nil {
		return fmt.Errorf("read client-secret flag: %w", err)
	}

	certPath, err := cmd.Flags().GetString("client-certificate-path")
	if err != nil {
		return fmt.Errorf("read client-certificate-path flag: %w", err)
	}

	fedTokenFile, err := cmd.Flags().GetString("federated-token-file")
	if err != nil {
		return fmt.Errorf("read federated-token-file flag: %w", err)
	}

	nextCredential := config.Credential{
		Name: credName,
		Credential: config.CredentialDetails{
			Type: credType,
			Azure: config.AzureCredential{
				ClientID:              clientID,
				ClientSecret:          clientSecret,
				ClientCertificatePath: certPath,
			},
			Token: config.Token{
				Source: config.TokenSourceFile,
				File: &config.FileSource{
					Path: fedTokenFile,
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
