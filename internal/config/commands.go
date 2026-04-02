package config

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// NewCommand returns the config command group.
func NewCommand(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage gcgo configuration properties",
	}

	cmd.AddCommand(
		newSetCommand(cfg),
		newGetCommand(cfg),
		newListCommand(cfg),
		newUnsetCommand(cfg),
	)

	return cmd
}

func newSetCommand(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "set KEY VALUE",
		Short: "Set a configuration property",
		Long:  "Valid keys: project, account, region, zone",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Set(args[0], args[1]); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated property [%s] to %q.\n", args[0], args[1])
			return nil
		},
	}
}

func newGetCommand(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "get KEY",
		Short: "Get a configuration property value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			val, ok := cfg.Get(args[0])
			if !ok {
				return fmt.Errorf("property %q is not set", args[0])
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), val)
			return nil
		},
	}
}

func newListCommand(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration properties",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			all := cfg.All()
			if len(all) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "(no properties set)")
				return nil
			}

			keys := make([]string, 0, len(all))
			for k := range all {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			for _, k := range keys {
				_, _ = fmt.Fprintf(tw, "%s\t%s\n", k, all[k])
			}
			return tw.Flush()
		},
	}
}

func newUnsetCommand(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "unset KEY",
		Short: "Remove a configuration property",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Unset(args[0]); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Unset property [%s].\n", args[0])
			return nil
		},
	}
}
