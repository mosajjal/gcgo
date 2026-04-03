package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

type mockAdvancedClient struct {
	moveErr    error
	catData    string
	catErr     error
	signURL    string
	signErr    error
	rsyncActs  []RsyncAction
	rsyncErr   error
	moveCalled bool
	catBucket  string
	catObject  string
}

func (m *mockAdvancedClient) Move(_ context.Context, srcBucket, srcObject, dstBucket, dstObject string) error {
	m.moveCalled = true
	_ = srcBucket
	_ = srcObject
	_ = dstBucket
	_ = dstObject
	return m.moveErr
}

func (m *mockAdvancedClient) Cat(_ context.Context, bucket, object string, w io.Writer) error {
	m.catBucket = bucket
	m.catObject = object
	if m.catErr != nil {
		return m.catErr
	}
	_, err := io.WriteString(w, m.catData)
	return err
}

func (m *mockAdvancedClient) SignURL(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return m.signURL, m.signErr
}

func (m *mockAdvancedClient) Rsync(_ context.Context, _ bool, _, _, _ string, _ bool) ([]RsyncAction, error) {
	return m.rsyncActs, m.rsyncErr
}

func TestMockMove(t *testing.T) {
	tests := []struct {
		name    string
		moveErr error
		wantErr bool
	}{
		{"success", nil, false},
		{"error", fmt.Errorf("permission denied"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAdvancedClient{moveErr: tt.moveErr}
			err := mock.Move(context.Background(), "src-bucket", "src/obj", "dst-bucket", "dst/obj")
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !mock.moveCalled {
				t.Error("Move was not called")
			}
		})
	}
}

func TestMockCat(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		catErr     error
		wantOutput string
		wantErr    bool
	}{
		{"success", "hello world\n", nil, "hello world\n", false},
		{"empty", "", nil, "", false},
		{"error", "", fmt.Errorf("not found"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAdvancedClient{catData: tt.data, catErr: tt.catErr}
			var buf strings.Builder
			err := mock.Cat(context.Background(), "bucket", "obj.txt", &buf)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if buf.String() != tt.wantOutput {
				t.Errorf("output: got %q, want %q", buf.String(), tt.wantOutput)
			}
			if mock.catBucket != "bucket" {
				t.Errorf("bucket: got %q", mock.catBucket)
			}
			if mock.catObject != "obj.txt" {
				t.Errorf("object: got %q", mock.catObject)
			}
		})
	}
}

func TestMockSignURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		signErr error
		wantErr bool
	}{
		{"success", "https://storage.googleapis.com/bucket/obj?signed=abc", nil, false},
		{"error", "", fmt.Errorf("no service account"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAdvancedClient{signURL: tt.url, signErr: tt.signErr}
			url, err := mock.SignURL(context.Background(), "bucket", "obj", 1*time.Hour)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if url != tt.url {
				t.Errorf("url: got %q, want %q", url, tt.url)
			}
		})
	}
}

func TestMockRsync(t *testing.T) {
	tests := []struct {
		name    string
		actions []RsyncAction
		err     error
		wantErr bool
		wantN   int
	}{
		{
			"upload actions",
			[]RsyncAction{
				{Action: "upload", Path: "file1.txt"},
				{Action: "upload", Path: "subdir/file2.txt"},
			},
			nil, false, 2,
		},
		{
			"download actions",
			[]RsyncAction{
				{Action: "download", Path: "remote.txt"},
			},
			nil, false, 1,
		},
		{
			"already in sync",
			nil, nil, false, 0,
		},
		{
			"error",
			nil, fmt.Errorf("access denied"), true, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAdvancedClient{rsyncActs: tt.actions, rsyncErr: tt.err}
			actions, err := mock.Rsync(context.Background(), true, "/tmp/dir", "bucket", "prefix/", false)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(actions) != tt.wantN {
				t.Errorf("actions count: got %d, want %d", len(actions), tt.wantN)
			}
		})
	}
}

func TestRsyncActionFields(t *testing.T) {
	a := RsyncAction{Action: "upload", Path: "dir/file.txt"}
	if a.Action != "upload" {
		t.Errorf("action: got %q", a.Action)
	}
	if a.Path != "dir/file.txt" {
		t.Errorf("path: got %q", a.Path)
	}
}
