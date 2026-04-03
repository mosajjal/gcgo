package eventarc

import "testing"

func TestParseFilter(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantAttr  string
		wantOp    string
		wantValue string
		wantErr   bool
	}{
		{name: "simple filter", input: "type=google.cloud.storage.object.v1.finalized", wantAttr: "type", wantValue: "google.cloud.storage.object.v1.finalized"},
		{name: "operator filter", input: "subject:path_pattern=/objects/*", wantAttr: "subject", wantOp: "path_pattern", wantValue: "/objects/*"},
		{name: "invalid", input: "broken", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := parseFilter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if filter.Attribute != tt.wantAttr || filter.Operator != tt.wantOp || filter.Value != tt.wantValue {
				t.Fatalf("got %+v", filter)
			}
		})
	}
}
