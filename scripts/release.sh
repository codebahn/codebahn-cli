#!/usr/bin/env bash
set -euo pipefail

TAG="${1:?Usage: $0 v0.1.0}"

if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: tag must match v<major>.<minor>.<patch>" >&2
  exit 1
fi

VERSION="${TAG#v}"
DIST="dist"
BUCKET="codebahn-releases"
ENDPOINT="https://s3.fr-par.scw.cloud"

TARGETS=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
)

rm -rf "$DIST"
mkdir -p "$DIST"

for target in "${TARGETS[@]}"; do
  read -r os arch <<< "$target"
  name="codebahn-${os}-${arch}"
  echo "Building ${name}..."
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build \
    -ldflags "-s -w -X main.version=${TAG}" \
    -o "${DIST}/${name}" \
    ./cmd/codebahn
done

echo "Generating checksums..."
(cd "$DIST" && sha256sum codebahn-* > checksums.txt)

echo "Signing checksums..."
gpg --batch --yes --detach-sign --armor "${DIST}/checksums.txt"

echo "Uploading to ${BUCKET}..."
for file in "${DIST}"/*; do
  aws s3 cp "$file" "s3://${BUCKET}/cli/${TAG}/$(basename "$file")" \
    --endpoint-url "$ENDPOINT" \
    --acl public-read \
    --quiet
done

echo "{\"version\":\"${VERSION}\"}" | aws s3 cp - \
  "s3://${BUCKET}/cli/latest.json" \
  --endpoint-url "$ENDPOINT" \
  --content-type application/json \
  --acl public-read \
  --quiet

echo "Released ${TAG}"
