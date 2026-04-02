package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

func FuzzConfigParsing(f *testing.F) {
	f.Add(`project = "my-proj"
account = "me@example.com"
region = "us-central1"
zone = "us-central1-a"
`)
	f.Add(``)
	f.Add(`project = ""`)
	f.Add(`invalid toml {{{{`)

	f.Fuzz(func(t *testing.T, input string) {
		dir := t.TempDir()
		path := filepath.Join(dir, "properties.toml")

		if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
			t.Fatal(err)
		}

		var props Properties
		_, err := toml.DecodeFile(path, &props)
		if err != nil {
			return // invalid TOML is fine
		}

		// If it parsed, all fields should be safe strings
		c := &Config{props: props, path: path}
		_ = c.All()
		_ = c.Project("")
		_ = c.Region()
		_ = c.Zone()
	})
}
