#!/bin/sh
set -e

git config --global --add safe.directory "${GITHUB_WORKSPACE}"

ARGS="-base-ref origin/${GITHUB_BASE_REF}"

if [ -n "${INPUT_SEARCH_PATH}" ]; then
  while IFS= read -r path; do
    [ -n "$path" ] && ARGS="$ARGS -search-path $path"
  done << EOF
${INPUT_SEARCH_PATH}
EOF
fi

exec kustomize-diff $ARGS
