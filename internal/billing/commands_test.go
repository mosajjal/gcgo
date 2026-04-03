package billing

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
)

func TestProjectsCommandIncludesDescribe(t *testing.T) {
	cmd := NewCommand(auth.New(""))

	for _, sub := range cmd.Commands() {
		if sub.Name() != "projects" {
			continue
		}
		for _, nested := range sub.Commands() {
			if nested.Name() == "describe" {
				return
			}
		}
		t.Fatal("expected describe subcommand under projects")
	}

	t.Fatal("expected projects command to be wired")
}

func TestBudgetsCommandIsWired(t *testing.T) {
	cmd := NewCommand(auth.New(""))

	for _, sub := range cmd.Commands() {
		if sub.Name() != "budgets" {
			continue
		}
		var listFound, describeFound, createFound, deleteFound bool
		for _, nested := range sub.Commands() {
			switch nested.Name() {
			case "list":
				listFound = true
			case "describe":
				describeFound = true
			case "create":
				createFound = true
			case "delete":
				deleteFound = true
			}
		}
		if !listFound || !describeFound || !createFound || !deleteFound {
			t.Fatalf("list=%v describe=%v create=%v delete=%v", listFound, describeFound, createFound, deleteFound)
		}
		return
	}

	t.Fatal("expected budgets command to be wired")
}
