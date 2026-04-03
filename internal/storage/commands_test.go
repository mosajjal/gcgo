package storage

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/spf13/cobra"
)

func TestCommandTreeIncludesPlaceholderAdminGroups(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))

	want := map[string][]string{
		"iam":       {"get-policy", "set-policy", "test-permissions"},
		"lifecycle": {"describe", "update"},
		"retention": {"describe", "update", "lock"},
	}

	for _, sub := range cmd.Commands() {
		names, ok := want[sub.Name()]
		if !ok {
			continue
		}
		if !hasCommands(sub, names...) {
			t.Fatalf("expected %s subcommands %v", sub.Name(), names)
		}
		delete(want, sub.Name())
	}

	if len(want) != 0 {
		t.Fatalf("missing placeholder groups: %v", want)
	}
}

func hasCommands(cmd interface{ Commands() []*cobra.Command }, names ...string) bool {
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		wanted[name] = false
	}
	for _, sub := range cmd.Commands() {
		if _, ok := wanted[sub.Name()]; ok {
			wanted[sub.Name()] = true
		}
	}
	for _, ok := range wanted {
		if !ok {
			return false
		}
	}
	return true
}
