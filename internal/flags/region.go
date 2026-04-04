package flags

import (
	"strings"

	"github.com/spf13/cobra"
)

// CommonRegions is a curated list of GCP regions, bundled so that shell
// completion and error hints work without a network call.
var CommonRegions = []string{
	"us-central1",
	"us-east1",
	"us-east4",
	"us-east5",
	"us-south1",
	"us-west1",
	"us-west2",
	"us-west3",
	"us-west4",
	"europe-west1",
	"europe-west2",
	"europe-west3",
	"europe-west4",
	"europe-west6",
	"europe-central2",
	"europe-north1",
	"asia-east1",
	"asia-east2",
	"asia-northeast1",
	"asia-northeast2",
	"asia-northeast3",
	"asia-south1",
	"asia-south2",
	"asia-southeast1",
	"asia-southeast2",
	"australia-southeast1",
	"australia-southeast2",
	"northamerica-northeast1",
	"northamerica-northeast2",
	"southamerica-east1",
	"southamerica-west1",
	"me-west1",
	"me-central1",
	"africa-south1",
}

// AddRegionFlag adds --region to cmd and registers shell completion backed
// by the hardcoded CommonRegions list.
func AddRegionFlag(cmd *cobra.Command) {
	cmd.Flags().String("region", "", "Region (falls back to config)")
	_ = cmd.RegisterFlagCompletionFunc("region", func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var matches []string
		for _, r := range CommonRegions {
			if strings.HasPrefix(r, toComplete) {
				matches = append(matches, r)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	})
}
