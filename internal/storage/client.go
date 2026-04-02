package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Object holds GCS object fields.
type Object struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Updated string `json:"updated"`
}

// BucketInfo holds bucket fields.
type BucketInfo struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Created  string `json:"created"`
}

// GSURI parses a gs://bucket/prefix URI.
type GSURI struct {
	Bucket string
	Prefix string
}

// ParseGSURI parses gs://bucket/path into components.
func ParseGSURI(s string) (GSURI, error) {
	if !strings.HasPrefix(s, "gs://") {
		return GSURI{}, fmt.Errorf("invalid gs:// URI: %q", s)
	}
	rest := strings.TrimPrefix(s, "gs://")
	bucket, prefix, _ := strings.Cut(rest, "/")
	if bucket == "" {
		return GSURI{}, fmt.Errorf("empty bucket in URI: %q", s)
	}
	return GSURI{Bucket: bucket, Prefix: prefix}, nil
}

// Client defines storage operations.
type Client interface {
	ListBuckets(ctx context.Context, project string) ([]*BucketInfo, error)
	ListObjects(ctx context.Context, bucket, prefix string) ([]*Object, error)
	Upload(ctx context.Context, bucket, object string, r io.Reader) error
	Download(ctx context.Context, bucket, object string, w io.Writer) error
	Delete(ctx context.Context, bucket, object string) error
	CreateBucket(ctx context.Context, project, bucket, location string) error
	DeleteBucket(ctx context.Context, bucket string) error
}

type gcpClient struct {
	sc *storage.Client
}

// NewClient creates a Client backed by the real GCS API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	sc, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}
	return &gcpClient{sc: sc}, nil
}

func (c *gcpClient) ListBuckets(ctx context.Context, project string) ([]*BucketInfo, error) {
	it := c.sc.Buckets(ctx, project)

	var buckets []*BucketInfo
	for {
		b, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list buckets: %w", err)
		}
		buckets = append(buckets, &BucketInfo{
			Name:     b.Name,
			Location: b.Location,
			Created:  b.Created.String(),
		})
	}
	return buckets, nil
}

func (c *gcpClient) ListObjects(ctx context.Context, bucket, prefix string) ([]*Object, error) {
	it := c.sc.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: prefix})

	var objects []*Object
	for {
		o, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list objects: %w", err)
		}
		objects = append(objects, &Object{
			Name:    o.Name,
			Size:    o.Size,
			Updated: o.Updated.String(),
		})
	}
	return objects, nil
}

func (c *gcpClient) Upload(ctx context.Context, bucket, object string, r io.Reader) error {
	w := c.sc.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return fmt.Errorf("upload %s/%s: %w", bucket, object, err)
	}
	return w.Close()
}

func (c *gcpClient) Download(ctx context.Context, bucket, object string, w io.Writer) error {
	r, err := c.sc.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("download %s/%s: %w", bucket, object, err)
	}
	defer func() { _ = r.Close() }()

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("download %s/%s: %w", bucket, object, err)
	}
	return nil
}

func (c *gcpClient) Delete(ctx context.Context, bucket, object string) error {
	if err := c.sc.Bucket(bucket).Object(object).Delete(ctx); err != nil {
		return fmt.Errorf("delete %s/%s: %w", bucket, object, err)
	}
	return nil
}

func (c *gcpClient) CreateBucket(ctx context.Context, project, bucket, location string) error {
	attrs := &storage.BucketAttrs{}
	if location != "" {
		attrs.Location = location
	}
	if err := c.sc.Bucket(bucket).Create(ctx, project, attrs); err != nil {
		return fmt.Errorf("create bucket %s: %w", bucket, err)
	}
	return nil
}

func (c *gcpClient) DeleteBucket(ctx context.Context, bucket string) error {
	if err := c.sc.Bucket(bucket).Delete(ctx); err != nil {
		return fmt.Errorf("delete bucket %s: %w", bucket, err)
	}
	return nil
}

// CopyPath resolves a cp source/destination. Returns (isGCS, bucket, path).
func CopyPath(path string) (isGCS bool, bucket, objPath string, err error) {
	if strings.HasPrefix(path, "gs://") {
		uri, err := ParseGSURI(path)
		if err != nil {
			return false, "", "", err
		}
		return true, uri.Bucket, uri.Prefix, nil
	}
	return false, "", filepath.Clean(path), nil
}

// OpenLocalFile opens a local file for reading.
func OpenLocalFile(path string) (*os.File, error) {
	f, err := os.Open(path) //nolint:gosec // user explicitly provides path as CLI argument
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return f, nil
}

// CreateLocalFile creates a local file for writing.
func CreateLocalFile(path string) (*os.File, error) {
	f, err := os.Create(path) //nolint:gosec // user explicitly provides path as CLI argument
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	return f, nil
}
