package secrets

import (
	"reflect"
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/spf13/cobra"
	secretmanager "google.golang.org/api/secretmanager/v1"
)

func TestCommandTreeIncludesIAM(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))
	tests := []struct {
		path []string
		want bool
	}{
		{path: []string{"iam", "get-policy"}, want: true},
		{path: []string{"iam", "set-policy"}, want: true},
		{path: []string{"iam", "test-permissions"}, want: true},
		{path: []string{"versions", "describe"}, want: true},
		{path: []string{"update"}, want: true},
	}

	for _, tt := range tests {
		if got := hasCommandPath(cmd, tt.path...); got != tt.want {
			t.Fatalf("path %v: got %v want %v", tt.path, got, tt.want)
		}
	}
}

func TestApplySecretBinding(t *testing.T) {
	t.Parallel()

	base := &secretmanager.Policy{
		Bindings: []*secretmanager.Binding{
			{Role: "roles/secretmanager.viewer", Members: []string{"user:alice@example.com"}},
		},
	}

	updated := applySecretBinding(cloneSecretPolicy(base), "user:bob@example.com", "roles/secretmanager.viewer", false)
	wantAdd := []string{"user:alice@example.com", "user:bob@example.com"}
	if !reflect.DeepEqual(updated.Bindings[0].Members, wantAdd) {
		t.Fatalf("add members = %v, want %v", updated.Bindings[0].Members, wantAdd)
	}

	removed := applySecretBinding(cloneSecretPolicy(updated), "user:alice@example.com", "roles/secretmanager.viewer", true)
	wantRemove := []string{"user:bob@example.com"}
	if !reflect.DeepEqual(removed.Bindings[0].Members, wantRemove) {
		t.Fatalf("remove members = %v, want %v", removed.Bindings[0].Members, wantRemove)
	}

	deleted := applySecretBinding(cloneSecretPolicy(base), "user:alice@example.com", "roles/secretmanager.viewer", true)
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

func cloneSecretPolicy(policy *secretmanager.Policy) *secretmanager.Policy {
	if policy == nil {
		return nil
	}
	cloned := &secretmanager.Policy{Etag: policy.Etag}
	if len(policy.Bindings) == 0 {
		return cloned
	}
	cloned.Bindings = make([]*secretmanager.Binding, len(policy.Bindings))
	for i, binding := range policy.Bindings {
		cloned.Bindings[i] = &secretmanager.Binding{
			Role:    binding.Role,
			Members: append([]string(nil), binding.Members...),
		}
	}
	return cloned
}
