package functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/flags"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the functions command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "functions",
		Short: "Manage Cloud Functions",
	}

	cmd.AddCommand(
		newListCommand(cfg, creds),
		newDescribeCommand(cfg, creds),
		newDeployCommand(cfg, creds),
		newDeleteCommand(cfg, creds),
		newCallCommand(cfg, creds),
	)
	return cmd
}

func requireProject(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func requireRegion(region string, cfg *config.Config) (string, error) {
	if region != "" {
		return region, nil
	}
	r := cfg.Region()
	if r == "" {
		return "", fmt.Errorf("--region is required (or set region in config)")
	}
	return r, nil
}

func addRegionCompletion(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("region", func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var matches []string
		for _, r := range flags.CommonRegions {
			if strings.HasPrefix(r, toComplete) {
				matches = append(matches, r)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	})
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, creds, opt)
}

func newListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List functions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err = requireRegion(region, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			funcs, err := client.List(ctx, project, region)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), funcs)
			}

			headers := []string{"NAME", "STATE", "RUNTIME", "REGION"}
			rows := make([][]string, len(funcs))
			for i, f := range funcs {
				rows[i] = []string{f.Name, f.State, f.Runtime, f.Region}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	return cmd
}

func newDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "describe FUNCTION",
		Short: "Describe a function",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err = requireRegion(region, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			fn, err := client.Get(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), fn)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", fn.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", fn.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Runtime:     %s\n", fn.Runtime)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Entry Point: %s\n", fn.EntryPoint)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Region:      %s\n", fn.Region)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "URL:         %s\n", fn.URL)
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	return cmd
}

func newDeployCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req DeployRequest
	var region string

	cmd := &cobra.Command{
		Use:   "deploy FUNCTION",
		Short: "Deploy a function",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err = requireRegion(region, cfg)
			if err != nil {
				return err
			}

			req.Name = args[0]
			req.Region = region

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.Deploy(ctx, project, region, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deployed function %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringVar(&req.Runtime, "runtime", "", "Runtime (e.g. go121, python312, nodejs20)")
	cmd.Flags().StringVar(&req.EntryPoint, "entry-point", "", "Entry point function name")
	cmd.Flags().StringVar(&req.Source, "source", "", "Source location (GCS bucket or local path)")
	cmd.Flags().BoolVar(&req.TriggerHTTP, "trigger-http", false, "Use HTTP trigger")
	cmd.Flags().StringVar(&req.TriggerTopic, "trigger-topic", "", "Pub/Sub topic trigger")
	cmd.Flags().StringVar(&req.Memory, "memory", "", "Memory limit (e.g. 256Mi)")
	cmd.Flags().StringVar(&req.Timeout, "timeout", "", "Function timeout (e.g. 60s)")
	_ = cmd.MarkFlagRequired("runtime")
	return cmd
}

func newDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "delete FUNCTION",
		Short: "Delete a function",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err = requireRegion(region, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.Delete(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted function %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	return cmd
}

func newCallCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var data string

	cmd := &cobra.Command{
		Use:   "call FUNCTION",
		Short: "Call a function",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err = requireRegion(region, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			result, err := client.Call(ctx, project, region, args[0], &CallRequest{Data: data})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), result)
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringVar(&data, "data", "", "Data to send to the function")
	return cmd
}
