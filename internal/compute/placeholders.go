package compute

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/flags"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// Images commands

func newImagesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Manage custom Compute Engine images",
	}
	cmd.AddCommand(
		newImagesListCommand(cfg, creds),
		newImagesDescribeCommand(cfg, creds),
		newImagesCreateCommand(cfg, creds),
		newImagesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newImagesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List custom images",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			images, err := client.ListImages(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), images)
			}
			headers := []string{"NAME", "FAMILY", "STATUS", "DISK_SIZE_GB"}
			rows := make([][]string, len(images))
			for i, img := range images {
				rows[i] = []string{img.Name, img.Family, img.Status, fmt.Sprintf("%d", img.DiskSizeGb)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newImagesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe IMAGE",
		Short: "Describe a custom image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			img, err := client.GetImage(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), img)
		},
	}
}

func newImagesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateImageRequest

	cmd := &cobra.Command{
		Use:   "create IMAGE",
		Short: "Create a custom image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if req.SourceDisk == "" {
				return fmt.Errorf("--source-disk is required")
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateImage(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created image %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.SourceDisk, "source-disk", "", "Source persistent disk (required)")
	cmd.Flags().StringVar(&req.Family, "family", "", "Image family")
	cmd.Flags().StringVar(&req.Description, "description", "", "Image description")
	return cmd
}

func newImagesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete IMAGE",
		Short: "Delete a custom image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				_, _ = fmt.Fprintf(os.Stderr, "Delete image %q in %s? (y/N): ", args[0], project)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(answer)) != "y" {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteImage(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted image %q.\n", args[0])
			return nil
		},
	}
}

// VPN Tunnels commands

func newVPNTunnelsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn-tunnels",
		Short: "Manage Cloud VPN tunnels",
	}
	cmd.AddCommand(
		newVPNTunnelsListCommand(cfg, creds),
		newVPNTunnelsDescribeCommand(cfg, creds),
		newVPNTunnelsCreateCommand(cfg, creds),
		newVPNTunnelsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newVPNTunnelsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloud VPN tunnels",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			tunnels, err := client.ListVPNTunnels(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), tunnels)
			}
			headers := []string{"NAME", "REGION", "STATUS", "PEER_IP"}
			rows := make([][]string, len(tunnels))
			for i, t := range tunnels {
				rows[i] = []string{t.Name, t.Region, t.Status, t.PeerIP}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	flags.AddRegionFlag(cmd)
	return cmd
}

func newVPNTunnelsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe TUNNEL",
		Short: "Describe a Cloud VPN tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			t, err := client.GetVPNTunnel(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), t)
		},
	}
	flags.AddRegionFlag(cmd)
	return cmd
}

func newVPNTunnelsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateVPNTunnelRequest

	cmd := &cobra.Command{
		Use:   "create TUNNEL",
		Short: "Create a Cloud VPN tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.PeerIP == "" {
				return fmt.Errorf("--peer-ip is required")
			}
			if req.SharedSecret == "" {
				return fmt.Errorf("--shared-secret is required")
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateVPNTunnel(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created VPN tunnel %q in %s/%s.\n", req.Name, project, region)
			return nil
		},
	}
	flags.AddRegionFlag(cmd)
	cmd.Flags().StringVar(&req.PeerIP, "peer-ip", "", "Peer VPN gateway IP address")
	cmd.Flags().StringVar(&req.SharedSecret, "shared-secret", "", "Shared secret for the tunnel")
	cmd.Flags().StringVar(&req.VPNGateway, "vpn-gateway", "", "VPN gateway self-link")
	cmd.Flags().Int32Var(&req.IKEVersion, "ike-version", 2, "IKE version (1 or 2)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Tunnel description")
	return cmd
}

func newVPNTunnelsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete TUNNEL",
		Short: "Delete a Cloud VPN tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteVPNTunnel(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted VPN tunnel %q.\n", args[0])
			return nil
		},
	}
	flags.AddRegionFlag(cmd)
	return cmd
}
