package auth

import "testing"

func TestAuthCommandIncludesTokenCommands(t *testing.T) {
	cmd := NewCommand(New(""))

	var accessFound, identityFound bool
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "print-access-token":
			accessFound = true
		case "print-identity-token":
			identityFound = true
		}
	}

	if !accessFound || !identityFound {
		t.Fatalf("print-access-token=%v print-identity-token=%v", accessFound, identityFound)
	}
}
