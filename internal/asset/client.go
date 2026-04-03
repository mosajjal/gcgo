package asset

import (
	"context"
	"fmt"
	"time"

	cloudasset "google.golang.org/api/cloudasset/v1"
	"google.golang.org/api/option"
)

// Resource holds search result fields for a resource.
type Resource struct {
	Name        string   `json:"name"`
	AssetType   string   `json:"asset_type"`
	Project     string   `json:"project"`
	DisplayName string   `json:"display_name"`
	Location    string   `json:"location"`
	Labels      []string `json:"labels,omitempty"`
}

// IAMPolicyResult holds search result fields for an IAM policy.
type IAMPolicyResult struct {
	Resource string `json:"resource"`
	Project  string `json:"project"`
	Policy   string `json:"policy"`
}

// ExportResult holds the output path from an export operation.
type ExportResult struct {
	OutputURI string `json:"output_uri"`
}

// AnalyzeIAMPolicyRequest holds IAM analysis parameters.
type AnalyzeIamPolicyRequest struct {
	Identity                       string
	Permission                     string
	ResourceName                   string
	Roles                          []string
	ExpandGroups                   bool
	ExpandResources                bool
	ExpandRoles                    bool
	OutputGroupEdges               bool
	OutputResourceEdges            bool
	AnalyzeServiceAccountImpersonation bool
	SavedAnalysisQuery             string
	AccessTime                     string
	ExecutionTimeout               string
}

// Feed holds asset feed metadata.
type Feed struct {
	Name        string   `json:"name"`
	ContentType string   `json:"content_type"`
	Topic       string   `json:"topic,omitempty"`
	AssetTypes  []string `json:"asset_types,omitempty"`
}

// Client defines asset inventory operations.
type Client interface {
	SearchAllResources(ctx context.Context, scope, query string, assetTypes []string) ([]*Resource, error)
	SearchAllIAMPolicies(ctx context.Context, scope, query string) ([]*IAMPolicyResult, error)
	AnalyzeIamPolicy(ctx context.Context, scope string, req *AnalyzeIamPolicyRequest) (*cloudasset.AnalyzeIamPolicyResponse, error)
	Export(ctx context.Context, parent, outputURI string, assetTypes []string, contentType string) (*ExportResult, error)
	ListFeeds(ctx context.Context, parent string) ([]*Feed, error)
	GetFeed(ctx context.Context, name string) (*Feed, error)
	CreateFeed(ctx context.Context, parent, feedID, topic string, assetTypes []string, contentType string) (*Feed, error)
	DeleteFeed(ctx context.Context, name string) error
}

type gcpClient struct {
	ac *cloudasset.Service
}

// NewClient creates a Client backed by the real GCP API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	ac, err := cloudasset.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create asset client: %w", err)
	}
	return &gcpClient{ac: ac}, nil
}

