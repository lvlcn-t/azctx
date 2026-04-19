package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

// setContextCmd creates or updates a context entry in config.
var setContextCmd = &cobra.Command{ //nolint:gochecknoglobals // Cobra command definition
	Use:               "set-context NAME",
	Short:             "Set a context entry in azctx config",
	Long:              "Set a context entry in azctx config. The context points to tenant and credential entries in the same merged azctx config.",
	Example:           "  azctx set-context prod --tenant corp --credential ci-sp --subscription 00000000-0000-0000-0000-000000000000",
	RunE:              runSetContext,
	DisableAutoGenTag: true,
	Args:              cobra.ExactArgs(1),
}

func init() { //nolint:gochecknoinits // Cobra command setup
	setContextCmd.Flags().String("tenant", "", "Tenant name for the context")
	setContextCmd.Flags().String("credential", "", "Credential name for the context")
	setContextCmd.Flags().String("subscription", "", "Optional subscription ID for the context")
}

// runSetContext executes the set-context command.
func runSetContext(cmd *cobra.Command, args []string) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}

	contextName := args[0]
	if contextName == "" {
		return fmt.Errorf("context name must not be empty")
	}

	tenantName, err := cmd.Flags().GetString("tenant")
	if err != nil {
		return fmt.Errorf("read tenant flag: %w", err)
	}

	credentialName, err := cmd.Flags().GetString("credential")
	if err != nil {
		return fmt.Errorf("read credential flag: %w", err)
	}

	subscriptionID, err := cmd.Flags().GetString("subscription")
	if err != nil {
		return fmt.Errorf("read subscription flag: %w", err)
	}

	nextContext, hasExistingInMerged := nextContextFromFlags(
		&loaded.Config,
		contextName,
		tenantName,
		credentialName,
		subscriptionID,
		cmd.Flags().Changed("subscription"),
	)

	if err := loaded.Config.ValidateContextReferences(nextContext); err != nil {
		return err
	}

	tenant, _ := loaded.Config.TenantByName(nextContext.Tenant)
	if tenant.ID == "" {
		return fmt.Errorf("tenant %q is missing id", tenant.Name)
	}

	credential, _ := loaded.Config.CredentialByName(nextContext.Credential)
	if err := credential.Validate(); err != nil {
		return err
	}

	writePath := loaded.PathForContext(contextName)
	fileConfig := loaded.FileConfig(writePath)

	fileConfig.UpsertContext(nextContext)
	if err := config.Write(writePath, &fileConfig); err != nil {
		return err
	}

	if hasExistingInMerged {
		_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Context %q modified.\n", contextName)
		return writeErr
	}

	_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "Context %q created.\n", contextName)
	return writeErr
}

// nextContextFromFlags resolves the effective context payload for upserts.
func nextContextFromFlags(
	cfg *config.Config,
	contextName string,
	tenantName string,
	credentialName string,
	subscriptionID string,
	subscriptionChanged bool,
) (config.Context, bool) {
	nextContext := config.Context{
		Name:         contextName,
		Tenant:       tenantName,
		Credential:   credentialName,
		Subscription: subscriptionID,
	}

	existingContext, hasExisting := cfg.ContextByName(contextName)
	if !hasExisting {
		return nextContext, false
	}

	nextContext = existingContext
	if tenantName != "" {
		nextContext.Tenant = tenantName
	}

	if credentialName != "" {
		nextContext.Credential = credentialName
	}

	if subscriptionChanged {
		nextContext.Subscription = subscriptionID
	}

	return nextContext, true
}
