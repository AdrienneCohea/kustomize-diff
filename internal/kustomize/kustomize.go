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

// FindOverlays walks root and returns the relative paths of all directories
// that contain a kustomization.yaml or kustomization.yml file. Hidden
// directories (names starting with ".") are skipped.
func FindOverlays(root string) ([]string, error) {
	var overlays []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
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
