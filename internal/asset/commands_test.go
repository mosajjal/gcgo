package asset

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
)

func TestCommandTreeIncludesFeeds(t *testing.T) {
	cmd := NewCommand(auth.New(""))
	for _, sub := range cmd.Commands() {
		if sub.Name() == "feeds" {
			var createFound, deleteFound bool
			for _, nested := range sub.Commands() {
				switch nested.Name() {
				case "create":
					createFound = true
				case "delete":
					deleteFound = true
				}
			}
			if !createFound || !deleteFound {
				t.Fatalf("create=%v delete=%v", createFound, deleteFound)
			}
			return
		}
	}
	t.Fatal("expected feeds command to be wired")
}

func TestCommandTreeIncludesAnalyzeIamPolicy(t *testing.T) {
	cmd := NewCommand(auth.New(""))
	for _, sub := range cmd.Commands() {
		if sub.Name() == "analyze-iam-policy" {
			return
		}
	}
	t.Fatal("expected analyze-iam-policy command to be wired")
}
