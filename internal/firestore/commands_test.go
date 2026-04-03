package firestore

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesOperations(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))
	var createFound, deleteFound, indexesFound, operationsFound bool

	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "create":
			createFound = true
		case "delete":
			deleteFound = true
		case "indexes":
			indexesFound = true
		case "operations":
			operationsFound = true
		}
	}
	if !createFound || !deleteFound || !indexesFound || !operationsFound {
		t.Fatalf("create=%v delete=%v indexes=%v operations=%v", createFound, deleteFound, indexesFound, operationsFound)
	}
}
