package pubsub

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	pubsubapi "google.golang.org/api/pubsub/v1"
)

// Topic holds Pub/Sub topic fields.
type Topic struct {
	Name string `json:"name"`
}

// Subscription holds Pub/Sub subscription fields.
type Subscription struct {
	Name               string `json:"name"`
	Topic              string `json:"topic"`
	AckDeadlineSeconds int    `json:"ack_deadline_seconds"`
}

// Schema holds Pub/Sub schema fields.
type Schema struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Definition string `json:"definition"`
}

// ReceivedMessage holds a pulled message.
type ReceivedMessage struct {
	ID          string            `json:"id"`
	Data        string            `json:"data"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	PublishTime string            `json:"publish_time"`
}

// Client defines Pub/Sub operations.
type Client interface {
	ListTopics(ctx context.Context, project string) ([]*Topic, error)
	GetTopic(ctx context.Context, project, topicID string) (*Topic, error)
	CreateTopic(ctx context.Context, project, topicID string) (*Topic, error)
	DeleteTopic(ctx context.Context, project, topicID string) error
	Publish(ctx context.Context, project, topicID, message string, attrs map[string]string) (string, error)

	ListSubscriptions(ctx context.Context, project string) ([]*Subscription, error)
	GetSubscription(ctx context.Context, project, subID string) (*Subscription, error)
	CreateSubscription(ctx context.Context, project, subID, topicID string, ackDeadline int) (*Subscription, error)
	DeleteSubscription(ctx context.Context, project, subID string) error
	Pull(ctx context.Context, project, subID string, maxMessages int) ([]*ReceivedMessage, error)

	ListSchemas(ctx context.Context, project string) ([]*Schema, error)
	GetSchema(ctx context.Context, project, schemaID string) (*Schema, error)
	CreateSchema(ctx context.Context, project, schemaID, schemaType, definition string) (*Schema, error)
	DeleteSchema(ctx context.Context, project, schemaID string) error
}

type gcpClient struct {
	newClientFn     func(ctx context.Context, project string, opts ...option.ClientOption) (*pubsub.Client, error)
	newRESTClientFn func(ctx context.Context, opts ...option.ClientOption) (*pubsubapi.Service, error)
	opts            []option.ClientOption
}

// NewClient creates a Client backed by the real Pub/Sub API.
func NewClient(_ context.Context, opts ...option.ClientOption) (Client, error) {
	return &gcpClient{
		newClientFn:     pubsub.NewClient,
		newRESTClientFn: pubsubapi.NewService,
		opts:            opts,
	}, nil
}

func (c *gcpClient) projectClient(ctx context.Context, project string) (*pubsub.Client, error) {
	pc, err := c.newClientFn(ctx, project, c.opts...)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}
	return pc, nil
}

func (c *gcpClient) restClient(ctx context.Context) (*pubsubapi.Service, error) {
	svc, err := c.newRESTClientFn(ctx, c.opts...)
	if err != nil {
		return nil, fmt.Errorf("create pubsub schema client: %w", err)
	}
	return svc, nil
}

func (c *gcpClient) ListTopics(ctx context.Context, project string) ([]*Topic, error) {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	it := pc.Topics(ctx)
	var topics []*Topic
	for {
		t, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list topics: %w", err)
		}
		topics = append(topics, &Topic{Name: t.ID()})
	}
	return topics, nil
}

func (c *gcpClient) GetTopic(ctx context.Context, project, topicID string) (*Topic, error) {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	t := pc.Topic(topicID)
	exists, err := t.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("check topic %s: %w", topicID, err)
	}
	if !exists {
		return nil, fmt.Errorf("topic %s not found", topicID)
	}
	return &Topic{Name: t.ID()}, nil
}

func (c *gcpClient) CreateTopic(ctx context.Context, project, topicID string) (*Topic, error) {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	t, err := pc.CreateTopic(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("create topic %s: %w", topicID, err)
	}
	return &Topic{Name: t.ID()}, nil
}

func (c *gcpClient) DeleteTopic(ctx context.Context, project, topicID string) error {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return err
	}
	defer pc.Close()

	if err := pc.Topic(topicID).Delete(ctx); err != nil {
		return fmt.Errorf("delete topic %s: %w", topicID, err)
	}
	return nil
}

func (c *gcpClient) Publish(ctx context.Context, project, topicID, message string, attrs map[string]string) (string, error) {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return "", err
	}
	defer pc.Close()

	result := pc.Topic(topicID).Publish(ctx, &pubsub.Message{
		Data:       []byte(message),
		Attributes: attrs,
	})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("publish to topic %s: %w", topicID, err)
	}
	return id, nil
}

func (c *gcpClient) ListSubscriptions(ctx context.Context, project string) ([]*Subscription, error) {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	it := pc.Subscriptions(ctx)
	var subs []*Subscription
	for {
		s, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list subscriptions: %w", err)
		}
		cfg, err := s.Config(ctx)
		if err != nil {
			return nil, fmt.Errorf("get subscription config %s: %w", s.ID(), err)
		}
		subs = append(subs, &Subscription{
			Name:               s.ID(),
			Topic:              cfg.Topic.ID(),
			AckDeadlineSeconds: int(cfg.AckDeadline.Seconds()),
		})
	}
	return subs, nil
}

func (c *gcpClient) GetSubscription(ctx context.Context, project, subID string) (*Subscription, error) {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	s := pc.Subscription(subID)
	exists, err := s.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("check subscription %s: %w", subID, err)
	}
	if !exists {
		return nil, fmt.Errorf("subscription %s not found", subID)
	}
	cfg, err := s.Config(ctx)
	if err != nil {
		return nil, fmt.Errorf("get subscription config %s: %w", subID, err)
	}
	return &Subscription{
		Name:               s.ID(),
		Topic:              cfg.Topic.ID(),
		AckDeadlineSeconds: int(cfg.AckDeadline.Seconds()),
	}, nil
}

func (c *gcpClient) CreateSubscription(ctx context.Context, project, subID, topicID string, ackDeadline int) (*Subscription, error) {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	deadline := time.Duration(ackDeadline) * time.Second
	if deadline == 0 {
		deadline = 10 * time.Second
	}

	s, err := pc.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
		Topic:       pc.Topic(topicID),
		AckDeadline: deadline,
	})
	if err != nil {
		return nil, fmt.Errorf("create subscription %s: %w", subID, err)
	}
	cfg, err := s.Config(ctx)
	if err != nil {
		return nil, fmt.Errorf("get subscription config %s: %w", subID, err)
	}
	return &Subscription{
		Name:               s.ID(),
		Topic:              cfg.Topic.ID(),
		AckDeadlineSeconds: int(cfg.AckDeadline.Seconds()),
	}, nil
}

func (c *gcpClient) DeleteSubscription(ctx context.Context, project, subID string) error {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return err
	}
	defer pc.Close()

	if err := pc.Subscription(subID).Delete(ctx); err != nil {
		return fmt.Errorf("delete subscription %s: %w", subID, err)
	}
	return nil
}

func (c *gcpClient) Pull(ctx context.Context, project, subID string, maxMessages int) ([]*ReceivedMessage, error) {
	pc, err := c.projectClient(ctx, project)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	if maxMessages <= 0 {
		maxMessages = 10
	}

	sub := pc.Subscription(subID)
	sub.ReceiveSettings.MaxOutstandingMessages = maxMessages
	sub.ReceiveSettings.Synchronous = true

	pullCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var msgs []*ReceivedMessage
	err = sub.Receive(pullCtx, func(_ context.Context, m *pubsub.Message) {
		msgs = append(msgs, &ReceivedMessage{
			ID:          m.ID,
			Data:        string(m.Data),
			Attributes:  m.Attributes,
			PublishTime: m.PublishTime.Format(time.RFC3339),
		})
		m.Ack()
		if len(msgs) >= maxMessages {
			cancel()
		}
	})
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return nil, fmt.Errorf("pull from subscription %s: %w", subID, err)
	}
	return msgs, nil
}

func (c *gcpClient) ListSchemas(ctx context.Context, project string) ([]*Schema, error) {
	svc, err := c.restClient(ctx)
	if err != nil {
		return nil, err
	}

	var schemas []*Schema
	parent := fmt.Sprintf("projects/%s", project)
	call := svc.Projects.Schemas.List(parent).View("BASIC")
	if err := call.Pages(ctx, func(resp *pubsubapi.ListSchemasResponse) error {
		for _, s := range resp.Schemas {
			schemas = append(schemas, &Schema{
				Name: s.Name,
				Type: s.Type,
			})
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list schemas: %w", err)
	}
	return schemas, nil
}

func (c *gcpClient) GetSchema(ctx context.Context, project, schemaID string) (*Schema, error) {
	svc, err := c.restClient(ctx)
	if err != nil {
		return nil, err
	}

	s, err := svc.Projects.Schemas.Get(fmt.Sprintf("projects/%s/schemas/%s", project, schemaID)).View("FULL").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get schema %s: %w", schemaID, err)
	}
	return &Schema{
		Name:       s.Name,
		Type:       s.Type,
		Definition: s.Definition,
	}, nil
}

func (c *gcpClient) CreateSchema(ctx context.Context, project, schemaID, schemaType, definition string) (*Schema, error) {
	svc, err := c.restClient(ctx)
	if err != nil {
		return nil, err
	}

	s, err := svc.Projects.Schemas.Create(fmt.Sprintf("projects/%s", project), &pubsubapi.Schema{
		Type:       schemaType,
		Definition: definition,
	}).SchemaId(schemaID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create schema %s: %w", schemaID, err)
	}
	return &Schema{
		Name:       s.Name,
		Type:       s.Type,
		Definition: s.Definition,
	}, nil
}

func (c *gcpClient) DeleteSchema(ctx context.Context, project, schemaID string) error {
	svc, err := c.restClient(ctx)
	if err != nil {
		return err
	}

	if _, err := svc.Projects.Schemas.Delete(fmt.Sprintf("projects/%s/schemas/%s", project, schemaID)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete schema %s: %w", schemaID, err)
	}
	return nil
}
