package artifacts

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the artifacts command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifacts",
		Short: "Manage Artifact Registry resources",
	}

	cmd.AddCommand(
		newRepositoriesCommand(cfg, creds),
		newPackagesCommand(cfg, creds),
		newVersionsCommand(cfg, creds),
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

func requireLocation(location string, cfg *config.Config) (string, error) {
	if location != "" {
		return location, nil
	}
	r := cfg.Region()
	if r == "" {
		return "", fmt.Errorf("--location is required (or set region in config)")
	}
	return r, nil
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

// Repositories subcommands

func newRepositoriesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repositories",
		Short: "Manage artifact repositories",
	}

	cmd.AddCommand(
		newRepoListCommand(cfg, creds),
		newRepoDescribeCommand(cfg, creds),
		newRepoCreateCommand(cfg, creds),
		newRepoDeleteCommand(cfg, creds),
	)

	return cmd
}

func newRepoListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(location, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			repos, err := client.ListRepositories(ctx, project, location)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), repos)
			}

			headers := []string{"NAME", "FORMAT", "DESCRIPTION", "SIZE_BYTES"}
			rows := make([][]string, len(repos))
			for i, r := range repos {
				rows[i] = []string{r.Name, r.Format, r.Description, fmt.Sprintf("%d", r.SizeBytes)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location (falls back to config region)")
	return cmd
}

func newRepoDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "describe REPOSITORY",
		Short: "Describe a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(location, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			repo, err := client.GetRepository(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), repo)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", repo.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Format:      %s\n", repo.Format)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", repo.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Size:        %d bytes\n", repo.SizeBytes)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created:     %s\n", repo.CreateTime)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated:     %s\n", repo.UpdateTime)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location (falls back to config region)")
	return cmd
}

func newRepoCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var req CreateRepositoryRequest

	cmd := &cobra.Command{
		Use:   "create REPOSITORY",
		Short: "Create a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(location, cfg)
			if err != nil {
				return err
			}

			req.RepositoryID = args[0]

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			repo, err := client.CreateRepository(ctx, project, location, &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created repository %s.\n", repo.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location (falls back to config region)")
	cmd.Flags().StringVar(&req.Format, "format", "docker", "Repository format (docker, maven, npm, python, go)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Repository description")
	return cmd
}

func newRepoDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "delete REPOSITORY",
		Short: "Delete a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(location, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteRepository(ctx, project, location, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted repository %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location (falls back to config region)")
	return cmd
}

// Packages subcommands

func newPackagesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packages",
		Short: "Manage artifact packages",
	}

	cmd.AddCommand(
		newPkgListCommand(cfg, creds),
		newPkgDeleteCommand(cfg, creds),
	)
	return cmd
}

func newPkgListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var repository string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List packages in a repository",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(location, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			pkgs, err := client.ListPackages(ctx, project, location, repository)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), pkgs)
			}

			headers := []string{"NAME", "CREATE_TIME", "UPDATE_TIME"}
			rows := make([][]string, len(pkgs))
			for i, p := range pkgs {
				rows[i] = []string{p.Name, p.CreateTime, p.UpdateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location (falls back to config region)")
	cmd.Flags().StringVar(&repository, "repository", "", "Repository name")
	_ = cmd.MarkFlagRequired("repository")
	return cmd
}

func newPkgDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var repository string

	cmd := &cobra.Command{
		Use:   "delete PACKAGE",
		Short: "Delete a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(location, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeletePackage(ctx, project, location, repository, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted package %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location (falls back to config region)")
	cmd.Flags().StringVar(&repository, "repository", "", "Repository name")
	_ = cmd.MarkFlagRequired("repository")
	return cmd
}

// Versions subcommands

func newVersionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "Manage artifact versions",
	}

	cmd.AddCommand(
		newVersionListCommand(cfg, creds),
		newVersionDeleteCommand(cfg, creds),
	)
	return cmd
}

func newVersionListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var repository string
	var pkg string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List versions of a package",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(location, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			versions, err := client.ListVersions(ctx, project, location, repository, pkg)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), versions)
			}

			headers := []string{"NAME", "CREATE_TIME", "UPDATE_TIME"}
			rows := make([][]string, len(versions))
			for i, v := range versions {
				rows[i] = []string{v.Name, v.CreateTime, v.UpdateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location (falls back to config region)")
	cmd.Flags().StringVar(&repository, "repository", "", "Repository name")
	cmd.Flags().StringVar(&pkg, "package", "", "Package name")
	_ = cmd.MarkFlagRequired("repository")
	_ = cmd.MarkFlagRequired("package")
	return cmd
}

func newVersionDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var repository string
	var pkg string

	cmd := &cobra.Command{
		Use:   "delete VERSION",
		Short: "Delete a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(location, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteVersion(ctx, project, location, repository, pkg, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted version %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location (falls back to config region)")
	cmd.Flags().StringVar(&repository, "repository", "", "Repository name")
	cmd.Flags().StringVar(&pkg, "package", "", "Package name")
	_ = cmd.MarkFlagRequired("repository")
	_ = cmd.MarkFlagRequired("package")
	return cmd
}
