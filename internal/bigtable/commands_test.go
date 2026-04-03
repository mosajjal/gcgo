package bigtable

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesOperations(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))

	var operationsFound, createFound, deleteFound, backupsFound bool
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "operations":
			operationsFound = true
		case "instances":
			for _, nested := range sub.Commands() {
				switch nested.Name() {
				case "create":
					createFound = true
				case "delete":
					deleteFound = true
				case "backups":
					backupsFound = true
				}
			}
		}
	}
	if !operationsFound || !createFound || !deleteFound || !backupsFound {
		t.Fatalf("operations=%v create=%v delete=%v backups=%v", operationsFound, createFound, deleteFound, backupsFound)
	}
}
