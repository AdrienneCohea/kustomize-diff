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

case "$(printenv 'INPUT_FORCE-TRUECOLOR' || true)" in
  true|True|TRUE) ARGS="$ARGS -force-truecolor" ;;
esac

case "$(printenv 'INPUT_RED-GREEN' || true)" in
  true|True|TRUE) ARGS="$ARGS -red-green" ;;
esac

case "$(printenv 'INPUT_SET-EXIT-CODE' || true)" in
  true|True|TRUE) ARGS="$ARGS -set-exit-code" ;;
esac

exec kustomize-diff $ARGS
