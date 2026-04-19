package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

// setCredentialCmd creates or updates a credential entry in config.
var setCredentialCmd = &cobra.Command{ //nolint:gochecknoglobals // Cobra command definition
	Use:   "set-credential NAME",
	Short: "Set a credential entry in azctx config",
	Long:  "Set a credential entry in azctx config.",
	Example: `  azctx set-credential ci-sp --type service-principal \
    --client-id 11111111-1111-1111-1111-111111111111 \
    --client-secret super-secret`,
	RunE:              runSetCredential,
	DisableAutoGenTag: true,
	Args:              cobra.ExactArgs(1),
}

func init() { //nolint:gochecknoinits // Cobra command setup
	setCredentialCmd.Flags().String("type", "", "Credential type: service-principal|user|managed-identity|oidc")
	setCredentialCmd.Flags().String("client-id", "", "Client ID")
	setCredentialCmd.Flags().String("client-secret", "", "Client secret")
	setCredentialCmd.Flags().String("client-certificate-path", "", "Client certificate path")
	setCredentialCmd.Flags().String("federated-token-file", "", "Path to federated token file")

	if err := setCredentialCmd.MarkFlagRequired("type"); err != nil {
		panic(fmt.Errorf("mark type flag required: %w", err))
	}
}

// runSetCredential executes the set-credential command.
func runSetCredential(cmd *cobra.Command, args []string) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}

	credentialName := args[0]
	if credentialName == "" {
		return fmt.Errorf("credential name must not be empty")
	}

	typeRaw, err := cmd.Flags().GetString("type")
	if err != nil {
		return fmt.Errorf("read type flag: %w", err)
	}

	credentialType, err := config.ParseCredentialType(typeRaw)
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

	certificatePath, err := cmd.Flags().GetString("client-certificate-path")
	if err != nil {
		return fmt.Errorf("read client-certificate-path flag: %w", err)
	}

	federatedTokenFile, err := cmd.Flags().GetString("federated-token-file")
	if err != nil {
		return fmt.Errorf("read federated-token-file flag: %w", err)
	}

	nextCredential := config.Credential{
		Name:                  credentialName,
		Type:                  credentialType,
		ClientID:              clientID,
		ClientSecret:          clientSecret,
		ClientCertificatePath: certificatePath,
		FederatedTokenFile:    federatedTokenFile,
	}

	if err := nextCredential.Validate(); err != nil {
		return err
	}

	wasExisting := false
	if _, found := loaded.Config.CredentialByName(credentialName); found {
		wasExisting = true
	}

	writePath := loaded.PathForCredential(credentialName)
	fileConfig := loaded.FileConfig(writePath)
	fileConfig.UpsertCredential(&nextCredential)

	if err := config.Write(writePath, &fileConfig); err != nil {
		return err
	}

	if wasExisting {
		_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Credential %q modified.\n", credentialName)
		return writeErr
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Credential %q created.\n", credentialName)
	return writeErr
}
