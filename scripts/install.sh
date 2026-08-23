#!/bin/sh
set -eu

BASE_URL="https://releases.codebahn.net/cli"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

main() {
  os="$(detect_os)"
  arch="$(detect_arch)"
  binary="codebahn-${os}-${arch}"

  printf "Fetching latest version... "
  version="$(fetch_version)"
  echo "v${version}"

  tag="v${version}"
  tag_url="${BASE_URL}/${tag}"

  printf "Downloading %s... " "$binary"
  curl -fsSL -o /tmp/codebahn "$tag_url/$binary"
  echo "done"

  printf "Verifying checksum... "
  curl -fsSL -o /tmp/codebahn-checksums.txt "$tag_url/checksums.txt"
  verify_checksum /tmp/codebahn /tmp/codebahn-checksums.txt "$binary"
  echo "ok"

  mkdir -p "$INSTALL_DIR"
  mv /tmp/codebahn "$INSTALL_DIR/codebahn"
  chmod +x "$INSTALL_DIR/codebahn"

  rm -f /tmp/codebahn-checksums.txt

  echo "Installed codebahn v${version} to ${INSTALL_DIR}/codebahn"

  if ! echo ":$PATH:" | grep -q ":${INSTALL_DIR}:"; then
    echo ""
    echo "Warning: ${INSTALL_DIR} is not in your PATH."
    echo "Add it with:  export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
}

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)             echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
}

fetch_version() {
  version_json="$(curl -fsSL "$BASE_URL/latest.json")"

  # Parse without jq dependency
  echo "$version_json" | sed 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/'
}

verify_checksum() {
  file="$1"
  checksums="$2"
  name="$3"

  expected="$(awk -v f="$name" '$2 == f || $2 == "*"f { print $1 }' "$checksums")"
  if [ -z "$expected" ]; then
    echo "No checksum found for $name" >&2
    exit 1
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$file" | cut -d' ' -f1)"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$file" | cut -d' ' -f1)"
  else
    echo "No sha256sum or shasum found; cannot verify checksum" >&2
    exit 1
  fi

  if [ "$actual" != "$expected" ]; then
    echo "Checksum mismatch: expected $expected, got $actual" >&2
    exit 1
  fi
}

main
