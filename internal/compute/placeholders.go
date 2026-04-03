package compute

import (
	"github.com/mosajjal/gcgo/internal/placeholder"
	"github.com/spf13/cobra"
)

func newImagesCommand() *cobra.Command {
	const docsURL = "https://cloud.google.com/compute/docs/images"
	return placeholder.NewGroup(
		"images",
		"Manage custom Compute Engine images",
		docsURL,
		placeholder.NewCommand("list", "List custom images", docsURL),
		placeholder.NewCommand("describe", "Describe a custom image", docsURL),
		placeholder.NewCommand("create", "Create a custom image", docsURL),
		placeholder.NewCommand("delete", "Delete a custom image", docsURL),
	)
}

func newVPNTunnelsCommand() *cobra.Command {
	const docsURL = "https://cloud.google.com/network-connectivity/docs/vpn/how-to/creating-static-vpns"
	return placeholder.NewGroup(
		"vpn-tunnels",
		"Manage Cloud VPN tunnels",
		docsURL,
		placeholder.NewCommand("list", "List Cloud VPN tunnels", docsURL),
		placeholder.NewCommand("describe", "Describe a Cloud VPN tunnel", docsURL),
		placeholder.NewCommand("create", "Create a Cloud VPN tunnel", docsURL),
		placeholder.NewCommand("delete", "Delete a Cloud VPN tunnel", docsURL),
	)
}
