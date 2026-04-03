package workflows

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesDeploy(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Name() == "deploy" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected deploy command to be wired")
	}
}
