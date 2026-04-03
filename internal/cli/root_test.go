package cli

import "testing"

func TestRootCommandIncludesExtendedServices(t *testing.T) {
	cmd := NewRootCommand()

	want := map[string]bool{
		"alloydb":   false,
		"bigtable":  false,
		"deploy":    false,
		"eventarc":  false,
		"firestore": false,
		"redis":     false,
		"services":  false,
		"tasks":     false,
		"workflows": false,
	}

	for _, sub := range cmd.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Fatalf("expected %s command to be wired", name)
		}
	}
}

func TestRootCommandIncludesImpersonationFlag(t *testing.T) {
	cmd := NewRootCommand()
	if cmd.PersistentFlags().Lookup("impersonate-service-account") == nil {
		t.Fatal("expected impersonate-service-account root flag to be registered")
	}
}
