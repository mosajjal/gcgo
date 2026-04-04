package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/compute"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/flags"
	"github.com/mosajjal/gcgo/internal/logging"
	"github.com/spf13/cobra"
)

// newWhoamiCommand returns "gcgo whoami" — prints the active account,
// project, region and zone in one shot.
func newWhoamiCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show active identity, project, region, and zone [shortcut]",
		Long:  "Prints the authenticated account, active project, default region, and default zone.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			account, err := creds.ActiveAccount()
			if err != nil {
				account = "(not authenticated — run 'gcgo auth login')"
			}

			project := cfg.Project("")
			if project == "" {
				project = "(not set — run 'gcgo config set project PROJECT_ID')"
			}

			region := cfg.Region()
			if region == "" {
				region = "(not set)"
			}

			zone := cfg.Zone()
			if zone == "" {
				zone = "(not set)"
			}

			format, _ := cmd.Root().PersistentFlags().GetString("format")
			if format == "json" {
				out := map[string]string{
					"account": account,
					"project": project,
					"region":  region,
					"zone":    zone,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			rows := []struct{ key, val string }{
				{"account", account},
				{"project", project},
				{"region", region},
				{"zone", zone},
			}
			for _, r := range rows {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-10s  %s\n", r.key, r.val)
			}
			return nil
		},
	}
}

// newUseCommand returns "gcgo use PROJECT [--region REGION] [--zone ZONE]"
// as a shorthand for setting config values in one step.
func newUseCommand(cfg *config.Config) *cobra.Command {
	var region, zone string

	cmd := &cobra.Command{
		Use:   "use PROJECT",
		Short: "Set project/region/zone in one step [shortcut for gcgo config set]",
		Long:  "Shorthand for 'gcgo config set project PROJECT'. Optionally set region and zone in the same command.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Set("project", args[0]); err != nil {
				return err
			}
			if region != "" {
				if err := cfg.Set("region", region); err != nil {
					return err
				}
			}
			if zone != "" {
				if err := cfg.Set("zone", zone); err != nil {
					return err
				}
			}
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "project  %s\n", args[0])
			if region != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "region   %s\n", region)
			}
			if zone != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "zone     %s\n", zone)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Also set the default region")
	_ = cmd.RegisterFlagCompletionFunc("region", func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var matches []string
		for _, r := range flags.CommonRegions {
			if strings.HasPrefix(r, toComplete) {
				matches = append(matches, r)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringVar(&zone, "zone", "", "Also set the default zone")
	return cmd
}

// consoleServices maps short names to console.cloud.google.com URL paths.
var consoleServices = map[string]string{
	"compute":    "compute/instances",
	"vm":         "compute/instances",
	"gke":        "kubernetes/list/overview",
	"container":  "kubernetes/list/overview",
	"run":        "run",
	"functions":  "functions/list",
	"sql":        "sql/instances",
	"spanner":    "spanner/instances",
	"firestore":  "firestore/databases",
	"bigtable":   "bigtable/instances",
	"storage":    "storage/browser",
	"gcs":        "storage/browser",
	"logging":    "logs/query",
	"logs":       "logs/query",
	"iam":        "iam-admin/iam",
	"secrets":    "security/secret-manager",
	"kms":        "security/kms/keyrings",
	"pubsub":     "cloudpubsub/topicList",
	"monitoring": "monitoring",
	"builds":     "cloud-build/builds",
	"artifacts":  "artifacts",
	"redis":      "memorystore/redis/instances",
	"scheduler":  "cloudscheduler",
	"tasks":      "cloudtasks",
	"billing":    "billing",
	"apis":       "apis/library",
	"dashboard":  "home/dashboard",
}

// newOpenCommand returns "gcgo open [SERVICE]" — opens the GCP console for
// the current project. Without a service name it opens the project dashboard.
func newOpenCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [SERVICE]",
		Short: "Open GCP console in browser [shortcut for console.cloud.google.com]",
		Long: fmt.Sprintf("Opens console.cloud.google.com for the current project.\nAvailable services: %s",
			strings.Join(serviceNames(), ", ")),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project := cfg.Project("")

			path := "home/dashboard"
			if len(args) == 1 {
				svc := strings.ToLower(args[0])
				p, ok := consoleServices[svc]
				if !ok {
					return fmt.Errorf("unknown service %q — available: %s", svc, strings.Join(serviceNames(), ", "))
				}
				path = p
			}

			url := "https://console.cloud.google.com/" + path
			if project != "" {
				url += "?project=" + project
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", url)
			return openURL(url)
		},
	}

	_ = cmd.RegisterFlagCompletionFunc("", func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return serviceNames(), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// newSSHCommand returns "gcgo ssh VM" — top-level shortcut for compute ssh.
func newSSHCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var user string

	cmd := &cobra.Command{
		Use:   "ssh VM [-- EXTRA_ARGS...]",
		Short: "SSH into a VM [shortcut for gcgo compute ssh]",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireAlias(cmd, cfg)
			if err != nil {
				return err
			}
			zone, _ := cmd.Flags().GetString("zone")
			if zone == "" {
				zone = cfg.Zone()
			}
			if zone == "" {
				return fmt.Errorf("no zone set — use --zone or 'gcgo config set zone ZONE'")
			}

			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}
			client, err := compute.NewClient(ctx, opt)
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}

			ip, err := compute.ResolveInstanceIP(ctx, client, project, zone, args[0])
			if err != nil {
				return err
			}
			return compute.ExecSSH(compute.SSHArgs(user, ip, args[1:]))
		},
	}

	cmd.Flags().StringVar(&user, "user", "", "SSH username")
	compute.AddZoneFlag(cmd)
	cmd.Flags().String("project", "", "GCP project ID")
	return cmd
}

