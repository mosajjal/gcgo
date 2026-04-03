package sql

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesOperations(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))
	var operationsFound, exportFound, importFound, cloneFound, promoteFound, updateFound, failoverFound bool

	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "operations":
			operationsFound = true
		case "instances":
			for _, nested := range sub.Commands() {
				switch nested.Name() {
				case "export":
					exportFound = true
				case "import":
					importFound = true
				case "clone":
					cloneFound = true
				case "promote-replica":
					promoteFound = true
				case "update":
					updateFound = true
				case "failover":
					failoverFound = true
				}
			}
		}
	}
	if !operationsFound || !exportFound || !importFound || !cloneFound || !promoteFound || !updateFound || !failoverFound {
		t.Fatalf("operations=%v export=%v import=%v clone=%v promote=%v update=%v failover=%v", operationsFound, exportFound, importFound, cloneFound, promoteFound, updateFound, failoverFound)
	}
}
