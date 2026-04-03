package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const helpLastUpdated = "2026-04-03"

type helpMetadata struct {
	Context string
	Docs    []string
}

var serviceHelp = map[string]helpMetadata{
	"auth": {
		Context: "Use this group for authentication flows, credential inspection, and revocation. Prefer ADC-compatible flows when automating Google Cloud access.",
		Docs:    []string{"https://cloud.google.com/docs/authentication", "https://cloud.google.com/docs/authentication/provide-credentials-adc"},
	},
	"config": {
		Context: "Use this group to manage gcgo-local defaults such as project, region, zone, and account. These settings are read by most other commands when explicit flags are omitted.",
		Docs:    []string{"https://cloud.google.com/sdk/docs/properties"},
	},
	"projects": {
		Context: "Use this group to enumerate accessible Google Cloud projects and inspect project metadata before operating on service resources.",
		Docs:    []string{"https://cloud.google.com/resource-manager/docs/creating-managing-projects"},
	},
	"compute": {
		Context: "Use this group for Compute Engine instances, VPC networks, subnetworks, firewall rules, reserved addresses, routes, Cloud Routers, and SSH/SCP workflows. Many commands require project plus zone or region context.",
		Docs:    []string{"https://cloud.google.com/compute/docs", "https://cloud.google.com/vpc/docs"},
	},
	"iam": {
		Context: "Use this group for service accounts, keys, project IAM policy bindings, custom roles, and workload identity federation resources.",
		Docs:    []string{"https://cloud.google.com/iam/docs", "https://cloud.google.com/iam/docs/workload-identity-federation"},
	},
	"storage": {
		Context: "Use this group for Cloud Storage buckets and objects. It supports local-to-GCS, GCS-to-local, and GCS-to-GCS copy flows with gs:// URIs.",
		Docs:    []string{"https://cloud.google.com/storage/docs"},
	},
	"container": {
		Context: "Use this group for GKE clusters and node pools, including credential bootstrapping for kubectl access and basic node pool lifecycle operations.",
		Docs:    []string{"https://cloud.google.com/kubernetes-engine/docs"},
	},
	"run": {
		Context: "Use this group for Cloud Run services, jobs, and job executions. Region is usually required unless already configured in gcgo config.",
		Docs:    []string{"https://cloud.google.com/run/docs"},
	},
	"logging": {
		Context: "Use this group to read and tail Cloud Logging entries. Prefer JSON output for machine processing when supported.",
		Docs:    []string{"https://cloud.google.com/logging/docs"},
	},
	"services": {
		Context: "Use this group for Service Usage API enablement and inspection, typically as a prerequisite for other product-specific commands.",
		Docs:    []string{"https://cloud.google.com/service-usage/docs"},
	},
	"deploy": {
		Context: "Use this group for Cloud Deploy delivery pipelines and releases.",
		Docs:    []string{"https://cloud.google.com/deploy/docs"},
	},
	"eventarc": {
		Context: "Use this group for Eventarc trigger lifecycle management.",
		Docs:    []string{"https://cloud.google.com/eventarc/docs"},
	},
	"tasks": {
		Context: "Use this group for Cloud Tasks queues and individual tasks.",
		Docs:    []string{"https://cloud.google.com/tasks/docs"},
	},
	"workflows": {
		Context: "Use this group for Workflows definitions and deployments.",
		Docs:    []string{"https://cloud.google.com/workflows/docs"},
	},
	"redis": {
		Context: "Use this group for Memorystore for Redis instances.",
		Docs:    []string{"https://cloud.google.com/memorystore/docs/redis"},
	},
	"alloydb": {
		Context: "Use this group for AlloyDB clusters and instances.",
		Docs:    []string{"https://cloud.google.com/alloydb/docs"},
	},
	"bigtable": {
		Context: "Use this group for Cloud Bigtable instances, tables, and operations.",
		Docs:    []string{"https://cloud.google.com/bigtable/docs"},
	},
	"firestore": {
		Context: "Use this group for Firestore admin database inspection and data import/export operations.",
		Docs:    []string{"https://cloud.google.com/firestore/docs"},
	},
	"billing": {
		Context: "Use this group for Cloud Billing account and project billing inspection.",
		Docs:    []string{"https://cloud.google.com/billing/docs"},
	},
	"asset": {
		Context: "Use this group for Cloud Asset Inventory searches, exports, and asset feed management.",
		Docs:    []string{"https://cloud.google.com/asset-inventory/docs"},
	},
	"ai": {
		Context: "Use this group for Vertex AI models, endpoints, custom jobs, and long-running operations.",
		Docs:    []string{"https://cloud.google.com/vertex-ai/docs"},
	},
	"kms": {
		Context: "Use this group for Cloud KMS key rings, crypto keys, versions, and encrypt/decrypt workflows.",
		Docs:    []string{"https://cloud.google.com/kms/docs"},
	},
	"secrets": {
		Context: "Use this group for Secret Manager secrets and secret versions.",
		Docs:    []string{"https://cloud.google.com/secret-manager/docs"},
	},
	"sql": {
		Context: "Use this group for Cloud SQL instances, databases, users, backups, operations, import/export, cloning, and replica promotion.",
		Docs:    []string{"https://cloud.google.com/sql/docs"},
	},
	"dataflow": {
		Context: "Use this group for Dataflow jobs, messages, metrics, snapshots, and flex template launches.",
		Docs:    []string{"https://cloud.google.com/dataflow/docs"},
	},
	"pubsub": {
		Context: "Use this group for Pub/Sub topics, subscriptions, and message workflows.",
		Docs:    []string{"https://cloud.google.com/pubsub/docs"},
	},
	"scheduler": {
		Context: "Use this group for Cloud Scheduler jobs.",
		Docs:    []string{"https://cloud.google.com/scheduler/docs"},
	},
	"artifacts": {
		Context: "Use this group for Artifact Registry repositories and images.",
		Docs:    []string{"https://cloud.google.com/artifact-registry/docs"},
	},
	"builds": {
		Context: "Use this group for Cloud Build build execution and inspection.",
		Docs:    []string{"https://cloud.google.com/build/docs"},
	},
	"functions": {
		Context: "Use this group for Cloud Functions lifecycle operations.",
		Docs:    []string{"https://cloud.google.com/functions/docs"},
	},
	"composer": {
		Context: "Use this group for Cloud Composer environments.",
		Docs:    []string{"https://cloud.google.com/composer/docs"},
	},
	"datacatalog": {
		Context: "Use this group for Data Catalog resources.",
		Docs:    []string{"https://cloud.google.com/data-catalog/docs"},
	},
	"dataplex": {
		Context: "Use this group for Dataplex lakes, zones, and assets.",
		Docs:    []string{"https://cloud.google.com/dataplex/docs"},
	},
	"dataproc": {
		Context: "Use this group for Dataproc clusters and jobs.",
		Docs:    []string{"https://cloud.google.com/dataproc/docs"},
	},
	"folders": {
		Context: "Use this group for Cloud Resource Manager folders.",
		Docs:    []string{"https://cloud.google.com/resource-manager/docs/creating-managing-folders"},
	},
	"organizations": {
		Context: "Use this group for Cloud Resource Manager organizations.",
		Docs:    []string{"https://cloud.google.com/resource-manager/docs/creating-managing-organization"},
	},
	"monitoring": {
		Context: "Use this group for Cloud Monitoring resources and policies.",
		Docs:    []string{"https://cloud.google.com/monitoring/docs"},
	},
	"scc": {
		Context: "Use this group for Security Command Center findings and related resources.",
		Docs:    []string{"https://cloud.google.com/security-command-center/docs"},
	},
	"spanner": {
		Context: "Use this group for Cloud Spanner instances and databases.",
		Docs:    []string{"https://cloud.google.com/spanner/docs"},
	},
}

