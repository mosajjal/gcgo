package run

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	runv1 "google.golang.org/api/run/v1"
	"google.golang.org/api/option"

	"github.com/spf13/cobra"
)

// DomainMapping holds Cloud Run domain mapping fields.
type DomainMapping struct {
	Name     string `json:"name"`
	Region   string `json:"region"`
	Status   string `json:"status,omitempty"`
	MappedTo string `json:"mapped_to,omitempty"`
}

// CreateDomainMappingRequest holds parameters for creating a domain mapping.
type CreateDomainMappingRequest struct {
	DomainName  string
	ServiceName string
}

// DomainMappingsClient defines Cloud Run domain mappings operations.
type DomainMappingsClient interface {
	ListDomainMappings(ctx context.Context, project, region string) ([]*DomainMapping, error)
	GetDomainMapping(ctx context.Context, project, region, name string) (*DomainMapping, error)
	CreateDomainMapping(ctx context.Context, project, region string, req *CreateDomainMappingRequest) (*DomainMapping, error)
	DeleteDomainMapping(ctx context.Context, project, region, name string) error
}

type gcpDomainMappingsClient struct {
	svc *runv1.APIService
}

// NewDomainMappingsClient creates a DomainMappingsClient backed by the Cloud Run v1 API.
func NewDomainMappingsClient(ctx context.Context, region string, opts ...option.ClientOption) (DomainMappingsClient, error) {
	endpoint := fmt.Sprintf("https://%s-run.googleapis.com/", region)
	allOpts := append([]option.ClientOption{option.WithEndpoint(endpoint)}, opts...)
	svc, err := runv1.NewService(ctx, allOpts...)
	if err != nil {
		return nil, fmt.Errorf("create domain mappings client: %w", err)
	}
	return &gcpDomainMappingsClient{svc: svc}, nil
}

func domainMappingsClient(ctx context.Context, creds *auth.Credentials, region string) (DomainMappingsClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewDomainMappingsClient(ctx, region, opt)
}

func (c *gcpDomainMappingsClient) ListDomainMappings(ctx context.Context, project, region string) ([]*DomainMapping, error) {
	parent := fmt.Sprintf("namespaces/%s", project)
	resp, err := c.svc.Namespaces.Domainmappings.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list domain mappings: %w", err)
	}
	var mappings []*DomainMapping
	for _, dm := range resp.Items {
		mappings = append(mappings, domainMappingFromAPI(dm, region))
	}
	return mappings, nil
}

func (c *gcpDomainMappingsClient) GetDomainMapping(ctx context.Context, project, region, name string) (*DomainMapping, error) {
	fullName := fmt.Sprintf("namespaces/%s/domainmappings/%s", project, name)
	dm, err := c.svc.Namespaces.Domainmappings.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get domain mapping %s: %w", name, err)
	}
	return domainMappingFromAPI(dm, region), nil
}

func (c *gcpDomainMappingsClient) CreateDomainMapping(ctx context.Context, project, region string, req *CreateDomainMappingRequest) (*DomainMapping, error) {
	parent := fmt.Sprintf("namespaces/%s", project)
	dm, err := c.svc.Namespaces.Domainmappings.Create(parent, &runv1.DomainMapping{
		Metadata: &runv1.ObjectMeta{
			Name:      req.DomainName,
			Namespace: project,
		},
		Spec: &runv1.DomainMappingSpec{
			RouteName: req.ServiceName,
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create domain mapping %s: %w", req.DomainName, err)
	}
	return domainMappingFromAPI(dm, region), nil
}

func (c *gcpDomainMappingsClient) DeleteDomainMapping(ctx context.Context, project, region, name string) error {
	fullName := fmt.Sprintf("namespaces/%s/domainmappings/%s", project, name)
	if _, err := c.svc.Namespaces.Domainmappings.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete domain mapping %s: %w", name, err)
	}
	return nil
}

func domainMappingFromAPI(dm *runv1.DomainMapping, region string) *DomainMapping {
	if dm == nil {
		return nil
	}
	var status string
	var mappedTo string
	if dm.Status != nil {
		for _, cond := range dm.Status.Conditions {
			if cond.Type == "Ready" {
				status = cond.Status
				break
			}
		}
		mappedTo = dm.Status.MappedRouteName
	}
	name := ""
	if dm.Metadata != nil {
		name = dm.Metadata.Name
	}
	return &DomainMapping{
		Name:     name,
		Region:   region,
		Status:   status,
		MappedTo: mappedTo,
	}
}

func newDomainMappingsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain-mappings",
		Short: "Manage Cloud Run domain mappings",
	}

	cmd.AddCommand(
		newDomainMappingsListCommand(cfg, creds),
		newDomainMappingsDescribeCommand(cfg, creds),
		newDomainMappingsCreateCommand(cfg, creds),
		newDomainMappingsDeleteCommand(cfg, creds),
	)

	return cmd
}

func newDomainMappingsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloud Run domain mappings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := domainMappingsClient(ctx, creds, region)
			if err != nil {
				return err
			}

			mappings, err := client.ListDomainMappings(ctx, project, region)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), mappings)
			}

			headers := []string{"NAME", "STATUS", "MAPPED_TO"}
			rows := make([][]string, len(mappings))
			for i, m := range mappings {
				rows[i] = []string{m.Name, m.Status, m.MappedTo}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")

	return cmd
}

func newDomainMappingsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a Cloud Run domain mapping",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := domainMappingsClient(ctx, creds, region)
			if err != nil {
				return err
			}

			dm, err := client.GetDomainMapping(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), dm)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")

	return cmd
}

func newDomainMappingsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var service string

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a Cloud Run domain mapping",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}
			if service == "" {
				return fmt.Errorf("--service is required")
			}

			ctx := context.Background()
			client, err := domainMappingsClient(ctx, creds, region)
			if err != nil {
				return err
			}

			dm, err := client.CreateDomainMapping(ctx, project, region, &CreateDomainMappingRequest{
				DomainName:  args[0],
				ServiceName: service,
			})
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), dm)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	cmd.Flags().StringVar(&service, "service", "", "Cloud Run service to map to")
	_ = cmd.MarkFlagRequired("service")

	return cmd
}

func newDomainMappingsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a Cloud Run domain mapping",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := domainMappingsClient(ctx, creds, region)
			if err != nil {
				return err
			}

			if err := client.DeleteDomainMapping(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted domain mapping %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")

	return cmd
}