func (c *gcpClient) SearchAllResources(ctx context.Context, scope, query string, assetTypes []string) ([]*Resource, error) {
	call := c.ac.V1.SearchAllResources(scope)
	if query != "" {
		call = call.Query(query)
	}
	if len(assetTypes) > 0 {
		call = call.AssetTypes(assetTypes...)
	}

	var resources []*Resource
	if err := call.Pages(ctx, func(resp *cloudasset.SearchAllResourcesResponse) error {
		for _, r := range resp.Results {
			res := &Resource{
				Name:        r.Name,
				AssetType:   r.AssetType,
				Project:     r.Project,
				DisplayName: r.DisplayName,
				Location:    r.Location,
			}
			for k, v := range r.Labels {
				res.Labels = append(res.Labels, k+"="+v)
			}
			resources = append(resources, res)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("search resources: %w", err)
	}
	return resources, nil
}

func (c *gcpClient) SearchAllIAMPolicies(ctx context.Context, scope, query string) ([]*IAMPolicyResult, error) {
	call := c.ac.V1.SearchAllIamPolicies(scope)
	if query != "" {
		call = call.Query(query)
	}

	var results []*IAMPolicyResult
	if err := call.Pages(ctx, func(resp *cloudasset.SearchAllIamPoliciesResponse) error {
		for _, r := range resp.Results {
			results = append(results, &IAMPolicyResult{
				Resource: r.Resource,
				Project:  r.Project,
				Policy:   fmt.Sprintf("%+v", r.Policy),
			})
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("search iam policies: %w", err)
	}
	return results, nil
}

func (c *gcpClient) AnalyzeIamPolicy(ctx context.Context, scope string, req *AnalyzeIamPolicyRequest) (*cloudasset.AnalyzeIamPolicyResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("analyze iam policy: nil request")
	}
	call := c.ac.V1.AnalyzeIamPolicy(scope)
	if req.Identity != "" {
		call = call.AnalysisQueryIdentitySelectorIdentity(req.Identity)
	}
	if req.Permission != "" {
		call = call.AnalysisQueryAccessSelectorPermissions(req.Permission)
	}
	if len(req.Roles) > 0 {
		call = call.AnalysisQueryAccessSelectorRoles(req.Roles...)
	}
	if req.ResourceName != "" {
		call = call.AnalysisQueryResourceSelectorFullResourceName(req.ResourceName)
	}
	if req.AccessTime != "" {
		call = call.AnalysisQueryConditionContextAccessTime(req.AccessTime)
	}
	if req.SavedAnalysisQuery != "" {
		call = call.SavedAnalysisQuery(req.SavedAnalysisQuery)
	}
	call = call.AnalysisQueryOptionsExpandGroups(req.ExpandGroups)
	call = call.AnalysisQueryOptionsExpandResources(req.ExpandResources)
	call = call.AnalysisQueryOptionsExpandRoles(req.ExpandRoles)
	call = call.AnalysisQueryOptionsOutputGroupEdges(req.OutputGroupEdges)
	call = call.AnalysisQueryOptionsOutputResourceEdges(req.OutputResourceEdges)
	call = call.AnalysisQueryOptionsAnalyzeServiceAccountImpersonation(req.AnalyzeServiceAccountImpersonation)
	if req.ExecutionTimeout != "" {
		call = call.ExecutionTimeout(req.ExecutionTimeout)
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("analyze iam policy: %w", err)
	}
	return resp, nil
}

func (c *gcpClient) Export(ctx context.Context, parent, outputURI string, assetTypes []string, contentType string) (*ExportResult, error) {
	req := &cloudasset.ExportAssetsRequest{
		AssetTypes:  assetTypes,
		ContentType: exportContentType(contentType),
		OutputConfig: &cloudasset.OutputConfig{
			GcsDestination: &cloudasset.GcsDestination{Uri: outputURI},
		},
	}

	op, err := c.ac.V1.ExportAssets(parent, req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("export assets: %w", err)
	}
	if op == nil {
		return &ExportResult{OutputURI: outputURI}, nil
	}

	// The generated API returns a generic operation. Poll it briefly so callers
	// see a consistent success/failure signal without depending on a helper that
	// does not exist in the generated client.
	for !op.Done {
		time.Sleep(500 * time.Millisecond)
		updated, err := c.ac.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("wait for asset export: %w", err)
		}
		op = updated
	}
	if op.Error != nil {
		return nil, fmt.Errorf("wait for asset export: %s", op.Error.Message)
	}
	return &ExportResult{OutputURI: outputURI}, nil
}

func (c *gcpClient) ListFeeds(ctx context.Context, parent string) ([]*Feed, error) {
	resp, err := c.ac.Feeds.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list asset feeds: %w", err)
	}
	var feeds []*Feed
	for _, feed := range resp.Feeds {
		feeds = append(feeds, feedFromAPI(feed))
	}
	return feeds, nil
}

func (c *gcpClient) GetFeed(ctx context.Context, name string) (*Feed, error) {
	feed, err := c.ac.Feeds.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get asset feed %s: %w", name, err)
	}
	return feedFromAPI(feed), nil
}

func (c *gcpClient) CreateFeed(ctx context.Context, parent, feedID, topic string, assetTypes []string, contentType string) (*Feed, error) {
	feed, err := c.ac.Feeds.Create(parent, &cloudasset.CreateFeedRequest{
		FeedId: feedID,
		Feed: &cloudasset.Feed{
			AssetTypes:  assetTypes,
			ContentType: exportContentType(contentType),
			FeedOutputConfig: &cloudasset.FeedOutputConfig{
				PubsubDestination: &cloudasset.PubsubDestination{Topic: topic},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create asset feed %s: %w", feedID, err)
	}
	return feedFromAPI(feed), nil
}

func (c *gcpClient) DeleteFeed(ctx context.Context, name string) error {
	if _, err := c.ac.Feeds.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete asset feed %s: %w", name, err)
	}
	return nil
}

func exportContentType(contentType string) string {
	switch contentType {
	case "resource":
		return "RESOURCE"
	case "iam-policy":
		return "IAM_POLICY"
	case "org-policy":
		return "ORG_POLICY"
	case "access-policy":
		return "ACCESS_POLICY"
	default:
		return ""
	}
}

func feedFromAPI(feed *cloudasset.Feed) *Feed {
	if feed == nil {
		return nil
	}
	out := &Feed{
		Name:        feed.Name,
		ContentType: feed.ContentType,
		AssetTypes:  feed.AssetTypes,
	}
	if feed.FeedOutputConfig != nil && feed.FeedOutputConfig.PubsubDestination != nil {
		out.Topic = feed.FeedOutputConfig.PubsubDestination.Topic
	}
	return out
}
