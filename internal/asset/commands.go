package asset

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the asset command group.
func NewCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "Cloud Asset Inventory",
	}

	cmd.AddCommand(
		newSearchAllResourcesCommand(creds),
		newSearchAllIAMPoliciesCommand(creds),
		newAnalyzeIamPolicyCommand(creds),
		newExportCommand(creds),
		newFeedsCommand(creds),
	)

	return cmd
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newSearchAllResourcesCommand(creds *auth.Credentials) *cobra.Command {
	var scope, query string
	var assetTypes []string

	cmd := &cobra.Command{
		Use:   "search-all-resources",
		Short: "Search all resources in a scope",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scope == "" {
				return fmt.Errorf("--scope is required (e.g. projects/my-project or organizations/123)")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			resources, err := client.SearchAllResources(ctx, scope, query, assetTypes)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), resources)
			}

			headers := []string{"NAME", "ASSET_TYPE", "PROJECT", "LOCATION"}
			rows := make([][]string, len(resources))
			for i, r := range resources {
				rows[i] = []string{r.Name, r.AssetType, r.Project, r.Location}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope to search (e.g. projects/my-project)")
	cmd.Flags().StringVar(&query, "query", "", "Query filter")
	cmd.Flags().StringSliceVar(&assetTypes, "asset-types", nil, "Asset types to filter")
	_ = cmd.MarkFlagRequired("scope")

	return cmd
}

func newSearchAllIAMPoliciesCommand(creds *auth.Credentials) *cobra.Command {
	var scope, query string

	cmd := &cobra.Command{
		Use:   "search-all-iam-policies",
		Short: "Search all IAM policies in a scope",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scope == "" {
				return fmt.Errorf("--scope is required (e.g. projects/my-project or organizations/123)")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			results, err := client.SearchAllIAMPolicies(ctx, scope, query)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), results)
			}

			headers := []string{"RESOURCE", "PROJECT"}
			rows := make([][]string, len(results))
			for i, r := range results {
				rows[i] = []string{r.Resource, r.Project}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope to search (e.g. projects/my-project)")
	cmd.Flags().StringVar(&query, "query", "", "Query filter")
	_ = cmd.MarkFlagRequired("scope")

	return cmd
}

func newExportCommand(creds *auth.Credentials) *cobra.Command {
	var scope, outputPath, contentType string
	var assetTypes []string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export asset inventory to GCS",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scope == "" {
				return fmt.Errorf("--scope is required (e.g. projects/my-project or organizations/123)")
			}
			if outputPath == "" {
				return fmt.Errorf("--output-path is required (gs://bucket/path)")
			}
			if !strings.HasPrefix(outputPath, "gs://") {
				return fmt.Errorf("--output-path must be a GCS URI (gs://bucket/path)")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			result, err := client.Export(ctx, scope, outputPath, assetTypes, contentType)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Exported assets to %s.\n", result.OutputURI)
			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Parent scope (e.g. projects/my-project)")
	cmd.Flags().StringVar(&outputPath, "output-path", "", "GCS output path (gs://bucket/path)")
	cmd.Flags().StringSliceVar(&assetTypes, "asset-types", nil, "Asset types to export")
	cmd.Flags().StringVar(&contentType, "content-type", "", "Content type: resource, iam-policy, org-policy, access-policy")
	_ = cmd.MarkFlagRequired("scope")
	_ = cmd.MarkFlagRequired("output-path")

	return cmd
}

func newFeedsCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feeds",
		Short: "Manage asset feeds",
	}
	cmd.AddCommand(
		newFeedsListCommand(creds),
		newFeedsDescribeCommand(creds),
		newFeedsCreateCommand(creds),
		newFeedsDeleteCommand(creds),
	)
	return cmd
}

func newFeedsListCommand(creds *auth.Credentials) *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List asset feeds in a scope",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scope == "" {
				return fmt.Errorf("--scope is required (e.g. projects/my-project or organizations/123)")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			feeds, err := client.ListFeeds(ctx, scope)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), feeds)
			}

			headers := []string{"NAME", "CONTENT_TYPE", "TOPIC"}
			rows := make([][]string, len(feeds))
			for i, feed := range feeds {
				rows[i] = []string{feed.Name, feed.ContentType, feed.Topic}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "", "Parent scope (e.g. projects/my-project)")
	_ = cmd.MarkFlagRequired("scope")
	return cmd
}

func newFeedsDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe FEED_NAME",
		Short: "Describe an asset feed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			feed, err := client.GetFeed(ctx, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), feed)
		},
	}
}

func newFeedsCreateCommand(creds *auth.Credentials) *cobra.Command {
	var scope, topic, contentType string
	var assetTypes []string

	cmd := &cobra.Command{
		Use:   "create FEED_ID",
		Short: "Create an asset feed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if scope == "" {
				return fmt.Errorf("--scope is required (e.g. projects/my-project or organizations/123)")
			}
			if topic == "" {
				return fmt.Errorf("--topic is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			feed, err := client.CreateFeed(ctx, scope, args[0], topic, assetTypes, contentType)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), feed)
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "", "Parent scope (e.g. projects/my-project)")
	cmd.Flags().StringVar(&topic, "topic", "", "Pub/Sub topic resource name")
	cmd.Flags().StringSliceVar(&assetTypes, "asset-types", nil, "Asset types to include")
	cmd.Flags().StringVar(&contentType, "content-type", "resource", "Feed content type")
	return cmd
}

func newFeedsDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete FEED_NAME",
		Short: "Delete an asset feed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteFeed(ctx, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted asset feed %s.\n", args[0])
			return nil
		},
	}
}

func newAnalyzeIamPolicyCommand(creds *auth.Credentials) *cobra.Command {
	var scope, identity, permission, resourceName, savedAnalysisQuery, accessTime, executionTimeout string
	var roles []string
	var expandGroups, expandResources, expandRoles, outputGroupEdges, outputResourceEdges, analyzeServiceAccountImpersonation bool

	cmd := &cobra.Command{
		Use:   "analyze-iam-policy",
		Short: "Analyze IAM policy for a scope",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scope == "" {
				return fmt.Errorf("--scope is required (e.g. projects/my-project or organizations/123)")
			}
			if identity == "" && permission == "" && resourceName == "" {
				return fmt.Errorf("at least one of --identity, --permission, or --resource-name is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			resp, err := client.AnalyzeIamPolicy(ctx, scope, &AnalyzeIamPolicyRequest{
				Identity:                       identity,
				Permission:                     permission,
				ResourceName:                   resourceName,
				Roles:                          roles,
				ExpandGroups:                   expandGroups,
				ExpandResources:                expandResources,
				ExpandRoles:                    expandRoles,
				OutputGroupEdges:               outputGroupEdges,
				OutputResourceEdges:            outputResourceEdges,
				AnalyzeServiceAccountImpersonation: analyzeServiceAccountImpersonation,
				SavedAnalysisQuery:             savedAnalysisQuery,
				AccessTime:                     accessTime,
				ExecutionTimeout:               executionTimeout,
			})
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), resp)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope to analyze")
	cmd.Flags().StringVar(&identity, "identity", "", "Principal identity")
	cmd.Flags().StringVar(&permission, "permission", "", "Permission to analyze")
	cmd.Flags().StringVar(&resourceName, "resource-name", "", "Full resource name to analyze")
	cmd.Flags().StringSliceVar(&roles, "role", nil, "Role to include in access analysis")
	cmd.Flags().BoolVar(&expandGroups, "expand-groups", false, "Expand Google groups")
	cmd.Flags().BoolVar(&expandResources, "expand-resources", false, "Expand lower resources")
	cmd.Flags().BoolVar(&expandRoles, "expand-roles", false, "Expand roles to permissions")
	cmd.Flags().BoolVar(&outputGroupEdges, "output-group-edges", false, "Output group edges")
	cmd.Flags().BoolVar(&outputResourceEdges, "output-resource-edges", false, "Output resource edges")
	cmd.Flags().BoolVar(&analyzeServiceAccountImpersonation, "analyze-service-account-impersonation", false, "Analyze service account impersonation paths")
	cmd.Flags().StringVar(&savedAnalysisQuery, "saved-analysis-query", "", "Saved analysis query name")
	cmd.Flags().StringVar(&accessTime, "access-time", "", "RFC3339 access time")
	cmd.Flags().StringVar(&executionTimeout, "execution-timeout", "", "Execution timeout")
	_ = cmd.MarkFlagRequired("scope")

	return cmd
}
