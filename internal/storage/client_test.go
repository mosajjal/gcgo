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

func TestSplitUploadParts(t *testing.T) {
	tests := []struct {
		name      string
		size      int64
		wantParts int
		wantLast  int64
	}{
		{name: "small file", size: parallelUploadThreshold, wantParts: 1, wantLast: parallelUploadThreshold},
		{name: "large file", size: parallelUploadThreshold + 1, wantParts: 2, wantLast: 1},
		{name: "multi part", size: parallelUploadChunkSize*3 + 17, wantParts: 4, wantLast: 17},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitUploadParts(tt.size)
			if len(parts) != tt.wantParts {
				t.Fatalf("parts: got %d, want %d", len(parts), tt.wantParts)
			}
			if got := parts[len(parts)-1].length; got != tt.wantLast {
				t.Fatalf("last length: got %d, want %d", got, tt.wantLast)
			}
		})
	}
}

func TestComposeGroups(t *testing.T) {
	tests := []struct {
		name       string
		partCount  int
		wantGroups int
		wantFirst  int
		wantLast   int
	}{
		{name: "single group", partCount: 4, wantGroups: 1, wantFirst: 4, wantLast: 4},
		{name: "two groups", partCount: 33, wantGroups: 2, wantFirst: 32, wantLast: 1},
		{name: "three groups", partCount: 65, wantGroups: 3, wantFirst: 32, wantLast: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := make([]string, tt.partCount)
			for i := range parts {
				parts[i] = "part"
			}
			groups := composeGroups(parts)
			if len(groups) != tt.wantGroups {
				t.Fatalf("groups: got %d, want %d", len(groups), tt.wantGroups)
			}
			if got := len(groups[0]); got != tt.wantFirst {
				t.Fatalf("first group: got %d, want %d", got, tt.wantFirst)
			}
			if got := len(groups[len(groups)-1]); got != tt.wantLast {
				t.Fatalf("last group: got %d, want %d", got, tt.wantLast)
			}
		})
	}
}
