// Package kustomize provides helpers for finding and building kustomize overlays.
package kustomize

import (
	"context"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
)

// FindOverlays returns the relative paths of all directories under root that
// contain a kustomization.yaml or kustomization.yml file.
//
// If searchPaths is empty the entire root is walked and hidden directories
// (names starting with ".") are skipped. If searchPaths is non-empty, only
// those paths (relative to root) are walked and no hidden-directory filtering
// is applied, allowing callers to target hidden directories explicitly.
func FindOverlays(root string, searchPaths []string) ([]string, error) {
	if len(searchPaths) == 0 {
		return walkForOverlays(root, root, true)
	}
	var all []string
	for _, sp := range searchPaths {
		overlays, err := walkForOverlays(root, filepath.Join(root, sp), false)
		if err != nil {
			return nil, err
		}
		all = append(all, overlays...)
	}
	return all, nil
}

func walkForOverlays(root, start string, skipHidden bool) ([]string, error) {
	var overlays []string
	err := filepath.WalkDir(start, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if skipHidden && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		for _, name := range []string{"kustomization.yaml", "kustomization.yml"} {
			if fileExists(filepath.Join(path, name)) {
				rel, relErr := filepath.Rel(root, path)
				if relErr != nil {
					return relErr
				}
				overlays = append(overlays, rel)
				break
			}
		}
		return nil
	})
	return overlays, err
}

func fileExists(path string) bool {
	_, err := filepath.EvalSymlinks(path)
	return err == nil
}

// Build runs kustomize build in dir and returns the rendered YAML output.
func Build(ctx context.Context, dir string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, "kustomize", "build", dir).Output()
	if err != nil {
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = strings.TrimSpace(string(exitErr.Stderr))
		}
		if stderr != "" {
			return nil, fmt.Errorf("%s: %w", stderr, err)
		}
		return nil, err
	}
	return out, nil
}
