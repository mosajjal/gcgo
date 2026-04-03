package functions

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	funcs     []*Function
	funcMap   map[string]*Function
	listErr   error
	getErr    error
	deployErr error
	deleteErr error
	callErr   error
	callResp  string
}

func (m *mockClient) List(_ context.Context, _, _ string) ([]*Function, error) {
	return m.funcs, m.listErr
}

func (m *mockClient) Get(_ context.Context, _, _, name string) (*Function, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	f, ok := m.funcMap[name]
	if !ok {
		return nil, fmt.Errorf("function %q not found", name)
	}
	return f, nil
}

func (m *mockClient) Deploy(_ context.Context, _, _ string, _ *DeployRequest) error {
	return m.deployErr
}

func (m *mockClient) Delete(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) Call(_ context.Context, _, _, _ string, _ *CallRequest) (string, error) {
	return m.callResp, m.callErr
}

func TestMockListFunctions(t *testing.T) {
	mock := &mockClient{
		funcs: []*Function{
			{Name: "fn-1", State: "ACTIVE", Runtime: "go121", Region: "us-central1"},
			{Name: "fn-2", State: "ACTIVE", Runtime: "python312", Region: "us-central1"},
		},
	}

	funcs, err := mock.List(context.Background(), "proj", "us-central1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(funcs) != 2 {
		t.Errorf("expected 2 functions, got %d", len(funcs))
	}
}

func TestMockListFunctionsError(t *testing.T) {
	mock := &mockClient{listErr: fmt.Errorf("permission denied")}

	_, err := mock.List(context.Background(), "proj", "us-central1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetFunction(t *testing.T) {
	mock := &mockClient{
		funcMap: map[string]*Function{
			"fn-1": {Name: "fn-1", State: "ACTIVE", Runtime: "go121"},
		},
	}

	f, err := mock.Get(context.Background(), "proj", "us-central1", "fn-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if f.Runtime != "go121" {
		t.Errorf("runtime: got %q", f.Runtime)
	}

	_, err = mock.Get(context.Background(), "proj", "us-central1", "nope")
	if err == nil {
		t.Fatal("expected error for missing function")
	}
}

func TestMockDeploy(t *testing.T) {
	mock := &mockClient{}
	err := mock.Deploy(context.Background(), "proj", "us-central1", &DeployRequest{
		Name:    "fn-1",
		Runtime: "go121",
	})
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}
}

func TestMockDeployError(t *testing.T) {
	mock := &mockClient{deployErr: fmt.Errorf("quota exceeded")}
	err := mock.Deploy(context.Background(), "proj", "us-central1", &DeployRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockDelete(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"success", nil},
		{"error", fmt.Errorf("not found")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{deleteErr: tt.err}
			err := mock.Delete(context.Background(), "proj", "us-central1", "fn-1")
			if tt.err != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockCall(t *testing.T) {
	mock := &mockClient{callResp: `{"result":"ok"}`}
	resp, err := mock.Call(context.Background(), "proj", "us-central1", "fn-1", &CallRequest{Data: "test"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if resp != `{"result":"ok"}` {
		t.Errorf("response: got %q", resp)
	}
}

func TestMockCallError(t *testing.T) {
	mock := &mockClient{callErr: fmt.Errorf("function not found")}
	_, err := mock.Call(context.Background(), "proj", "us-central1", "fn-1", &CallRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}
