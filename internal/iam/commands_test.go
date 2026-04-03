package iam

import (
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestCommandTreeIncludesRolesAndWorkloadIdentity(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))

	var rolesFound, workloadFound, foldersFound, orgsFound, projectTroubleshootFound, folderTroubleshootFound, orgTroubleshootFound, denyPoliciesFound, orgPoliciesFound bool
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "deny-policies":
			denyPoliciesFound = true
		case "org-policies":
			orgPoliciesFound = true
		case "roles":
			rolesFound = true
		case "folders":
			foldersFound = true
			for _, nested := range sub.Commands() {
				if nested.Name() == "policy" {
					for _, leaf := range nested.Commands() {
						if leaf.Name() == "troubleshoot" {
							folderTroubleshootFound = true
						}
					}
				}
			}
		case "organizations":
			orgsFound = true
			for _, nested := range sub.Commands() {
				if nested.Name() == "policy" {
					for _, leaf := range nested.Commands() {
						if leaf.Name() == "troubleshoot" {
							orgTroubleshootFound = true
						}
					}
				}
			}
		case "workload-identity":
			workloadFound = true
		case "policy":
			for _, nested := range sub.Commands() {
				if nested.Name() == "troubleshoot" {
					projectTroubleshootFound = true
				}
			}
		}
	}

	if !rolesFound || !workloadFound || !foldersFound || !orgsFound || !projectTroubleshootFound || !folderTroubleshootFound || !orgTroubleshootFound || !denyPoliciesFound || !orgPoliciesFound {
		t.Fatalf("roles=%v workloadIdentity=%v folders=%v organizations=%v projectTroubleshoot=%v folderTroubleshoot=%v orgTroubleshoot=%v denyPolicies=%v orgPolicies=%v",
			rolesFound, workloadFound, foldersFound, orgsFound, projectTroubleshootFound, folderTroubleshootFound, orgTroubleshootFound, denyPoliciesFound, orgPoliciesFound)
	}
}
