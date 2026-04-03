package pubsub

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the pubsub command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubsub",
		Short: "Manage Cloud Pub/Sub",
	}

	cmd.AddCommand(
		newTopicsCommand(cfg, creds),
		newSubscriptionsCommand(cfg, creds),
		newSchemasCommand(cfg, creds),
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

func pubsubClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

// --- Topics ---

func newTopicsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topics",
		Short: "Manage Pub/Sub topics",
	}

	cmd.AddCommand(
		newTopicsListCommand(cfg, creds),
		newTopicsDescribeCommand(cfg, creds),
		newTopicsCreateCommand(cfg, creds),
		newTopicsDeleteCommand(cfg, creds),
		newTopicsPublishCommand(cfg, creds),
	)

	return cmd
}

func newTopicsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List topics",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			topics, err := client.ListTopics(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), topics)
			}

			headers := []string{"NAME"}
			rows := make([][]string, len(topics))
			for i, t := range topics {
				rows[i] = []string{t.Name}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newTopicsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe TOPIC",
		Short: "Describe a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			topic, err := client.GetTopic(ctx, project, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), topic)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", topic.Name)
			return nil
		},
	}
}

func newTopicsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "create TOPIC",
		Short: "Create a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			topic, err := client.CreateTopic(ctx, project, args[0])
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created topic %s.\n", topic.Name)
			return nil
		},
	}
}

func newTopicsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete TOPIC",
		Short: "Delete a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteTopic(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted topic %s.\n", args[0])
			return nil
		},
	}
}

func newTopicsPublishCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var attrSlice []string

	cmd := &cobra.Command{
		Use:   "publish TOPIC --message=MESSAGE",
		Short: "Publish a message to a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			message, _ := cmd.Flags().GetString("message")
			attrs := make(map[string]string)
			for _, a := range attrSlice {
				k, v, ok := strings.Cut(a, "=")
				if !ok {
					return fmt.Errorf("invalid attribute format %q (expected KEY=VALUE)", a)
				}
				attrs[k] = v
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			id, err := client.Publish(ctx, project, args[0], message, attrs)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Published message %s.\n", id)
			return nil
		},
	}

	cmd.Flags().String("message", "", "Message data")
	cmd.Flags().StringSliceVar(&attrSlice, "attribute", nil, "Message attributes (KEY=VALUE)")
	_ = cmd.MarkFlagRequired("message")

	return cmd
}

// --- Subscriptions ---

func newSubscriptionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Manage Pub/Sub subscriptions",
	}

	cmd.AddCommand(
		newSubsListCommand(cfg, creds),
		newSubsDescribeCommand(cfg, creds),
		newSubsCreateCommand(cfg, creds),
		newSubsDeleteCommand(cfg, creds),
		newSubsPullCommand(cfg, creds),
	)

	return cmd
}

func newSubsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List subscriptions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			subs, err := client.ListSubscriptions(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), subs)
			}

			headers := []string{"NAME", "TOPIC", "ACK_DEADLINE"}
			rows := make([][]string, len(subs))
			for i, s := range subs {
				rows[i] = []string{s.Name, s.Topic, fmt.Sprintf("%ds", s.AckDeadlineSeconds)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newSubsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe SUBSCRIPTION",
		Short: "Describe a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			sub, err := client.GetSubscription(ctx, project, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), sub)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", sub.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Topic:        %s\n", sub.Topic)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Ack Deadline: %ds\n", sub.AckDeadlineSeconds)
			return nil
		},
	}
}

func newSubsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create SUBSCRIPTION --topic=TOPIC",
		Short: "Create a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			topic, _ := cmd.Flags().GetString("topic")
			ackDeadline, _ := cmd.Flags().GetInt("ack-deadline")

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			sub, err := client.CreateSubscription(ctx, project, args[0], topic, ackDeadline)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created subscription %s.\n", sub.Name)
			return nil
		},
	}

	cmd.Flags().String("topic", "", "Topic to subscribe to")
	cmd.Flags().Int("ack-deadline", 10, "Ack deadline in seconds")
	_ = cmd.MarkFlagRequired("topic")

	return cmd
}

func newSubsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete SUBSCRIPTION",
		Short: "Delete a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteSubscription(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted subscription %s.\n", args[0])
			return nil
		},
	}
}

func newSubsPullCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull SUBSCRIPTION",
		Short: "Pull messages from a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			maxMessages, _ := cmd.Flags().GetInt("max-messages")

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			msgs, err := client.Pull(ctx, project, args[0], maxMessages)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), msgs)
			}

			headers := []string{"ID", "DATA", "PUBLISH_TIME"}
			rows := make([][]string, len(msgs))
			for i, m := range msgs {
				rows[i] = []string{m.ID, m.Data, m.PublishTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().Int("max-messages", 10, "Maximum number of messages to pull")

	return cmd
}

// --- Schemas ---

func newSchemasCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schemas",
		Short: "Manage Pub/Sub schemas",
	}

	cmd.AddCommand(
		newSchemasListCommand(cfg, creds),
		newSchemasDescribeCommand(cfg, creds),
		newSchemasCreateCommand(cfg, creds),
		newSchemasDeleteCommand(cfg, creds),
	)

	return cmd
}

func newSchemasListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List schemas",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			schemas, err := client.ListSchemas(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), schemas)
			}

			headers := []string{"NAME", "TYPE"}
			rows := make([][]string, len(schemas))
			for i, s := range schemas {
				rows[i] = []string{s.Name, s.Type}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newSchemasDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe SCHEMA",
		Short: "Describe a schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			schema, err := client.GetSchema(ctx, project, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), schema)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:       %s\n", schema.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:       %s\n", schema.Type)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Definition: %s\n", schema.Definition)
			return nil
		},
	}
}

func newSchemasCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var schemaType string
	var definition string

	cmd := &cobra.Command{
		Use:   "create SCHEMA",
		Short: "Create a schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if definition == "" {
				return fmt.Errorf("--definition is required")
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			schema, err := client.CreateSchema(ctx, project, args[0], schemaType, definition)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created schema %s.\n", schema.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&schemaType, "type", "AVRO", "Schema type (AVRO or PROTOCOL_BUFFER)")
	cmd.Flags().StringVar(&definition, "definition", "", "Schema definition")
	_ = cmd.MarkFlagRequired("definition")

	return cmd
}

func newSchemasDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete SCHEMA",
		Short: "Delete a schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := pubsubClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteSchema(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted schema %s.\n", args[0])
			return nil
		},
	}
}
