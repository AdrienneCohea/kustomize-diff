package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/homeport/dyff/pkg/dyff"
)

func TestDetect_Terminal(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "")
	if got := Detect(); got != ModeTerminal {
		t.Errorf("Detect() = %v, want ModeTerminal", got)
	}
}

func TestDetect_GitHubActions(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	if got := Detect(); got != ModeGitHubActions {
		t.Errorf("Detect() = %v, want ModeGitHubActions", got)
	}
}

func TestDetect_GitHubActions_WrongValue(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "1")
	if got := Detect(); got != ModeTerminal {
		t.Errorf("Detect() with GITHUB_ACTIONS=1 = %v, want ModeTerminal", got)
	}
}

func TestWriteSummary_Counts(t *testing.T) {
	tests := []struct {
		name    string
		results []OverlayResult
		want    string
	}{
		{
			name:    "all zeros",
			results: nil,
			want:    "0 changed, 0 added, 0 removed, 0 unchanged",
		},
		{
			name: "mixed statuses",
			results: []OverlayResult{
				{Status: StatusChanged},
				{Status: StatusChanged},
				{Status: StatusAdded},
				{Status: StatusRemoved},
				{Status: StatusUnchanged},
				{Status: StatusUnchanged},
				{Status: StatusUnchanged},
			},
			want: "2 changed, 1 added, 1 removed, 3 unchanged",
		},
		{
			name: "with errors",
			results: []OverlayResult{
				{Status: StatusChanged},
				{Status: StatusError, Path: "overlays/prod", Err: errors.New("build failed")},
			},
			want: "1 changed, 0 added, 0 removed, 0 unchanged, 1 error(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			WriteSummary(&buf, tt.results, ModeTerminal)
			got := buf.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("WriteSummary() output = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

func TestWriteSummary_ErrorsListed(t *testing.T) {
	results := []OverlayResult{
		{Status: StatusError, Path: "overlays/prod", Err: errors.New("build failed")},
	}
	var buf bytes.Buffer
	WriteSummary(&buf, results, ModeTerminal)
	got := buf.String()
	if !strings.Contains(got, "overlays/prod") {
		t.Errorf("WriteSummary() = %q, want it to mention error path", got)
	}
	if !strings.Contains(got, "build failed") {
		t.Errorf("WriteSummary() = %q, want it to mention error message", got)
	}
}

func TestWriteOverlay_Terminal_Header(t *testing.T) {
	result := OverlayResult{Path: "overlays/prod", Status: StatusChanged, NumDiffs: 2}
	var buf bytes.Buffer
	if err := WriteOverlay(&buf, result, dyff.Report{}, ModeTerminal); err != nil {
		t.Fatalf("WriteOverlay() error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "overlays/prod") {
		t.Errorf("WriteOverlay(terminal) = %q, want overlay path in output", got)
	}
	if !strings.Contains(got, "changed") {
		t.Errorf("WriteOverlay(terminal) = %q, want status in output", got)
	}
}

func TestWriteOverlay_GitHubActions_GroupMarkers(t *testing.T) {
	result := OverlayResult{Path: "overlays/prod", Status: StatusChanged, NumDiffs: 1}
	var buf bytes.Buffer
	if err := WriteOverlay(&buf, result, dyff.Report{}, ModeGitHubActions); err != nil {
		t.Fatalf("WriteOverlay() error = %v", err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "::group::overlays/prod") {
		t.Errorf("WriteOverlay(gha) = %q, want ::group:: prefix", got)
	}
	if !strings.Contains(got, "::endgroup::") {
		t.Errorf("WriteOverlay(gha) = %q, want ::endgroup::", got)
	}
}

func TestWriteOverlay_Error_PrintsMessage(t *testing.T) {
	result := OverlayResult{
		Path:   "overlays/staging",
		Status: StatusError,
		Err:    errors.New("kustomize plugin not found"),
	}
	var buf bytes.Buffer
	if err := WriteOverlay(&buf, result, dyff.Report{}, ModeTerminal); err != nil {
		t.Fatalf("WriteOverlay() error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "kustomize plugin not found") {
		t.Errorf("WriteOverlay(error) = %q, want error message in output", got)
	}
}

func TestWriteOverlay_Added_Status(t *testing.T) {
	result := OverlayResult{Path: "overlays/new", Status: StatusAdded}
	var buf bytes.Buffer
	if err := WriteOverlay(&buf, result, dyff.Report{}, ModeTerminal); err != nil {
		t.Fatalf("WriteOverlay() error = %v", err)
	}
	if !strings.Contains(buf.String(), "added") {
		t.Errorf("WriteOverlay(added) = %q, want 'added' in output", buf.String())
	}
}
