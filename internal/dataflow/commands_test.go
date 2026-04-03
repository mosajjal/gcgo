package dataflow

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesSnapshotsAndMessages(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))

	var jobsFound, messagesFound, metricsFound, snapshotsFound, flexTemplatesFound, launchFound, templatesFound, templateLaunchFound bool
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "jobs":
			jobsFound = true
			for _, nested := range sub.Commands() {
				switch nested.Name() {
				case "messages":
					messagesFound = true
				case "metrics":
					metricsFound = true
				}
			}
		case "snapshots":
			snapshotsFound = true
		case "flex-templates":
			flexTemplatesFound = true
			for _, nested := range sub.Commands() {
				if nested.Name() == "launch" {
					launchFound = true
				}
			}
		case "templates":
			templatesFound = true
			for _, nested := range sub.Commands() {
				if nested.Name() == "launch" {
					templateLaunchFound = true
				}
			}
		}
	}

	if !jobsFound || !messagesFound || !metricsFound || !snapshotsFound || !flexTemplatesFound || !launchFound || !templatesFound || !templateLaunchFound {
		t.Fatalf("jobs=%v messages=%v metrics=%v snapshots=%v flexTemplates=%v launch=%v templates=%v templateLaunch=%v", jobsFound, messagesFound, metricsFound, snapshotsFound, flexTemplatesFound, launchFound, templatesFound, templateLaunchFound)
	}
}
