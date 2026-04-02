package storage

import "testing"

func FuzzParseGSURI(f *testing.F) {
	f.Add("gs://bucket/path")
	f.Add("gs://bucket")
	f.Add("gs://")
	f.Add("")
	f.Add("s3://bucket/path")
	f.Add("gs://a/b/c/d/e/f")
	f.Add("gs://bucket-with-dashes/prefix/obj.txt")

	f.Fuzz(func(t *testing.T, input string) {
		uri, err := ParseGSURI(input)
		if err != nil {
			return // invalid input is fine
		}

		// Bucket must not be empty if parse succeeded
		if uri.Bucket == "" {
			t.Error("parsed successfully but bucket is empty")
		}
	})
}
