package functions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	cloudfunctions "google.golang.org/api/cloudfunctions/v2"
	"google.golang.org/api/option"
)

// Function holds the fields we care about.
type Function struct {
	Name        string `json:"name"`
	State       string `json:"state"`
	Runtime     string `json:"runtime"`
	EntryPoint  string `json:"entry_point"`
	Region      string `json:"region"`
	URL         string `json:"url"`
	Environment string `json:"environment"`
}

// DeployRequest holds parameters for function deployment.
type DeployRequest struct {
	Name         string
	Runtime      string
	EntryPoint   string
	Source       string
	TriggerHTTP  bool
	TriggerTopic string
	Region       string
	Memory       string
	Timeout      string
}

// CallRequest holds parameters for function invocation.
type CallRequest struct {
	Data string
}

// Client defines the operations we use for Cloud Functions.
type Client interface {
	List(ctx context.Context, project, region string) ([]*Function, error)
	Get(ctx context.Context, project, region, name string) (*Function, error)
	Deploy(ctx context.Context, project, region string, req *DeployRequest) error
	Delete(ctx context.Context, project, region, name string) error
	Call(ctx context.Context, project, region, name string, req *CallRequest) (string, error)
}

type gcpClient struct {
	svc *cloudfunctions.Service
}

// NewClient creates a Client backed by the real Cloud Functions API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := cloudfunctions.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud functions client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) List(ctx context.Context, project, region string) ([]*Function, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	call := c.svc.Projects.Locations.Functions.List(parent).Context(ctx)

	var funcs []*Function
	if err := call.Pages(ctx, func(resp *cloudfunctions.ListFunctionsResponse) error {
		for _, f := range resp.Functions {
			funcs = append(funcs, funcFromAPI(f, region))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list functions: %w", err)
	}
	return funcs, nil
}

func (c *gcpClient) Get(ctx context.Context, project, region, name string) (*Function, error) {
	fullName := fmt.Sprintf("projects/%s/locations/%s/functions/%s", project, region, name)
	f, err := c.svc.Projects.Locations.Functions.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get function %s: %w", name, err)
	}
	return funcFromAPI(f, region), nil
}

func (c *gcpClient) Deploy(ctx context.Context, project, region string, req *DeployRequest) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	fullName := fmt.Sprintf("%s/functions/%s", parent, req.Name)

	pbFunc := &cloudfunctions.Function{
		Name: fullName,
		BuildConfig: &cloudfunctions.BuildConfig{
			Runtime:    req.Runtime,
			EntryPoint: req.EntryPoint,
		},
		ServiceConfig: &cloudfunctions.ServiceConfig{},
	}

	if source := storageSourceFromRequest(req.Source); source != nil {
		pbFunc.BuildConfig.Source = &cloudfunctions.Source{
			StorageSource: source,
		}
	}

	if req.Memory != "" {
		pbFunc.ServiceConfig.AvailableMemory = req.Memory
	}
	if req.Timeout != "" {
		if timeout, err := parseTimeoutSeconds(req.Timeout); err == nil {
			pbFunc.ServiceConfig.TimeoutSeconds = timeout
		}
	}

	if req.TriggerTopic != "" {
		pbFunc.EventTrigger = &cloudfunctions.EventTrigger{
			EventType:   "google.cloud.pubsub.topic.v1.messagePublished",
			PubsubTopic: req.TriggerTopic,
		}
	}

	_, err := c.svc.Projects.Locations.Functions.Create(parent, pbFunc).
		FunctionId(req.Name).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("deploy function %s: %w", req.Name, err)
	}
	return nil
}

func (c *gcpClient) Delete(ctx context.Context, project, region, name string) error {
	fullName := fmt.Sprintf("projects/%s/locations/%s/functions/%s", project, region, name)
	if _, err := c.svc.Projects.Locations.Functions.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete function %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) Call(_ context.Context, _, _, _ string, _ *CallRequest) (string, error) {
	// Cloud Functions v2 does not have a direct "call" RPC.
	return "", fmt.Errorf("call not implemented for v2 functions (invoke the function URL directly)")
}

func funcFromAPI(f *cloudfunctions.Function, region string) *Function {
	fn := &Function{
		Name:        f.Name,
		State:       f.State,
		Region:      region,
		URL:         f.Url,
		Environment: f.Environment,
	}
	if bc := f.BuildConfig; bc != nil {
		fn.Runtime = bc.Runtime
		fn.EntryPoint = bc.EntryPoint
	}
	return fn
}

func storageSourceFromRequest(source string) *cloudfunctions.StorageSource {
	if source == "" {
		return nil
	}
	ss := &cloudfunctions.StorageSource{}
	if strings.HasPrefix(source, "gs://") {
		trimmed := strings.TrimPrefix(source, "gs://")
		bucket, object, ok := strings.Cut(trimmed, "/")
		ss.Bucket = bucket
		if ok {
			ss.Object = object
		}
		return ss
	}
	ss.Bucket = source
	return ss
}

func parseTimeoutSeconds(timeout string) (int64, error) {
	if timeout == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(timeout)
	if err == nil {
		return int64(d.Seconds()), nil
	}
	if seconds, err := strconv.ParseInt(timeout, 10, 64); err == nil {
		return seconds, nil
	}
	return 0, err
}
