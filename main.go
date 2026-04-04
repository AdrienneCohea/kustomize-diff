// kustomize-diff compares kustomize build output between the current working
// tree and a remote git ref, showing what would change across all overlays.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AdrienneCohea/kustomize-diff/internal/diff"
	"github.com/AdrienneCohea/kustomize-diff/internal/git"
	"github.com/AdrienneCohea/kustomize-diff/internal/kustomize"
	"github.com/AdrienneCohea/kustomize-diff/internal/output"
	"github.com/gonvenience/bunt"
	"github.com/homeport/dyff/pkg/dyff"
)

// stringSliceFlag is a repeatable string flag (e.g. -search-path a -search-path b).
type stringSliceFlag []string

func (f *stringSliceFlag) String() string  { return strings.Join(*f, ", ") }
func (f *stringSliceFlag) Set(v string) error { *f = append(*f, v); return nil }

func main() {
	log.SetFlags(0)

	baseRef := flag.String("base-ref", "", "git ref to compare against (default: auto-detect from origin)")
	noFetch := flag.Bool("no-fetch", false, "skip git fetch before comparing")
	forceColor := flag.Bool("force-color", false, "force colored output even when stdout is not a terminal")
	forceTrueColor := flag.Bool("force-truecolor", false, "force 24-bit true color output (implies --force-color)")
	reportFormat := flag.String("report-format", "auto", "output format: auto, terminal, or github-actions")
	var searchPaths stringSliceFlag
	flag.Var(&searchPaths, "search-path", "relative path within the repo to search for overlays (repeatable); hidden directories are not filtered when this flag is set (default: search entire repo)")
	flag.Parse()

	if len(searchPaths) == 0 {
		// GitHub Actions passes Docker action inputs as INPUT_<name> preserving
		// hyphens, so this is INPUT_SEARCH-PATH, not INPUT_SEARCH_PATH.
		if v := os.Getenv("INPUT_SEARCH-PATH"); v != "" {
			for _, p := range strings.Split(v, "\n") {
				if p = strings.TrimSpace(p); p != "" {
					searchPaths = append(searchPaths, p)
				}
			}
		}
	}

	if *forceTrueColor {
		bunt.SetColorSettings(bunt.ON, bunt.ON)
	} else if *forceColor {
		bunt.SetColorSettings(bunt.ON, bunt.AUTO)
	}

	repoRoot := "."
	if flag.NArg() > 0 {
		repoRoot = flag.Arg(0)
	}

	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		log.Fatalf("resolving repo root %q: %v", repoRoot, err)
	}

	ctx := context.Background()

	if !*noFetch {
		if err := git.FetchOrigin(ctx, absRoot); err != nil {
			fmt.Fprintf(os.Stderr, "warning: git fetch origin: %v\n", err)
		}
	}

	ref := *baseRef
	if ref == "" {
		ref, err = git.DefaultRef(ctx, absRoot)
		if err != nil {
			log.Fatalf("detecting default branch: %v", err)
		}
	}

	tmpDir, err := os.MkdirTemp("", "kustomize-diff-*")
	if err != nil {
		log.Fatalf("creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	noBaseline := false
	if err := git.ExtractRef(ctx, absRoot, ref, tmpDir); err != nil {
		if *baseRef != "" {
			log.Fatalf("base ref %q not found: %v", ref, err)
		}
		noBaseline = true
		fmt.Fprintf(os.Stderr, "warning: could not extract %s, treating as empty baseline: %v\n", ref, err)
	}

	workingOverlays, err := kustomize.FindOverlays(absRoot, searchPaths)
	if err != nil {
		log.Fatalf("finding overlays in working tree: %v", err)
	}

	var baselineOverlays []string
	if !noBaseline {
		baselineOverlays, err = kustomize.FindOverlays(tmpDir, searchPaths)
		if err != nil {
			log.Fatalf("finding overlays in baseline: %v", err)
		}
	}

	allOverlays := union(workingOverlays, baselineOverlays)
	sort.Strings(allOverlays)

	var mode output.Mode
	switch *reportFormat {
	case "auto":
		mode = output.Detect()
	case "terminal":
		mode = output.ModeTerminal
	case "github-actions":
		mode = output.ModeGitHubActions
	default:
		log.Fatalf("unknown --report-format %q: must be auto, terminal, or github-actions", *reportFormat)
	}
	results := make([]output.OverlayResult, 0, len(allOverlays))

	for _, relPath := range allOverlays {
		result, report := compareOverlay(ctx, relPath, absRoot, tmpDir, noBaseline)
		results = append(results, result)
		if result.Status != output.StatusUnchanged {
			if err := output.WriteOverlay(os.Stdout, result, report, mode); err != nil {
				fmt.Fprintf(os.Stderr, "error rendering %s: %v\n", relPath, err)
			}
		}
	}

	output.WriteSummary(os.Stdout, results, mode)
}

// compareOverlay builds both sides of a single overlay and diffs them.
// On error, the returned Report is a zero value and result.Status is StatusError.
func compareOverlay(ctx context.Context, relPath, absRoot, tmpDir string, noBaseline bool) (output.OverlayResult, dyff.Report) {
	baselineYAML, err := buildIfExists(ctx, filepath.Join(tmpDir, relPath), noBaseline)
	if err != nil {
		return output.OverlayResult{
			Path:   relPath,
			Status: output.StatusError,
			Err:    fmt.Errorf("kustomize build (baseline) %s: %w", relPath, err),
		}, dyff.Report{}
	}

	workingYAML, err := buildIfExists(ctx, filepath.Join(absRoot, relPath), false)
	if err != nil {
		return output.OverlayResult{
			Path:   relPath,
			Status: output.StatusError,
			Err:    fmt.Errorf("kustomize build (working) %s: %w", relPath, err),
		}, dyff.Report{}
	}

	report, err := diff.Compare(baselineYAML, workingYAML)
	if err != nil {
		return output.OverlayResult{
			Path:   relPath,
			Status: output.StatusError,
			Err:    fmt.Errorf("comparing %s: %w", relPath, err),
		}, dyff.Report{}
	}

	status := output.StatusUnchanged
	switch {
	case len(baselineYAML) == 0 && len(workingYAML) > 0:
		status = output.StatusAdded
	case len(baselineYAML) > 0 && len(workingYAML) == 0:
		status = output.StatusRemoved
	case len(report.Diffs) > 0:
		status = output.StatusChanged
	}

	return output.OverlayResult{
		Path:     relPath,
		Status:   status,
		NumDiffs: len(report.Diffs),
	}, report
}

// buildIfExists runs kustomize build on dir if it exists and skip is false.
// Returns nil (not an error) when the directory doesn't exist or skip is true.
func buildIfExists(ctx context.Context, dir string, skip bool) ([]byte, error) {
	if skip {
		return nil, nil
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	return kustomize.Build(ctx, dir)
}

func union(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for _, s := range a {
		seen[s] = struct{}{}
	}
	for _, s := range b {
		seen[s] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for s := range seen {
		result = append(result, s)
	}
	return result
}
