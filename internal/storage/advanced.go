package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type uploadPart struct {
	offset int64
	length int64
}

func (c *gcpClient) Copy(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) error {
	_, err := c.sc.Bucket(dstBucket).Object(dstObject).CopierFrom(c.sc.Bucket(srcBucket).Object(srcObject)).Run(ctx)
	if err != nil {
		return fmt.Errorf("copy %s/%s to %s/%s: %w", srcBucket, srcObject, dstBucket, dstObject, err)
	}
	return nil
}

func (c *gcpClient) Move(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) error {
	if srcBucket == dstBucket {
		_, err := c.sc.Bucket(srcBucket).Object(srcObject).Move(ctx, storage.MoveObjectDestination{Object: dstObject})
		if err != nil {
			return fmt.Errorf("move %s/%s to %s/%s: %w", srcBucket, srcObject, dstBucket, dstObject, err)
		}
		return nil
	}

	if err := c.Copy(ctx, srcBucket, srcObject, dstBucket, dstObject); err != nil {
		return err
	}
	if err := c.Delete(ctx, srcBucket, srcObject); err != nil {
		return fmt.Errorf("delete source %s/%s after move: %w", srcBucket, srcObject, err)
	}
	return nil
}

func (c *gcpClient) Cat(ctx context.Context, bucket, object string, w io.Writer) error {
	r, err := c.sc.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("read %s/%s: %w", bucket, object, err)
	}
	defer func() { _ = r.Close() }()

	if _, err := io.Copy(w, r); err != nil {
		return fmt.Errorf("stream %s/%s: %w", bucket, object, err)
	}
	return nil
}

func (c *gcpClient) SignURL(_ context.Context, bucket, object string, duration time.Duration) (string, error) {
	url, err := c.sc.Bucket(bucket).SignedURL(object, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(duration),
	})
	if err != nil {
		return "", fmt.Errorf("sign url %s/%s: %w", bucket, object, err)
	}
	return url, nil
}

func (c *gcpClient) Rsync(ctx context.Context, srcLocal bool, localDir, bucket, prefix string, dryRun bool) ([]RsyncAction, error) {
	if srcLocal {
		return c.rsyncUpload(ctx, localDir, bucket, prefix, dryRun)
	}
	return c.rsyncDownload(ctx, localDir, bucket, prefix, dryRun)
}

func (c *gcpClient) rsyncUpload(ctx context.Context, localDir, bucket, prefix string, dryRun bool) ([]RsyncAction, error) {
	remote := make(map[string]int64)
	it := c.sc.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: prefix})
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list objects: %w", err)
		}
		key := strings.TrimPrefix(attrs.Name, prefix)
		key = strings.TrimPrefix(key, "/")
		remote[key] = attrs.Size
	}

	var actions []RsyncAction
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("resolve relative path: %w", err)
		}
		rel = filepath.ToSlash(rel)

		localSize := info.Size()
		remoteSize, exists := remote[rel]
		if !exists || localSize != remoteSize {
			actions = append(actions, RsyncAction{Action: "upload", Path: rel})
			if !dryRun {
				objName := rel
				if prefix != "" {
					objName = strings.TrimSuffix(prefix, "/") + "/" + rel
				}
				f, err := os.Open(path) //nolint:gosec // user provided path
				if err != nil {
					return fmt.Errorf("open %s: %w", path, err)
				}
				if err := c.Upload(ctx, bucket, objName, f); err != nil {
					_ = f.Close()
					return err
				}
				_ = f.Close()
			}
		}
		delete(remote, rel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk local dir: %w", err)
	}

	return actions, nil
}

func (c *gcpClient) rsyncDownload(ctx context.Context, localDir, bucket, prefix string, dryRun bool) ([]RsyncAction, error) {
	local := make(map[string]int64)
	_ = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("resolve relative path: %w", err)
		}
		local[filepath.ToSlash(rel)] = info.Size()
		return nil
	})

	var actions []RsyncAction
	it := c.sc.Bucket(bucket).Objects(ctx, &storage.Query{Prefix: prefix})
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list objects: %w", err)
		}
		key := strings.TrimPrefix(attrs.Name, prefix)
		key = strings.TrimPrefix(key, "/")
		if key == "" {
			continue
		}

		localSize, exists := local[key]
		if !exists || localSize != attrs.Size {
			actions = append(actions, RsyncAction{Action: "download", Path: key})
			if !dryRun {
				dst := filepath.Join(localDir, filepath.FromSlash(key))
				if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
					return nil, fmt.Errorf("create dir %s: %w", filepath.Dir(dst), err)
				}
				f, err := os.Create(dst) //nolint:gosec // user provided path
				if err != nil {
					return nil, fmt.Errorf("create %s: %w", dst, err)
				}
				if err := c.Download(ctx, bucket, attrs.Name, f); err != nil {
					_ = f.Close()
					return nil, err
				}
				_ = f.Close()
			}
		}
	}

	return actions, nil
}
