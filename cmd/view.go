package cmd

import (
	"fmt"

	"github.com/lvlcn-t/azctx/config"
	"github.com/lvlcn-t/azctx/output"
	"github.com/spf13/cobra"
)

type viewCommand struct {
	writer config.Writer
	loader config.Loader
}

// newViewCmd renders merged azctx config.
func newViewCmd() *cobra.Command {
	command := &viewCommand{
		loader: config.NewLoader(),
		writer: config.NewWriter(),
	}

	cmd := &cobra.Command{
		Use:               "view",
		Short:             "Display merged azctx config",
		Long:              "Display merged azctx config from AZCTX path list or the default config path.",
		Example:           "  azctx view\n  azctx view -o json\n  azctx view --raw -o json",
		RunE:              command.run,
		DisableAutoGenTag: true,
	}

	bindOutputFlag(cmd)
	cmd.Flags().Bool("raw", false, "Print the source file instead of merged output")

	return cmd
}

// run executes the view command.
func (command *viewCommand) run(cmd *cobra.Command, _ []string) error {
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return fmt.Errorf("read raw flag: %w", err)
	}

	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}

	store, err := command.loader.Load()
	if err != nil {
		return err
	}

	cfg := store.Config
	if raw {
		if store.WritePath == "" {
			return fmt.Errorf("cannot resolve config write path")
		}

		cfg, err = command.loader.Read(store.WritePath)
		if err != nil {
			return err
		}
	}

	switch format {
	case output.FormatJSON:
		return output.PrintJSON(cmd.OutOrStdout(), cfg)
	case output.FormatText, output.FormatTable:
		return output.PrintTable(
			cmd.OutOrStdout(),
			[]string{"SECTION", "NAME", "VALUE"}, //nolint:goconst // column headers
			viewRows(&cfg),
		)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

// viewRows flattens config sections into a table-friendly shape.
func viewRows(cfg *config.Config) [][]string {
	if cfg == nil {
		cfg = &config.Config{}
	}

	rows := make([][]string, 0, len(cfg.Tenants)+len(cfg.Credentials)+len(cfg.Contexts)+1)

	rows = append(rows, []string{"meta", "current-context", emptyIfUnset(cfg.CurrentContext)})

	for _, tenant := range cfg.Tenants {
		rows = append(rows, []string{"tenant", tenant.Name, tenant.Details.ID})
	}

	for _, credential := range cfg.Credentials {
		rows = append(rows, []string{"credential", credential.Name, string(credential.Details.Type)})
	}

	for _, context := range cfg.Contexts {
		rows = append(rows, []string{
			"context",
			context.Name,
			fmt.Sprintf(
				"tenant=%s credential=%s subscription=%s",
				context.Details.Tenant,
				context.Details.Credential,
				emptyIfUnset(context.Details.Subscription),
			),
		})
	}

	return rows
}
