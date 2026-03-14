// Package output handles rendering diff results to a terminal or GitHub Actions.
package output

import (
	"fmt"
	"io"
	"os"

	"github.com/homeport/dyff/pkg/dyff"
)

// Mode controls how results are formatted.
type Mode int

const (
	// ModeTerminal renders plain colored output suitable for a terminal.
	ModeTerminal Mode = iota
	// ModeGitHubActions wraps each overlay in ::group:: blocks and writes a
	// markdown table to $GITHUB_STEP_SUMMARY.
	ModeGitHubActions
)

// Detect returns the appropriate Mode for the current environment.
func Detect() Mode {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return ModeGitHubActions
	}
	return ModeTerminal
}

// OverlayStatus describes the outcome of comparing one overlay.
type OverlayStatus string

const (
	StatusChanged   OverlayStatus = "changed"
	StatusAdded     OverlayStatus = "added"
	StatusRemoved   OverlayStatus = "removed"
	StatusUnchanged OverlayStatus = "unchanged"
	StatusError     OverlayStatus = "error"
)

// OverlayResult holds the comparison outcome for one overlay path.
type OverlayResult struct {
	Path     string
	Status   OverlayStatus
	NumDiffs int
	Err      error
}

// WriteOverlay renders the dyff report for a single overlay to w.
func WriteOverlay(w io.Writer, result OverlayResult, report dyff.Report, mode Mode) error {
	switch mode {
	case ModeGitHubActions:
		fmt.Fprintf(w, "::group::%s [%s]\n", result.Path, result.Status)
		defer fmt.Fprintln(w, "::endgroup::")
	case ModeTerminal:
		fmt.Fprintf(w, "\n=== %s [%s] ===\n", result.Path, result.Status)
	}

	if result.Status == StatusError {
		fmt.Fprintf(w, "error: %v\n", result.Err)
		return nil
	}

	hr := &dyff.HumanReport{
		Report:     report,
		OmitHeader: true,
	}
	return hr.WriteReport(w)
}

// WriteSummary prints a summary of all overlay results to w and, when running
// in GitHub Actions, appends a markdown table to $GITHUB_STEP_SUMMARY.
func WriteSummary(w io.Writer, results []OverlayResult, mode Mode) {
	var changed, added, removed, unchanged, errored int
	for _, r := range results {
		switch r.Status {
		case StatusChanged:
			changed++
		case StatusAdded:
			added++
		case StatusRemoved:
			removed++
		case StatusUnchanged:
			unchanged++
		case StatusError:
			errored++
		}
	}

	fmt.Fprintf(w, "\nSummary: %d changed, %d added, %d removed, %d unchanged",
		changed, added, removed, unchanged)
	if errored > 0 {
		fmt.Fprintf(w, ", %d error(s)", errored)
	}
	fmt.Fprintln(w)

	for _, r := range results {
		if r.Status == StatusError {
			fmt.Fprintf(w, "  error in %s: %v\n", r.Path, r.Err)
		}
	}

	if mode == ModeGitHubActions {
		if path := os.Getenv("GITHUB_STEP_SUMMARY"); path != "" {
			if err := writeStepSummary(path, results); err != nil {
				fmt.Fprintf(os.Stderr, "warning: writing step summary: %v\n", err)
			}
		}
	}
}

func writeStepSummary(path string, results []OverlayResult) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "## Kustomize Diff")
	fmt.Fprintln(f, "| Overlay | Status | Diffs |")
	fmt.Fprintln(f, "|---------|--------|-------|")
	for _, r := range results {
		if r.Status == StatusUnchanged {
			continue
		}
		if r.Status == StatusError {
			fmt.Fprintf(f, "| `%s` | error | %v |\n", r.Path, r.Err)
		} else {
			fmt.Fprintf(f, "| `%s` | %s | %d |\n", r.Path, r.Status, r.NumDiffs)
		}
	}
	return nil
}
