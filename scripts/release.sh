#!/usr/bin/env bash
set -euo pipefail

TAG="${1:?Usage: $0 v0.1.0}"

if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: tag must match v<major>.<minor>.<patch>" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Building with GoReleaser..."
GORELEASER_CURRENT_TAG="$TAG" goreleaser release --clean --skip=publish

echo "Uploading..."
"$SCRIPT_DIR/upload-release.sh" "$TAG"
