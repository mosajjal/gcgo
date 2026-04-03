package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	parallelUploadThreshold = 5 * 1024 * 1024
	parallelUploadChunkSize = 5 * 1024 * 1024
	maxComposeSources       = 32
	maxUploadWorkers        = 4
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

// RsyncAction describes a single rsync operation.
type RsyncAction struct {
	Action string `json:"action"`
	Path   string `json:"path"`
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
	Copy(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) error
	Move(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) error
	Cat(ctx context.Context, bucket, object string, w io.Writer) error
	SignURL(ctx context.Context, bucket, object string, duration time.Duration) (string, error)
	Rsync(ctx context.Context, srcLocal bool, localDir, bucket, prefix string, dryRun bool) ([]RsyncAction, error)
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
	if file, ok := r.(*os.File); ok {
		info, err := file.Stat()
		if err == nil && info.Size() > parallelUploadThreshold {
			return c.uploadFileParallel(ctx, bucket, object, file, info.Size())
		}
	}

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

func (c *gcpClient) uploadFileParallel(ctx context.Context, bucket, object string, file *os.File, size int64) error {
	parts := splitUploadParts(size)
	if len(parts) == 1 {
		return c.uploadSection(ctx, bucket, object, file, parts[0].offset, parts[0].length)
	}

	tempNames := make([]string, len(parts))
	for i := range parts {
		tempNames[i] = tempUploadName(object, i)
	}

	var mu sync.Mutex
	uploaded := make([]string, 0, len(parts))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxUploadWorkers)
	for i := range parts {
		i := i
		g.Go(func() error {
			part := parts[i]
			if err := c.uploadSection(gctx, bucket, tempNames[i], file, part.offset, part.length); err != nil {
				return err
			}
			mu.Lock()
			uploaded = append(uploaded, tempNames[i])
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		cleanupUploadedParts(ctx, c.sc, bucket, uploaded)
		return err
	}

	defer cleanupUploadedParts(context.Background(), c.sc, bucket, uploaded)
	return c.composeUploadedParts(ctx, bucket, object, tempNames)
}

func (c *gcpClient) uploadSection(ctx context.Context, bucket, object string, file *os.File, offset, length int64) error {
	reader := io.NewSectionReader(file, offset, length)
	w := c.sc.Bucket(bucket).Object(object).NewWriter(ctx)
	w.ChunkSize = parallelUploadChunkSize
	if _, err := io.Copy(w, reader); err != nil {
		_ = w.Close()
		return fmt.Errorf("upload %s/%s: %w", bucket, object, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("finalize upload %s/%s: %w", bucket, object, err)
	}
	return nil
}

func (c *gcpClient) composeUploadedParts(ctx context.Context, bucket, object string, parts []string) error {
	if len(parts) == 0 {
		return fmt.Errorf("compose %s/%s: no source parts", bucket, object)
	}
	groups := composeGroups(parts)
	if len(groups) == 1 {
		return c.composeOnce(ctx, bucket, object, groups[0])
	}

	var temps []string
	defer cleanupUploadedParts(context.Background(), c.sc, bucket, temps)

	next := make([]string, 0, len(groups))
	for i, group := range groups {
		name := tempUploadName(object, i+len(parts))
		if err := c.composeOnce(ctx, bucket, name, group); err != nil {
			return err
		}
		temps = append(temps, name)
		next = append(next, name)
	}
	return c.composeUploadedParts(ctx, bucket, object, next)
}

func (c *gcpClient) composeOnce(ctx context.Context, bucket, dst string, srcs []string) error {
	if len(srcs) == 0 {
		return fmt.Errorf("compose %s/%s: no source objects", bucket, dst)
	}
	handles := make([]*storage.ObjectHandle, len(srcs))
	for i, src := range srcs {
		handles[i] = c.sc.Bucket(bucket).Object(src)
	}
	_, err := c.sc.Bucket(bucket).Object(dst).ComposerFrom(handles...).Run(ctx)
	if err != nil {
		return fmt.Errorf("compose %s/%s: %w", bucket, dst, err)
	}
	return nil
}

func splitUploadParts(size int64) []uploadPart {
	if size <= parallelUploadThreshold {
		return []uploadPart{{offset: 0, length: size}}
	}

	parts := make([]uploadPart, 0, (size+parallelUploadChunkSize-1)/parallelUploadChunkSize)
	for offset := int64(0); offset < size; offset += parallelUploadChunkSize {
		length := int64(parallelUploadChunkSize)
		if remaining := size - offset; remaining < length {
			length = remaining
		}
		parts = append(parts, uploadPart{offset: offset, length: length})
	}
	return parts
}

func composeGroups(parts []string) [][]string {
	if len(parts) <= maxComposeSources {
		return [][]string{parts}
	}

	groups := make([][]string, 0, (len(parts)+maxComposeSources-1)/maxComposeSources)
	for i := 0; i < len(parts); i += maxComposeSources {
		end := i + maxComposeSources
		if end > len(parts) {
			end = len(parts)
		}
		groups = append(groups, parts[i:end])
	}
	return groups
}

func tempUploadName(object string, part int) string {
	safe := strings.NewReplacer("/", "_", ":", "_").Replace(object)
	return fmt.Sprintf(".gcgo-upload/%s/%d-%d", safe, part, time.Now().UnixNano())
}

func cleanupUploadedParts(ctx context.Context, client *storage.Client, bucket string, names []string) {
	for _, name := range names {
		_ = client.Bucket(bucket).Object(name).Delete(ctx)
	}
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
