package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// KnownError wraps an error with a suggestion for the user.
type KnownError struct {
	Err     error
	Suggest string
}

func (e *KnownError) Error() string {
	return e.Err.Error()
}

func (e *KnownError) Unwrap() error {
	return e.Err
}

// FormatError writes a user-friendly error message to w.
func FormatError(w io.Writer, err error) {
	var known *KnownError
	if errors.As(err, &known) {
		_, _ = fmt.Fprintf(w, "ERROR: %s\n", known.Err)
		if known.Suggest != "" {
			_, _ = fmt.Fprintf(w, "  hint: %s\n", known.Suggest)
		}
		return
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "could not find default credentials"):
		_, _ = fmt.Fprintf(w, "ERROR: %s\n", msg)
		_, _ = fmt.Fprintln(w, "  hint: run 'gcgo auth login' to authenticate")
	case strings.Contains(msg, "project"):
		_, _ = fmt.Fprintf(w, "ERROR: %s\n", msg)
		_, _ = fmt.Fprintln(w, "  hint: run 'gcgo config set project PROJECT_ID'")
	default:
		_, _ = fmt.Fprintf(w, "ERROR: %s\n", msg)
	}
}
