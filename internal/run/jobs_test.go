package run

import (
	"context"
	"fmt"
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

type mockJobsClient struct {
	jobs       []*Job
	jobMap     map[string]*Job
	listErr    error
	getErr     error
	createErr  error
	deleteErr  error
	executeErr error
}

func (m *mockJobsClient) ListJobs(_ context.Context, _, _ string) ([]*Job, error) {
	return m.jobs, m.listErr
}

func (m *mockJobsClient) GetJob(_ context.Context, _, _, name string) (*Job, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	job, ok := m.jobMap[name]
	if !ok {
		return nil, fmt.Errorf("job %q not found", name)
	}
	return job, nil
}

func (m *mockJobsClient) CreateJob(_ context.Context, _, _ string, _ *CreateJobRequest) error {
	return m.createErr
}

func (m *mockJobsClient) DeleteJob(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockJobsClient) ExecuteJob(_ context.Context, _, _, _ string) error {
	return m.executeErr
}

type mockExecutionsClient struct {
	executions []*Execution
	execMap    map[string]*Execution
	listErr    error
	getErr     error
	cancelErr  error
}

func (m *mockExecutionsClient) ListExecutions(_ context.Context, _, _, _ string) ([]*Execution, error) {
	return m.executions, m.listErr
}

func (m *mockExecutionsClient) GetExecution(_ context.Context, _, _, _, execution string) (*Execution, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	exec, ok := m.execMap[execution]
	if !ok {
		return nil, fmt.Errorf("execution %q not found", execution)
	}
	return exec, nil
}

func (m *mockExecutionsClient) CancelExecution(_ context.Context, _, _, _, _ string) error {
	return m.cancelErr
}

func TestMockListJobs(t *testing.T) {
	mock := &mockJobsClient{
		jobs: []*Job{
			{Name: "job-1", Region: "us-central1", Image: "gcr.io/proj/img:latest", ExecutionCount: 4},
		},
	}

	jobs, err := mock.ListJobs(context.Background(), "proj", "us-central1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Image != "gcr.io/proj/img:latest" {
		t.Fatalf("image: got %q", jobs[0].Image)
	}
}

func TestMockJobLookups(t *testing.T) {
	mock := &mockJobsClient{
		jobMap: map[string]*Job{
			"job-1": {Name: "job-1", Region: "us-central1"},
		},
	}

	job, err := mock.GetJob(context.Background(), "proj", "us-central1", "job-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if job.Name != "job-1" {
		t.Fatalf("name: got %q", job.Name)
	}

	if _, err := mock.GetJob(context.Background(), "proj", "us-central1", "missing"); err == nil {
		t.Fatal("expected lookup error")
	}
}

func TestMockExecutionLookups(t *testing.T) {
	mock := &mockExecutionsClient{
		execMap: map[string]*Execution{
			"exec-1": {Name: "exec-1", Job: "job-1", Status: "RUNNING"},
		},
	}

	exec, err := mock.GetExecution(context.Background(), "proj", "us-central1", "job-1", "exec-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if exec.Status != "RUNNING" {
		t.Fatalf("status: got %q", exec.Status)
	}

	if _, err := mock.GetExecution(context.Background(), "proj", "us-central1", "job-1", "missing"); err == nil {
		t.Fatal("expected lookup error")
	}
}

func TestRunCommandTreeIncludesJobs(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))
	var jobsFound, revisionsFound bool
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "jobs":
			jobsFound = true
		case "revisions":
			revisionsFound = true
		}
	}
	if !jobsFound || !revisionsFound {
		t.Fatalf("jobs=%v revisions=%v", jobsFound, revisionsFound)
	}
}

func TestJobsCommandIncludesExecutions(t *testing.T) {
	cmd := newJobsCommand(&config.Config{}, auth.New(""))
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Name() == "executions" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected executions command to be wired")
	}
}

func TestServicesCommandIncludesUpdateTraffic(t *testing.T) {
	cmd := newServicesCommand(&config.Config{}, auth.New(""))
	var updateFound, trafficFound, rollbackFound, iamFound bool
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "update":
			updateFound = true
		case "update-traffic":
			trafficFound = true
		case "rollback":
			rollbackFound = true
		case "iam":
			iamFound = true
		}
	}
	if !updateFound || !trafficFound || !rollbackFound || !iamFound {
		t.Fatalf("update=%v updateTraffic=%v rollback=%v iam=%v", updateFound, trafficFound, rollbackFound, iamFound)
	}
}

func TestServicesIAMCommandIncludesPolicySubcommands(t *testing.T) {
	cmd := newServicesIAMCommand(&config.Config{}, auth.New(""))
	var getFound, setFound, testFound bool
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "get-policy":
			getFound = true
		case "set-policy":
			setFound = true
		case "test-permissions":
			testFound = true
		}
	}
	if !getFound || !setFound || !testFound {
		t.Fatalf("get=%v set=%v test=%v", getFound, setFound, testFound)
	}
}
