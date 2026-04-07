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
| `-no-fetch` | `false` | Skip `git fetch origin` before comparing. |
| `-search-path` | (none) | Relative path within the repo to search for overlays. Repeatable; if omitted, the entire repo is searched. |
| `-report-format` | `auto` | Output format: `auto`, `terminal`, or `github-actions`. `auto` detects GitHub Actions via `$GITHUB_ACTIONS`. |
| `-set-exit-code` | `false` | Exit 1 when differences are found, 0 when none. Fatal errors always exit 255. |
| `-force-color` | `false` | Force colored output even when stdout is not a terminal. |
| `-force-truecolor` | `false` | Force 24-bit true color output (implies `-force-color`). |
| `-red-green` | `false` | Use a red/green color palette instead of the default blue/orange/yellow (Okabe-Ito) palette. |

### Color scheme

By default, `kustomize-diff` uses an Okabe-Ito blue/orange/yellow palette, which is distinguishable across the most common forms of color blindness. dyff also renders `+`/`-`/`±` symbols alongside color, so output remains readable regardless of palette choice. Pass `-red-green` to use the traditional red/green scheme.

### Examples

```sh
# Compare working tree against the remote default branch (auto-detected)
kustomize-diff

# Compare against a specific ref
kustomize-diff -base-ref origin/release-1.2

# Skip fetching (useful offline or in CI when fetch already happened)
kustomize-diff -no-fetch

# Search only specific overlay directories
kustomize-diff -search-path overlays -search-path base

# Exit with a non-zero code when diffs are found (useful in CI)
kustomize-diff -set-exit-code

# Run against a repo at a different path
kustomize-diff /path/to/my-repo
```

## GitHub Actions

### Using the public action

The easiest way to use `kustomize-diff` in a PR workflow is via the published GitHub Action:

```yaml
- uses: actions/checkout@v4
- uses: AdrienneCohea/kustomize-diff@v1
```

To target specific overlay directories, use the `search-path` input (newline-separated):

```yaml
- uses: AdrienneCohea/kustomize-diff@v1
  with:
    search-path: |
      overlays
      base
```

The action automatically diffs against the PR's merge target branch and requires no other configuration. Output is formatted using `::group::` blocks for collapsible sections in the Actions log, and a markdown summary table is written to `$GITHUB_STEP_SUMMARY`.

#### Action inputs

| Input | Default | Description |
|-------|---------|-------------|
| `search-path` | (none) | Newline-separated list of relative paths to search for overlays. If omitted, the entire repo is searched. |
| `set-exit-code` | `false` | Set to `"true"` to exit 1 when differences are found. Useful for blocking merges when manifests change unexpectedly. |
| `force-truecolor` | `false` | Set to `"true"` to force 24-bit true color output (implies force-color). |
| `red-green` | `false` | Set to `"true"` to use a red/green color palette instead of the default Okabe-Ito palette. |

#### Full example

```yaml
- uses: actions/checkout@v4
- uses: AdrienneCohea/kustomize-diff@v1
  with:
    search-path: |
      overlays
      base
    set-exit-code: "true"
    force-truecolor: "true"
```

### Running the CLI directly

You can also invoke the binary directly in a workflow step:

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
