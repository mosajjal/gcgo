package pubsub

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	topics    []*Topic
	topicMap  map[string]*Topic
	subs      []*Subscription
	subMap    map[string]*Subscription
	schemas   []*Schema
	schemaMap map[string]*Schema
	msgs      []*ReceivedMessage

	listTopicsErr  error
	getTopicErr    error
	createTopicErr error
	deleteTopicErr error
	publishErr     error
	publishID      string

	listSubsErr      error
	getSubErr        error
	createSubErr     error
	deleteSubErr     error
	pullErr          error
	ackErr           error
	modifyAckErr     error
	seekErr          error

	listSchemasErr  error
	getSchemaErr    error
	createSchemaErr error
	deleteSchemaErr error
}

func (m *mockClient) ListTopics(_ context.Context, _ string) ([]*Topic, error) {
	return m.topics, m.listTopicsErr
}

func (m *mockClient) GetTopic(_ context.Context, _, id string) (*Topic, error) {
	if m.getTopicErr != nil {
		return nil, m.getTopicErr
	}
	t, ok := m.topicMap[id]
	if !ok {
		return nil, fmt.Errorf("topic %q not found", id)
	}
	return t, nil
}

func (m *mockClient) CreateTopic(_ context.Context, _, id string) (*Topic, error) {
	if m.createTopicErr != nil {
		return nil, m.createTopicErr
	}
	return &Topic{Name: id}, nil
}

func (m *mockClient) DeleteTopic(_ context.Context, _, _ string) error {
	return m.deleteTopicErr
}

func (m *mockClient) Publish(_ context.Context, _, _, _ string, _ map[string]string) (string, error) {
	if m.publishErr != nil {
		return "", m.publishErr
	}
	return m.publishID, nil
}

func (m *mockClient) ListSubscriptions(_ context.Context, _ string) ([]*Subscription, error) {
	return m.subs, m.listSubsErr
}

func (m *mockClient) GetSubscription(_ context.Context, _, id string) (*Subscription, error) {
	if m.getSubErr != nil {
		return nil, m.getSubErr
	}
	s, ok := m.subMap[id]
	if !ok {
		return nil, fmt.Errorf("subscription %q not found", id)
	}
	return s, nil
}

func (m *mockClient) CreateSubscription(_ context.Context, _, subID, topicID string, ackDeadline int) (*Subscription, error) {
	if m.createSubErr != nil {
		return nil, m.createSubErr
	}
	return &Subscription{Name: subID, Topic: topicID, AckDeadlineSeconds: ackDeadline}, nil
}

func (m *mockClient) DeleteSubscription(_ context.Context, _, _ string) error {
	return m.deleteSubErr
}

func (m *mockClient) Pull(_ context.Context, _, _ string, _ int) ([]*ReceivedMessage, error) {
	if m.pullErr != nil {
		return nil, m.pullErr
	}
	return m.msgs, nil
}

func (m *mockClient) AcknowledgeMessages(_ context.Context, _, _ string, _ []string) error {
	return m.ackErr
}

func (m *mockClient) ModifyAckDeadline(_ context.Context, _, _ string, _ []string, _ int32) error {
	return m.modifyAckErr
}

func (m *mockClient) SeekSubscription(_ context.Context, _, _ string, _ *SeekRequest) error {
	return m.seekErr
}

func (m *mockClient) ListSchemas(_ context.Context, _ string) ([]*Schema, error) {
	return m.schemas, m.listSchemasErr
}

func (m *mockClient) GetSchema(_ context.Context, _, id string) (*Schema, error) {
	if m.getSchemaErr != nil {
		return nil, m.getSchemaErr
	}
	s, ok := m.schemaMap[id]
	if !ok {
		return nil, fmt.Errorf("schema %q not found", id)
	}
	return s, nil
}

func (m *mockClient) CreateSchema(_ context.Context, _, id, schemaType, def string) (*Schema, error) {
	if m.createSchemaErr != nil {
		return nil, m.createSchemaErr
	}
	return &Schema{Name: id, Type: schemaType, Definition: def}, nil
}

func (m *mockClient) DeleteSchema(_ context.Context, _, _ string) error {
	return m.deleteSchemaErr
}

func TestMockListTopics(t *testing.T) {
	mock := &mockClient{
		topics: []*Topic{
			{Name: "topic-1"},
			{Name: "topic-2"},
		},
	}

	topics, err := mock.ListTopics(context.Background(), "proj")
	if err != nil {
		t.Fatalf("list topics: %v", err)
	}
	if len(topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(topics))
	}
}

func TestMockListTopicsError(t *testing.T) {
	mock := &mockClient{listTopicsErr: fmt.Errorf("permission denied")}

	_, err := mock.ListTopics(context.Background(), "proj")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetTopic(t *testing.T) {
	mock := &mockClient{
		topicMap: map[string]*Topic{
			"topic-1": {Name: "topic-1"},
		},
	}

	topic, err := mock.GetTopic(context.Background(), "proj", "topic-1")
	if err != nil {
		t.Fatalf("get topic: %v", err)
	}
	if topic.Name != "topic-1" {
		t.Errorf("name: got %q", topic.Name)
	}

	_, err = mock.GetTopic(context.Background(), "proj", "nope")
	if err == nil {
		t.Fatal("expected error for missing topic")
	}
}

func TestMockCreateTopic(t *testing.T) {
	mock := &mockClient{}

	topic, err := mock.CreateTopic(context.Background(), "proj", "new-topic")
	if err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if topic.Name != "new-topic" {
		t.Errorf("name: got %q", topic.Name)
	}
}

func TestMockCreateTopicError(t *testing.T) {
	mock := &mockClient{createTopicErr: fmt.Errorf("already exists")}

	_, err := mock.CreateTopic(context.Background(), "proj", "dup")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockDeleteTopic(t *testing.T) {
	mock := &mockClient{}
	if err := mock.DeleteTopic(context.Background(), "proj", "topic-1"); err != nil {
		t.Fatalf("delete topic: %v", err)
	}
}

func TestMockPublish(t *testing.T) {
	mock := &mockClient{publishID: "msg-123"}

	id, err := mock.Publish(context.Background(), "proj", "topic-1", "hello", nil)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if id != "msg-123" {
		t.Errorf("id: got %q", id)
	}
}

func TestMockPublishError(t *testing.T) {
	mock := &mockClient{publishErr: fmt.Errorf("topic not found")}

	_, err := mock.Publish(context.Background(), "proj", "topic-1", "hello", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockListSubscriptions(t *testing.T) {
	mock := &mockClient{
		subs: []*Subscription{
			{Name: "sub-1", Topic: "topic-1", AckDeadlineSeconds: 10},
		},
	}

	subs, err := mock.ListSubscriptions(context.Background(), "proj")
	if err != nil {
		t.Fatalf("list subs: %v", err)
	}
	if len(subs) != 1 {
		t.Errorf("expected 1 sub, got %d", len(subs))
	}
}

func TestMockGetSubscription(t *testing.T) {
	mock := &mockClient{
		subMap: map[string]*Subscription{
			"sub-1": {Name: "sub-1", Topic: "topic-1", AckDeadlineSeconds: 10},
		},
	}

	sub, err := mock.GetSubscription(context.Background(), "proj", "sub-1")
	if err != nil {
		t.Fatalf("get sub: %v", err)
	}
	if sub.Topic != "topic-1" {
		t.Errorf("topic: got %q", sub.Topic)
	}

	_, err = mock.GetSubscription(context.Background(), "proj", "nope")
	if err == nil {
		t.Fatal("expected error for missing sub")
	}
}
