package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/spf13/cobra"
)

type setCtxCmd struct {
	writer config.Writer
	loader config.Loader
}

// newSetCtxCmd creates or updates a context entry in config.
func newSetCtxCmd() *cobra.Command {
	command := &setCtxCmd{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
	}

	cmd := &cobra.Command{
		Use:               "set-context NAME",
		Short:             "Set a context entry in azctx config",
		Long:              "Set a context entry in azctx config. The context points to tenant and credential entries in the same merged azctx config.",
		Example:           "  azctx set-context prod --tenant corp --credential ci-sp --subscription 00000000-0000-0000-0000-000000000000",
		RunE:              command.run,
		DisableAutoGenTag: true,
		Args:              cobra.ExactArgs(1),
	}

	cmd.Flags().String("tenant", "", "Tenant name for the context")
	cmd.Flags().String("credential", "", "Credential name for the context")
	cmd.Flags().String("subscription", "", "Optional subscription ID for the context")

	return cmd
}

// run executes the set-context command.
func (c *setCtxCmd) run(cmd *cobra.Command, args []string) error {
	store, err := c.loader.Load()
	if err != nil {
		return err
	}

	ctx := args[0]
	if ctx == "" {
		return fmt.Errorf("context name must not be empty")
	}

	tenant, err := cmd.Flags().GetString("tenant")
	if err != nil {
		return fmt.Errorf("read tenant flag: %w", err)
	}

	credential, err := cmd.Flags().GetString("credential")
	if err != nil {
		return fmt.Errorf("read credential flag: %w", err)
	}

	subscriptionID, err := cmd.Flags().GetString("subscription")
	if err != nil {
		return fmt.Errorf("read subscription flag: %w", err)
	}

	nextCtx, hasExistingInMerged := nextContextFromFlags(
		&store.Config,
		ctx,
		tenant,
		credential,
		subscriptionID,
		cmd.Flags().Changed("subscription"),
	)

	if err = store.Config.ValidateContextReferences(nextCtx); err != nil {
		return err
	}

	t, _ := store.Config.TenantByName(nextCtx.Details.Tenant)
	if t.Details.ID == "" {
		return fmt.Errorf("tenant %q is missing id", t.Name)
	}

	cred, _ := store.Config.CredentialByName(nextCtx.Details.Credential)
	if err = cred.Validate(); err != nil {
		return err
	}

	path := store.PathForContext(ctx)
	cfg := store.FileConfig(path)
	cfg.UpsertContext(nextCtx)
	if err = c.writer.Write(path, &cfg); err != nil {
		return err
	}

	if hasExistingInMerged {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Context %q modified.\n", ctx)
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Context %q created.\n", ctx)
	return err
}

// nextContextFromFlags resolves the effective context payload for upserts.
func nextContextFromFlags(
	cfg *config.Config,
	ctx string,
	tenant string,
	credential string,
	subscriptionID string,
	subscriptionChanged bool,
) (config.Context, bool) {
	nextCtx := config.Context{
		Name: ctx,
		Details: config.ContextDetails{
			Tenant:       tenant,
			Credential:   credential,
			Subscription: subscriptionID,
		},
	}

	existing, ok := cfg.ContextByName(ctx)
	if !ok {
		return nextCtx, false
	}

	nextCtx = existing
	if tenant != "" {
		nextCtx.Details.Tenant = tenant
	}

	if credential != "" {
		nextCtx.Details.Credential = credential
	}

	if subscriptionChanged {
		nextCtx.Details.Subscription = subscriptionID
	}

	return nextCtx, true
}
