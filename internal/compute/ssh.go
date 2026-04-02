package compute

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SSHArgs builds the ssh command arguments for connecting to an instance.
func SSHArgs(user, ip string, extraArgs []string) []string {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}
	target := ip
	if user != "" {
		target = user + "@" + ip
	}
	args = append(args, target)
	args = append(args, extraArgs...)
	return args
}

// SCPArgs builds the scp command arguments.
// src/dst that contain ":" are treated as remote paths and get the IP substituted.
func SCPArgs(user, ip, src, dst string) []string {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
	}

	resolve := func(path string) string {
		if strings.Contains(path, ":") {
			// instance:path -> user@ip:path
			parts := strings.SplitN(path, ":", 2)
			remote := ip
			if user != "" {
				remote = user + "@" + ip
			}
			_ = parts[0] // instance name, already resolved
			return remote + ":" + parts[1]
		}
		return path
	}

	args = append(args, resolve(src), resolve(dst))
	return args
}

// ResolveInstanceIP looks up an instance's external IP. Falls back to internal if no external.
func ResolveInstanceIP(ctx context.Context, client Client, project, zone, name string) (string, error) {
	inst, err := client.GetInstance(ctx, project, zone, name)
	if err != nil {
		return "", err
	}
	if inst.ExternalIP != "" {
		return inst.ExternalIP, nil
	}
	if inst.InternalIP != "" {
		return inst.InternalIP, nil
	}
	return "", fmt.Errorf("instance %q has no IP address", name)
}

// ExecSSH executes the ssh binary with the given args. Does not return on success.
func ExecSSH(args []string) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	cmd := exec.Command(sshPath, args...) //nolint:gosec // args are built by SSHArgs, not raw user input
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ExecSCP executes the scp binary with the given args.
func ExecSCP(args []string) error {
	scpPath, err := exec.LookPath("scp")
	if err != nil {
		return fmt.Errorf("scp not found in PATH: %w", err)
	}

	cmd := exec.Command(scpPath, args...) //nolint:gosec // args are built by SCPArgs, not raw user input
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
