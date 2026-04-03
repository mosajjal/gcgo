package logging

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesSinks(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))

	want := map[string][]string{
		"sinks":      {"list", "describe", "create", "delete"},
		"metrics":    {"list", "describe", "create", "delete"},
		"exclusions": {"list", "describe", "create", "update", "delete"},
		"buckets":    {"list", "describe", "create", "update", "delete"},
	}

	for _, sub := range cmd.Commands() {
		names, ok := want[sub.Name()]
		if !ok {
			continue
		}
		for _, nested := range sub.Commands() {
			for i, name := range names {
				if nested.Name() == name {
					names[i] = ""
				}
			}
		}
		for _, name := range names {
			if name != "" {
				t.Fatalf("expected %s command to include %q", sub.Name(), name)
			}
		}
		delete(want, sub.Name())
	}

	if len(want) != 0 {
		t.Fatalf("expected placeholder groups to be wired: %v", want)
	}
}