func configureHelp(root *cobra.Command) {
	root.Long = strings.TrimSpace(`gcgo is a Go-native Google Cloud CLI focused on practical daily operations rather than full gcloud parity.

LLM context:
- Prefer exact subcommands with explicit flags instead of relying on implicit defaults.
- When structured output is available, prefer --format json for machine consumption.
- Most service commands inherit --project from the root command, may also require --region or --zone depending on the resource type, and can use --impersonate-service-account for service account impersonation.
- Last updated: ` + helpLastUpdated + `
- Official docs:
  - https://cloud.google.com/sdk/docs
  - https://cloud.google.com/docs`)

	annotateCommandTree(root, nil)
}

func annotateCommandTree(cmd *cobra.Command, inherited *helpMetadata) {
	current := inherited
	if parent := cmd.Parent(); parent != nil && parent.Name() == "gcgo" {
		if md, ok := serviceHelp[cmd.Name()]; ok {
			mdCopy := md
			current = &mdCopy
		}
	}

	if cmd.Name() != "gcgo" {
		cmd.Long = enrichLong(cmd, current)
	}

	for _, sub := range cmd.Commands() {
		annotateCommandTree(sub, current)
	}
}

func enrichLong(cmd *cobra.Command, md *helpMetadata) string {
	base := strings.TrimSpace(cmd.Long)
	if base == "" {
		base = strings.TrimSpace(cmd.Short)
	}

	lines := []string{base, "", "LLM context:"}
	lines = append(lines, fmt.Sprintf("- Command path: %s", cmd.CommandPath()))
	if md != nil && md.Context != "" {
		lines = append(lines, fmt.Sprintf("- Service context: %s", md.Context))
	}
	lines = append(lines,
		"- Use explicit flags for required location or resource scope values instead of assuming ambient state.",
		"- Prefer --format json when the command returns structured data and you need machine-readable output.",
		"- Global root flags inherited here include --project, --format, --impersonate-service-account, and --quiet.",
		fmt.Sprintf("- Last updated: %s", helpLastUpdated),
	)
	if md != nil && len(md.Docs) > 0 {
		lines = append(lines, "- Official docs:")
		for _, doc := range md.Docs {
			lines = append(lines, fmt.Sprintf("  - %s", doc))
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
