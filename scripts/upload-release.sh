#!/usr/bin/env bash
set -euo pipefail

TAG="${1:?Usage: $0 v0.1.0}"
VERSION="${TAG#v}"
BUCKET="codebahn-releases"
ENDPOINT="https://s3.fr-par.scw.cloud"

upload() {
  local src="$1" dst="$2"
  aws s3 cp "$src" "s3://${BUCKET}/cli/${TAG}/${dst}" \
    --endpoint-url "$ENDPOINT" \
    --acl public-read \
    --quiet
}

echo "Uploading binaries..."
for dir in dist/codebahn-cli_*; do
  [ -d "$dir" ] || continue
  binary="$dir/codebahn"
  [ -f "$binary" ] || continue

  # Extract os and arch from directory name (codebahn-cli_linux_amd64_v1)
  name=$(basename "$dir" | sed 's/codebahn-cli_//; s/_v[0-9.]*$//')
  os=$(echo "$name" | cut -d_ -f1)
  arch=$(echo "$name" | cut -d_ -f2)
  upload "$binary" "codebahn-${os}-${arch}"
done

echo "Uploading archives..."
for archive in dist/*.tar.gz; do
  [ -f "$archive" ] || continue
  upload "$archive" "$(basename "$archive")"
done

echo "Uploading checksums..."
upload dist/checksums.txt checksums.txt
[ -f dist/checksums.txt.asc ] && upload dist/checksums.txt.asc checksums.txt.asc

echo "Uploading install script..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
aws s3 cp "$SCRIPT_DIR/install.sh" "s3://${BUCKET}/cli/install.sh" \
  --endpoint-url "$ENDPOINT" \
  --content-type text/plain \
  --acl public-read \
  --quiet

echo "Updating latest.json..."
echo "{\"version\":\"${VERSION}\"}" | aws s3 cp - \
  "s3://${BUCKET}/cli/latest.json" \
  --endpoint-url "$ENDPOINT" \
  --content-type application/json \
  --acl public-read \
  --quiet

echo "Released ${TAG}"