// newLogsCommand returns "gcgo logs [FILTER]" — top-level shortcut for
// gcgo logging read / tail.
func newLogsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var limit int
	var tail bool

	cmd := &cobra.Command{
		Use:   "logs [FILTER]",
		Short: "Read or stream logs [shortcut for gcgo logging read/tail]",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireAlias(cmd, cfg)
			if err != nil {
				return err
			}

			filter := ""
			if len(args) > 0 {
				filter = args[0]
			}

			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			if tail {
				return logging.TailLogs(ctx, cmd.OutOrStdout(), project, filter, opt)
			}

			client, err := logging.NewClient(ctx, opt)
			if err != nil {
				return fmt.Errorf("create logging client: %w", err)
			}
			entries, err := client.ReadLogs(ctx, project, filter, limit)
			if err != nil {
				return err
			}
			for _, e := range entries {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s  %-8s  %s\n", e.Timestamp, e.Severity, e.Payload)
			}
			return nil
		},
	}

	cmd.Flags().String("project", "", "GCP project ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of entries")
	cmd.Flags().BoolVar(&tail, "tail", false, "Stream log entries in real-time (like tail -f)")
	return cmd
}

// newTokenCommand returns "gcgo token" — prints the current access token.
// Useful for scripting: curl -H "Authorization: Bearer $(gcgo token)" ...
func newTokenCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print current access token [shortcut for gcgo auth print-access-token]",
		Long:  "Prints the OAuth2 access token for the active credentials. Useful for: curl -H \"Authorization: Bearer $(gcgo token)\" ...",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()
			token, err := creds.AccessToken(ctx, "", nil)
			if err != nil {
				return fmt.Errorf("get token: %w", err)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), token)
			return nil
		},
	}
}

// requireAlias reads --project flag then falls back to config. Used by alias
// commands that add their own --project flag (not the persistent root one).
func requireAlias(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func serviceNames() []string {
	names := make([]string, 0, len(consoleServices))
	seen := make(map[string]bool)
	for k, v := range consoleServices {
		if !seen[v] {
			names = append(names, k)
			seen[v] = true
		}
	}
	return names
}

func openURL(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "linux":
		cmd, args = "xdg-open", []string{url}
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform %s — open %s manually", runtime.GOOS, url)
	}
	return exec.Command(filepath.Clean(cmd), args...).Start() //nolint:gosec // cmd is a fixed string per platform
}
