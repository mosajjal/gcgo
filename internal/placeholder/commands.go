package placeholder

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const IssueURL = "https://github.com/mosajjal/gcgo/issues/new/choose"

// NewCommand returns a leaf placeholder command for an unimplemented feature.
func NewCommand(use, short, docsURL string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  longText(short, docsURL),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("%s is not built yet. If you need it, raise an issue at %s", cmd.CommandPath(), IssueURL)
		},
	}
}

// NewGroup returns a placeholder command group for an unimplemented feature area.
func NewGroup(use, short, docsURL string, children ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  longText(short, docsURL),
	}
	cmd.AddCommand(children...)
	return cmd
}

func longText(short, docsURL string) string {
	lines := []string{
		short,
		"",
		"Status: This command line exists for compatibility, but the implementation is not built yet.",
		fmt.Sprintf("Need it? Raise an issue here: %s", IssueURL),
	}
	if docsURL != "" {
		lines = append(lines, fmt.Sprintf("Official docs: %s", docsURL))
	}
	return strings.Join(lines, "\n")
}
