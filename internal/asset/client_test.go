package asset

import (
	"context"
	"fmt"
	"testing"

	cloudasset "google.golang.org/api/cloudasset/v1"
)

type mockClient struct {
	resources []*Resource
	policies  []*IAMPolicyResult
	analysis  *cloudasset.AnalyzeIamPolicyResponse
	exported  *ExportResult
	searchErr error
	policyErr error
	analysisErr error
	exportErr error
}

func (m *mockClient) SearchAllResources(_ context.Context, _, _ string, _ []string) ([]*Resource, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.resources, nil
}

func (m *mockClient) SearchAllIAMPolicies(_ context.Context, _, _ string) ([]*IAMPolicyResult, error) {
	if m.policyErr != nil {
		return nil, m.policyErr
	}
	return m.policies, nil
}

func (m *mockClient) AnalyzeIamPolicy(_ context.Context, _ string, _ *AnalyzeIamPolicyRequest) (*cloudasset.AnalyzeIamPolicyResponse, error) {
	if m.analysisErr != nil {
		return nil, m.analysisErr
	}
	if m.analysis != nil {
		return m.analysis, nil
	}
	return &cloudasset.AnalyzeIamPolicyResponse{}, nil
}

func (m *mockClient) Export(_ context.Context, _, _ string, _ []string, _ string) (*ExportResult, error) {
	if m.exportErr != nil {
		return nil, m.exportErr
	}
	if m.exported != nil {
		return m.exported, nil
	}
	return &ExportResult{OutputURI: "gs://bucket/output.json"}, nil
}

func TestMockSearchAllResources(t *testing.T) {
	mock := &mockClient{
		resources: []*Resource{
			{Name: "//compute.googleapis.com/projects/p/zones/z/instances/i", AssetType: "compute.googleapis.com/Instance", Project: "projects/123"},
			{Name: "//storage.googleapis.com/projects/_/buckets/b", AssetType: "storage.googleapis.com/Bucket", Project: "projects/123"},
		},
	}

	resources, err := mock.SearchAllResources(context.Background(), "projects/p", "", nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

func TestMockSearchAllResourcesError(t *testing.T) {
	mock := &mockClient{searchErr: fmt.Errorf("scope not found")}
	_, err := mock.SearchAllResources(context.Background(), "projects/p", "", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockSearchAllResourcesWithFilters(t *testing.T) {
	mock := &mockClient{
		resources: []*Resource{
			{Name: "//compute.googleapis.com/projects/p/zones/z/instances/i", AssetType: "compute.googleapis.com/Instance"},
		},
	}

	resources, err := mock.SearchAllResources(context.Background(), "organizations/123", "name:test", []string{"compute.googleapis.com/Instance"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}
}

func TestMockSearchAllIAMPolicies(t *testing.T) {
	mock := &mockClient{
		policies: []*IAMPolicyResult{
			{Resource: "//cloudresourcemanager.googleapis.com/projects/p", Project: "projects/123"},
		},
	}

	results, err := mock.SearchAllIAMPolicies(context.Background(), "projects/p", "")
	if err != nil {
		t.Fatalf("search policies: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestMockSearchAllIAMPoliciesError(t *testing.T) {
	mock := &mockClient{policyErr: fmt.Errorf("access denied")}
	_, err := mock.SearchAllIAMPolicies(context.Background(), "projects/p", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockExport(t *testing.T) {
	mock := &mockClient{}
	result, err := mock.Export(context.Background(), "projects/p", "gs://bucket/out", nil, "resource")
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if result.OutputURI == "" {
		t.Error("expected non-empty output URI")
	}
}

func TestMockExportError(t *testing.T) {
	mock := &mockClient{exportErr: fmt.Errorf("bucket not writable")}
	_, err := mock.Export(context.Background(), "projects/p", "gs://bucket/out", nil, "")
	if err == nil {
		t.Fatal("expected error")
	}
}
