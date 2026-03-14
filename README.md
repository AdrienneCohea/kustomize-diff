# kustomize-diff

`kustomize-diff` compares `kustomize build` output across all overlays in a repository between your working tree and a git ref (by default, the remote default branch). It's useful for reviewing what Kubernetes manifests would actually change before merging a PR.

## How it works

1. Fetches `origin` to get the latest remote state.
2. Checks out the baseline ref into a temp directory.
3. Finds every overlay (directory containing a `kustomization.yaml` or `kustomization.yml`) in both the working tree and the baseline.
4. Runs `kustomize build` on each overlay for both sides.
5. Diffs the rendered YAML using [dyff](https://github.com/homeport/dyff) and prints a human-readable report.

Overlays that are added, removed, or changed are shown; unchanged overlays are silently skipped. A summary line is printed at the end.

## Requirements

- [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/) must be on your `PATH`.
- Git must be on your `PATH`.
- Go 1.23+ (to build from source).

## Installation

```sh
go install github.com/AdrienneCohea/kustomize-diff@latest
```

Or build from source:

```sh
git clone https://github.com/AdrienneCohea/kustomize-diff.git
cd kustomize-diff
go build -o kustomize-diff .
```

## Usage

```
kustomize-diff [flags] [repo-root]
```

Run from within a git repository, or pass the path to the repo root as a positional argument.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-base-ref` | auto-detect | Git ref to compare the working tree against. Defaults to the remote default branch (e.g. `origin/main`). |
| `-no-fetch` | false | Skip `git fetch origin` before comparing. |

### Examples

```sh
# Compare working tree against the remote default branch (auto-detected)
kustomize-diff

# Compare against a specific ref
kustomize-diff -base-ref origin/release-1.2

# Skip fetching (useful offline or in CI when fetch already happened)
kustomize-diff -no-fetch

# Run against a repo at a different path
kustomize-diff /path/to/my-repo
```

## GitHub Actions

When run inside a GitHub Actions workflow (`GITHUB_ACTIONS=true`), output is formatted using `::group::` blocks for collapsible sections, and a markdown summary table is written to `$GITHUB_STEP_SUMMARY`.

Example workflow step:

```yaml
- name: kustomize-diff
  run: kustomize-diff -no-fetch -base-ref origin/${{ github.base_ref }}
```

## License

Copyright (C) 2026 Adrienne Cohea

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version. See [LICENSE](LICENSE) for the full text.
