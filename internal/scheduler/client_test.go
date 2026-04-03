package scheduler

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	jobs   []*Job
	jobMap map[string]*Job

	listErr   error
	getErr    error
	createErr error
	deleteErr error
	pauseErr  error
	resumeErr error
	runErr    error
}

func (m *mockClient) ListJobs(_ context.Context, _, _ string) ([]*Job, error) {
	return m.jobs, m.listErr
}

func (m *mockClient) GetJob(_ context.Context, _, _, id string) (*Job, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	j, ok := m.jobMap[id]
	if !ok {
		return nil, fmt.Errorf("job %q not found", id)
	}
	return j, nil
}

func (m *mockClient) CreateJob(_ context.Context, _, _ string, req *CreateJobRequest) (*Job, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &Job{
		Name:     req.Name,
		Schedule: req.Schedule,
		TimeZone: req.TimeZone,
		State:    "ENABLED",
	}, nil
}

func (m *mockClient) DeleteJob(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) PauseJob(_ context.Context, _, _, _ string) error {
	return m.pauseErr
}

func (m *mockClient) ResumeJob(_ context.Context, _, _, _ string) error {
	return m.resumeErr
}

func (m *mockClient) RunJob(_ context.Context, _, _, _ string) error {
	return m.runErr
}

func TestMockListJobs(t *testing.T) {
	mock := &mockClient{
		jobs: []*Job{
			{Name: "job-1", Schedule: "*/5 * * * *", TimeZone: "UTC", State: "ENABLED"},
			{Name: "job-2", Schedule: "0 9 * * 1", TimeZone: "US/Eastern", State: "PAUSED"},
		},
	}

	jobs, err := mock.ListJobs(context.Background(), "proj", "us-central1")
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestMockListJobsError(t *testing.T) {
	mock := &mockClient{listErr: fmt.Errorf("permission denied")}

	_, err := mock.ListJobs(context.Background(), "proj", "us-central1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetJob(t *testing.T) {
	mock := &mockClient{
		jobMap: map[string]*Job{
			"job-1": {Name: "job-1", Schedule: "*/5 * * * *", State: "ENABLED"},
		},
	}

	job, err := mock.GetJob(context.Background(), "proj", "us-central1", "job-1")
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if job.Schedule != "*/5 * * * *" {
		t.Errorf("schedule: got %q", job.Schedule)
	}

	_, err = mock.GetJob(context.Background(), "proj", "us-central1", "nope")
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestMockCreateJob(t *testing.T) {
	mock := &mockClient{}

	job, err := mock.CreateJob(context.Background(), "proj", "us-central1", &CreateJobRequest{
		Name:     "job-1",
		Schedule: "*/5 * * * *",
		URI:      "https://example.com/callback",
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if job.Name != "job-1" {
		t.Errorf("name: got %q", job.Name)
	}
}

func TestMockCreateJobError(t *testing.T) {
	mock := &mockClient{createErr: fmt.Errorf("already exists")}

	_, err := mock.CreateJob(context.Background(), "proj", "us-central1", &CreateJobRequest{Name: "dup"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockDeleteJob(t *testing.T) {
	mock := &mockClient{}
	if err := mock.DeleteJob(context.Background(), "proj", "us-central1", "job-1"); err != nil {
		t.Fatalf("delete job: %v", err)
	}
}

func TestMockJobLifecycle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func(Client) error
		err  error
	}{
		{
			name: "pause success",
			fn:   func(c Client) error { return c.PauseJob(ctx, "p", "l", "j") },
		},
		{
			name: "resume success",
			fn:   func(c Client) error { return c.ResumeJob(ctx, "p", "l", "j") },
		},
		{
			name: "run success",
			fn:   func(c Client) error { return c.RunJob(ctx, "p", "l", "j") },
		},
		{
			name: "pause error",
			fn:   func(c Client) error { return c.PauseJob(ctx, "p", "l", "j") },
			err:  fmt.Errorf("not found"),
		},
		{
			name: "resume error",
			fn:   func(c Client) error { return c.ResumeJob(ctx, "p", "l", "j") },
			err:  fmt.Errorf("not found"),
		},
		{
			name: "run error",
			fn:   func(c Client) error { return c.RunJob(ctx, "p", "l", "j") },
			err:  fmt.Errorf("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				pauseErr:  tt.err,
				resumeErr: tt.err,
				runErr:    tt.err,
			}
			err := tt.fn(mock)
			if tt.err != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
