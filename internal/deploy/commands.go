package deploy

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the deploy command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Manage Cloud Deploy resources",
	}

	cmd.AddCommand(
		newDeliveryPipelinesCommand(cfg, creds),
		newReleasesCommand(cfg, creds),
	)
	return cmd
}

func deployClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func requireProject(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func requireLocation(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("location")
	if flagVal != "" {
		return flagVal, nil
	}
	loc := cfg.Region()
	if loc == "" {
		return "", fmt.Errorf("--location is required (or set region in config)")
	}
	return loc, nil
}

func newDeliveryPipelinesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delivery-pipelines",
		Short: "Manage Cloud Deploy delivery pipelines",
	}

	cmd.AddCommand(
		newDeliveryPipelinesListCommand(cfg, creds),
		newDeliveryPipelinesDescribeCommand(cfg, creds),
		newDeliveryPipelinesCreateCommand(cfg, creds),
		newDeliveryPipelinesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newDeliveryPipelinesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List delivery pipelines",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := deployClient(ctx, creds)
			if err != nil {
				return err
			}
			pipelines, err := client.ListDeliveryPipelines(ctx, project, location)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), pipelines)
			}
			headers := []string{"NAME", "DESCRIPTION", "SUSPENDED", "UPDATED"}
			rows := make([][]string, len(pipelines))
			for i, p := range pipelines {
				rows[i] = []string{p.Name, p.Description, fmt.Sprintf("%v", p.Suspended), p.UpdateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Deploy location")
	return cmd
}

func newDeliveryPipelinesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "describe PIPELINE",
		Short: "Describe a delivery pipeline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := deployClient(ctx, creds)
			if err != nil {
				return err
			}
			pipeline, err := client.GetDeliveryPipeline(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), pipeline)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", pipeline.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", pipeline.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Suspended:   %v\n", pipeline.Suspended)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated:     %s\n", pipeline.UpdateTime)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Deploy location")
	return cmd
}

func newDeliveryPipelinesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var description string
	var suspended bool
	var targets []string

	cmd := &cobra.Command{
		Use:   "create PIPELINE",
		Short: "Create a delivery pipeline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := deployClient(ctx, creds)
			if err != nil {
				return err
			}
			pipeline, err := client.CreateDeliveryPipeline(ctx, project, location, &CreateDeliveryPipelineRequest{
				Name:        args[0],
				Description: description,
				Suspended:   suspended,
				Targets:     targets,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created delivery pipeline %s.\n", pipeline.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Deploy location")
	cmd.Flags().StringVar(&description, "description", "", "Pipeline description")
	cmd.Flags().BoolVar(&suspended, "suspended", false, "Create the pipeline in suspended state")
	cmd.Flags().StringArrayVar(&targets, "target", nil, "Promotion target IDs in order")
	return cmd
}

func newDeliveryPipelinesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "delete PIPELINE",
		Short: "Delete a delivery pipeline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := deployClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteDeliveryPipeline(ctx, project, location, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted delivery pipeline %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Deploy location")
	return cmd
}

func newReleasesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases",
		Short: "Manage Cloud Deploy releases",
	}

	cmd.AddCommand(
		newReleasesListCommand(cfg, creds),
		newReleasesDescribeCommand(cfg, creds),
		newReleasesCreateCommand(cfg, creds),
	)
	return cmd
}

func newReleasesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "list PIPELINE",
		Short: "List releases in a delivery pipeline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := deployClient(ctx, creds)
			if err != nil {
				return err
			}
			releases, err := client.ListReleases(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), releases)
			}
			headers := []string{"NAME", "DESCRIPTION", "RENDER_STATE", "UPDATED"}
			rows := make([][]string, len(releases))
			for i, r := range releases {
				rows[i] = []string{r.Name, r.Description, r.RenderState, r.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Deploy location")
	return cmd
}

func newReleasesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "describe PIPELINE RELEASE",
		Short: "Describe a release",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := deployClient(ctx, creds)
			if err != nil {
				return err
			}
			release, err := client.GetRelease(ctx, project, location, args[0], args[1])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), release)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", release.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description:  %s\n", release.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Render State: %s\n", release.RenderState)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created:      %s\n", release.CreateTime)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Deploy location")
	return cmd
}

func newReleasesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var description string
	var skaffoldConfigURI string
	var skaffoldConfigPath string
	var skaffoldVersion string

	cmd := &cobra.Command{
		Use:   "create PIPELINE RELEASE",
		Short: "Create a release",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}
			if skaffoldConfigURI == "" {
				return fmt.Errorf("--skaffold-config-uri is required")
			}

			ctx := context.Background()
			client, err := deployClient(ctx, creds)
			if err != nil {
				return err
			}
			release, err := client.CreateRelease(ctx, project, location, args[0], &CreateReleaseRequest{
				Name:               args[1],
				Description:        description,
				SkaffoldConfigUri:  skaffoldConfigURI,
				SkaffoldConfigPath: skaffoldConfigPath,
				SkaffoldVersion:    skaffoldVersion,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created release %s.\n", release.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Deploy location")
	cmd.Flags().StringVar(&description, "description", "", "Release description")
	cmd.Flags().StringVar(&skaffoldConfigURI, "skaffold-config-uri", "", "Cloud Storage URI for skaffold config tarball")
	cmd.Flags().StringVar(&skaffoldConfigPath, "skaffold-config-path", "", "Skaffold config path inside the archive")
	cmd.Flags().StringVar(&skaffoldVersion, "skaffold-version", "", "Skaffold version")
	return cmd
}
