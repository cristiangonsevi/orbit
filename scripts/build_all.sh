#!/usr/bin/env bash
#
# Orbit cross-platform builder.
# Builds release binaries and optional tar.gz archives for GitHub Releases.
#
# Usage:
#   ./scripts/build_all.sh
#   GOOS=linux GOARCH=amd64 ./scripts/build_all.sh
#   VERSION=v0.2.0 ./scripts/build_all.sh --clean
#   ./scripts/build_all.sh --output-dir ./dist --clean
#

set -euo pipefail

APP_NAME="orbit"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$PROJECT_ROOT/build"
VERSION="${VERSION:-dev}"
LDFLAGS="${LDFLAGS:--s -w -X github.com/cristiangonsevi/orbit/cmd.Version=${VERSION}}"
CLEAN=false

DEFAULT_TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

usage() {
  cat <<'EOF'
Usage: ./scripts/build_all.sh [options]

Options:
  --output-dir <path>   Write build artifacts to <path> (default: ./build)
  --clean               Remove the output directory before building
  --help, -h            Show this help message

Environment:
  GOOS/GOARCH           Build a single target when both are set
  TARGETS               Comma or space separated targets like linux/amd64,darwin/arm64
  VERSION               Version embedded in the binary via ldflags (default: dev)
  LDFLAGS               Extra linker flags appended to the default release flags
EOF
}

info() {
  echo "[build] $*"
}

fail() {
  echo "[build] ERROR: $*" >&2
  exit 1
}

parse_targets() {
  if [[ -n "${GOOS:-}" || -n "${GOARCH:-}" ]]; then
    if [[ -z "${GOOS:-}" || -z "${GOARCH:-}" ]]; then
      fail "GOOS and GOARCH must be set together when building a single target"
    fi
    TARGET_MATRIX=("$GOOS/$GOARCH")
    return
  fi

  if [[ -n "${TARGETS:-}" ]]; then
    read -r -a TARGET_MATRIX <<< "${TARGETS//,/ }"
    return
  fi

  TARGET_MATRIX=("${DEFAULT_TARGETS[@]}")
}

build_target() {
  local target="$1"
  local os="${target%/*}"
  local arch="${target#*/}"
  local binary_name="$APP_NAME-$os-$arch"
  local archive_name="$binary_name.tar.gz"
  local staging_dir="$OUTPUT_DIR/.staging/$binary_name"
  local staged_binary_name="$binary_name"

  if [[ "$os" == "windows" ]]; then
    binary_name+=".exe"
    archive_name="$binary_name.tar.gz"
    staged_binary_name="$binary_name"
  fi

  info "Building $os/$arch -> $binary_name"

  rm -rf "$staging_dir"
  mkdir -p "$staging_dir"

  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build \
    -trimpath \
    -ldflags="$LDFLAGS" \
    -o "$staging_dir/$staged_binary_name" \
    .

  cp "$staging_dir/$binary_name" "$OUTPUT_DIR/$binary_name"

  tar -czf "$OUTPUT_DIR/$archive_name" -C "$staging_dir" "$binary_name"
  info "Packaged $archive_name"

  rm -rf "$staging_dir"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output-dir)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --clean)
      CLEAN=true
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      fail "Unknown option: $1"
      ;;
  esac
done

parse_targets

cd "$PROJECT_ROOT"

if [[ "$CLEAN" == true ]]; then
  rm -rf "$OUTPUT_DIR"
fi

mkdir -p "$OUTPUT_DIR"

for target in "${TARGET_MATRIX[@]}"; do
  build_target "$target"
done

info "Builds completed in $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
