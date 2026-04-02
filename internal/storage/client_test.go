package storage

import "testing"

func TestParseGSURI(t *testing.T) {
	tests := []struct {
		input      string
		wantBucket string
		wantPrefix string
		wantErr    bool
	}{
		{"gs://my-bucket", "my-bucket", "", false},
		{"gs://my-bucket/path/to/obj", "my-bucket", "path/to/obj", false},
		{"gs://my-bucket/", "my-bucket", "", false},
		{"s3://wrong", "", "", true},
		{"gs://", "", "", true},
		{"/local/path", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			uri, err := ParseGSURI(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if uri.Bucket != tt.wantBucket {
				t.Errorf("bucket: got %q, want %q", uri.Bucket, tt.wantBucket)
			}
			if uri.Prefix != tt.wantPrefix {
				t.Errorf("prefix: got %q, want %q", uri.Prefix, tt.wantPrefix)
			}
		})
	}
}

func TestCopyPath(t *testing.T) {
	tests := []struct {
		input    string
		wantGCS  bool
		wantPath string
	}{
		{"gs://bucket/obj", true, "obj"},
		{"/tmp/file.txt", false, "/tmp/file.txt"},
		{"./relative", false, "relative"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			isGCS, _, path, err := CopyPath(tt.input)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if isGCS != tt.wantGCS {
				t.Errorf("isGCS: got %v, want %v", isGCS, tt.wantGCS)
			}
			if path != tt.wantPath {
				t.Errorf("path: got %q, want %q", path, tt.wantPath)
			}
		})
	}
}
