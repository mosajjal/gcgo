package eventarc

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the eventarc command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eventarc",
		Short: "Manage Eventarc resources",
	}

	cmd.AddCommand(newTriggersCommand(cfg, creds))
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

func requireLocation(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("location")
	if flagVal != "" {
		return flagVal, nil
	}
	location := cfg.Region()
	if location == "" {
		return "", fmt.Errorf("--location is required (or set region in config)")
	}
	return location, nil
}

func eventarcClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newTriggersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "triggers",
		Short: "Manage Eventarc triggers",
	}

	cmd.AddCommand(
		newTriggersListCommand(cfg, creds),
		newTriggersDescribeCommand(cfg, creds),
		newTriggersCreateCommand(cfg, creds),
		newTriggersDeleteCommand(cfg, creds),
	)
	return cmd
}

func newTriggersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List triggers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := eventarcClient(ctx, creds)
			if err != nil {
				return err
			}

			triggers, err := client.ListTriggers(ctx, project, location)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), triggers)
			}

			headers := []string{"NAME", "DESTINATION", "EVENT_TYPE", "CHANNEL", "UPDATED"}
			rows := make([][]string, len(triggers))
			for i, trigger := range triggers {
				rows[i] = []string{trigger.Name, trigger.Destination, trigger.EventType, trigger.Channel, trigger.UpdateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Eventarc location")
	return cmd
}

func newTriggersDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "describe TRIGGER",
		Short: "Describe a trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := eventarcClient(ctx, creds)
			if err != nil {
				return err
			}

			trigger, err := client.GetTrigger(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), trigger)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:           %s\n", trigger.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Destination:    %s\n", trigger.Destination)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Event Type:     %s\n", trigger.EventType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Channel:        %s\n", trigger.Channel)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ServiceAccount: %s\n", trigger.ServiceAccount)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Filters:        %s\n", filtersString(trigger.EventFilters))
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Eventarc location")
	return cmd
}

func newTriggersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var eventType string
	var filters []string
	var workflow string
	var cloudRunService string
	var cloudRunRegion string
	var cloudRunPath string
	var httpEndpoint string
	var serviceAccount string
	var channel string
	var contentType string

	cmd := &cobra.Command{
		Use:   "create TRIGGER",
		Short: "Create a trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if eventType == "" {
				return fmt.Errorf("--event-type is required")
			}

			req := &CreateTriggerRequest{
				Name:                 args[0],
				EventType:            eventType,
				Workflow:             workflow,
				CloudRunService:      cloudRunService,
				CloudRunRegion:       cloudRunRegion,
				CloudRunPath:         cloudRunPath,
				HttpEndpoint:         httpEndpoint,
				ServiceAccount:       serviceAccount,
				Channel:              channel,
				EventDataContentType: contentType,
			}

			for _, raw := range filters {
				filter, err := parseFilter(raw)
				if err != nil {
					return err
				}
				req.EventFilters = append(req.EventFilters, &filter)
			}

			ctx := context.Background()
			client, err := eventarcClient(ctx, creds)
			if err != nil {
				return err
			}

			trigger, err := client.CreateTrigger(ctx, project, location, req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created trigger %s.\n", trigger.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Eventarc location")
	cmd.Flags().StringVar(&eventType, "event-type", "", "CloudEvent type filter")
	cmd.Flags().StringArrayVar(&filters, "filter", nil, "Additional filter as attribute=value or attribute:operator=value")
	cmd.Flags().StringVar(&workflow, "workflow", "", "Workflow resource name destination")
	cmd.Flags().StringVar(&cloudRunService, "cloud-run-service", "", "Cloud Run service destination")
	cmd.Flags().StringVar(&cloudRunRegion, "cloud-run-region", "", "Cloud Run region")
	cmd.Flags().StringVar(&cloudRunPath, "cloud-run-path", "", "Cloud Run path")
	cmd.Flags().StringVar(&httpEndpoint, "http-endpoint", "", "HTTP endpoint destination URI")
	cmd.Flags().StringVar(&serviceAccount, "service-account", "", "Trigger service account")
	cmd.Flags().StringVar(&channel, "channel", "", "Channel resource name")
	cmd.Flags().StringVar(&contentType, "event-data-content-type", "application/json", "Event data content type")
	_ = cmd.MarkFlagRequired("event-type")
	return cmd
}

func newTriggersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "delete TRIGGER",
		Short: "Delete a trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := eventarcClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteTrigger(ctx, project, location, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted trigger %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Eventarc location")
	return cmd
}

func parseFilter(raw string) (EventFilter, error) {
	left, value, ok := strings.Cut(raw, "=")
	if !ok {
		return EventFilter{}, fmt.Errorf("invalid filter %q, expected attribute=value", raw)
	}
	attribute, operator, hasOp := strings.Cut(left, ":")
	filter := EventFilter{Attribute: attribute, Value: value}
	if hasOp {
		filter.Operator = operator
	}
	return filter, nil
}

func filtersString(filters []*EventFilter) string {
	if len(filters) == 0 {
		return ""
	}

	parts := make([]string, 0, len(filters))
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		if filter.Operator != "" {
			parts = append(parts, fmt.Sprintf("%s:%s=%s", filter.Attribute, filter.Operator, filter.Value))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", filter.Attribute, filter.Value))
	}
	return strings.Join(parts, ",")
}
