package run

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesDomainMappingsPlaceholder(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))

	for _, sub := range cmd.Commands() {
		if sub.Name() != "domain-mappings" {
			continue
		}
		want := map[string]bool{
			"list":     false,
			"describe": false,
			"create":   false,
			"delete":   false,
		}
		for _, nested := range sub.Commands() {
			if _, ok := want[nested.Name()]; ok {
				want[nested.Name()] = true
			}
		}
		for name, ok := range want {
			if !ok {
				t.Fatalf("missing %s command under domain-mappings", name)
			}
		}
		return
	}

	t.Fatal("expected domain-mappings command to be wired")
}
