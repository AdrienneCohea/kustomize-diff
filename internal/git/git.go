// Package git provides helpers for interacting with a git repository.
package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// FetchOrigin runs git fetch origin in the given repo root.
func FetchOrigin(ctx context.Context, repoRoot string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "fetch", "origin")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// DefaultRef returns the remote default branch as a local ref (e.g. "origin/main").
// It uses git ls-remote so it works even without a local tracking branch.
func DefaultRef(ctx context.Context, repoRoot string) (string, error) {
	out, err := exec.CommandContext(ctx, "git", "-C", repoRoot, "ls-remote", "--symref", "origin", "HEAD").Output()
	if err != nil {
		// Fall back to origin/HEAD; caller may have already fetched.
		return "origin/HEAD", nil
	}
	return parseSymref(out), nil
}

// parseSymref extracts the default branch from git ls-remote --symref output
// and returns it as an "origin/<branch>" ref. Falls back to "origin/HEAD".
func parseSymref(out []byte) string {
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "ref: refs/heads/") {
			branch := strings.TrimPrefix(strings.SplitN(line, "\t", 2)[0], "ref: refs/heads/")
			return "origin/" + branch
		}
	}
	return "origin/HEAD"
}

// ExtractRef extracts the contents of the given git ref into destDir using
// git archive. The directory will mirror the repo root layout.
func ExtractRef(ctx context.Context, repoRoot, ref, destDir string) error {
	archive := exec.CommandContext(ctx, "git", "-C", repoRoot, "archive", ref)
	archiveOut, err := archive.Output()
	if err != nil {
		return fmt.Errorf("git archive %s: %w", ref, err)
	}

	tar := exec.CommandContext(ctx, "tar", "-x", "-C", destDir)
	tar.Stdin = bytes.NewReader(archiveOut)
	if out, err := tar.CombinedOutput(); err != nil {
		return fmt.Errorf("tar extract: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}
