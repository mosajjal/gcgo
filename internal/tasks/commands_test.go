package tasks

import "testing"

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "two headers",
			input: []string{"Content-Type=application/json", "X-Test=yes"},
			want: map[string]string{
				"Content-Type": "application/json",
				"X-Test":       "yes",
			},
		},
		{name: "invalid header", input: []string{"broken"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, err := parseHeaders(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(headers) != len(tt.want) {
				t.Fatalf("got %d headers want %d", len(headers), len(tt.want))
			}
			for k, v := range tt.want {
				if headers[k] != v {
					t.Fatalf("header %s: got %q want %q", k, headers[k], v)
				}
			}
		})
	}
}
