package kms

import (
	"reflect"
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

func TestCommandTreeIncludesIAMAndAsymmetricSign(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))
	tests := []struct {
		path []string
		want bool
	}{
		{path: []string{"keys", "iam", "get-policy"}, want: true},
		{path: []string{"keys", "iam", "set-policy"}, want: true},
		{path: []string{"keys", "iam", "test-permissions"}, want: true},
		{path: []string{"keyrings", "iam", "get-policy"}, want: true},
		{path: []string{"keyrings", "iam", "set-policy"}, want: true},
		{path: []string{"keyrings", "iam", "test-permissions"}, want: true},
		{path: []string{"keys", "versions", "asymmetric-sign"}, want: true},
		{path: []string{"keys", "versions", "create"}, want: true},
	}

	for _, tt := range tests {
		if got := hasCommandPath(cmd, tt.path...); got != tt.want {
			t.Fatalf("path %v: got %v want %v", tt.path, got, tt.want)
		}
	}
}

func TestApplyKMSBinding(t *testing.T) {
	t.Parallel()

	base := &cloudkms.Policy{
		Bindings: []*cloudkms.Binding{
			{Role: "roles/viewer", Members: []string{"user:alice@example.com"}},
		},
	}

	updated := applyKMSBinding(cloneKMSPolicy(base), "user:bob@example.com", "roles/viewer", false)
	wantAdd := []string{"user:alice@example.com", "user:bob@example.com"}
	if !reflect.DeepEqual(updated.Bindings[0].Members, wantAdd) {
		t.Fatalf("add members = %v, want %v", updated.Bindings[0].Members, wantAdd)
	}

	removed := applyKMSBinding(cloneKMSPolicy(updated), "user:alice@example.com", "roles/viewer", true)
	wantRemove := []string{"user:bob@example.com"}
	if !reflect.DeepEqual(removed.Bindings[0].Members, wantRemove) {
		t.Fatalf("remove members = %v, want %v", removed.Bindings[0].Members, wantRemove)
	}

	deleted := applyKMSBinding(cloneKMSPolicy(base), "user:alice@example.com", "roles/viewer", true)
	if len(deleted.Bindings) != 0 {
		t.Fatalf("expected binding removal, got %#v", deleted.Bindings)
	}
}

func hasCommandPath(cmd *cobra.Command, path ...string) bool {
	if len(path) == 0 {
		return true
	}
	for _, sub := range cmd.Commands() {
		if sub.Name() == path[0] {
			return hasCommandPath(sub, path[1:]...)
		}
	}
	return false
}

func cloneKMSPolicy(policy *cloudkms.Policy) *cloudkms.Policy {
	if policy == nil {
		return nil
	}
	cloned := &cloudkms.Policy{Etag: policy.Etag}
	if len(policy.Bindings) == 0 {
		return cloned
	}
	cloned.Bindings = make([]*cloudkms.Binding, len(policy.Bindings))
	for i, binding := range policy.Bindings {
		cloned.Bindings[i] = &cloudkms.Binding{
			Role:    binding.Role,
			Members: append([]string(nil), binding.Members...),
		}
	}
	return cloned
}
