#!/usr/bin/env bash
#
# Orbit installer.
# Downloads a release artifact from GitHub Releases, extracts the binary,
# and installs it to the chosen prefix.
#
# Usage:
#   ./scripts/install.sh
#   ./scripts/install.sh --prefix ~/.local/bin
#   ./scripts/install.sh --version v0.2.0
#   ./scripts/install.sh --dry-run
#

set -euo pipefail

APP_NAME="orbit"
REPO_OWNER="${REPO_OWNER:-cristiangonsevi}"
REPO_NAME="${REPO_NAME:-orbit}"
GITHUB_API="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases"
GITHUB_DL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download"

INSTALL_PREFIX="/usr/local/bin"
VERSION="latest"
DRY_RUN=false

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

usage() {
  cat <<EOF
Usage: $0 [options]

Options:
  --prefix <path>   Install the binary into <path> (default: /usr/local/bin)
  --version <tag>   Install a specific release tag (default: latest)
  --dry-run         Show the resolved download/install steps without changing the system
  --help, -h        Show this help message

Environment:
  REPO_OWNER        GitHub owner or organization (default: $REPO_OWNER)
  REPO_NAME         GitHub repository name (default: $REPO_NAME)

Examples:
  $0
  $0 --version v0.2.0
  $0 --prefix ~/.local/bin
EOF
}

info() {
  echo -e "${BLUE}[INFO]${NC} $*"
}

ok() {
  echo -e "${GREEN}[OK]${NC}   $*"
}

warn() {
  echo -e "${YELLOW}[WARN]${NC} $*"
}

fail() {
  echo -e "${RED}[FAIL]${NC} $*"
}

detect_platform() {
  local os arch

  case "$(uname -s)" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *)
      fail "Unsupported OS: $(uname -s). Only Linux and macOS are supported."
      exit 1
      ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    armv7l|armv8l) arch="arm64" ;;
    *)
      fail "Unsupported architecture: $(uname -m). Only amd64 and arm64 are supported."
      exit 1
      ;;
  esac

  echo "$os-$arch"
}

have_tool() {
  command -v "$1" &>/dev/null
}

download_file() {
  local url="$1"
  local output_path="$2"

  if have_tool curl; then
    curl -fsSL "$url" -o "$output_path"
  elif have_tool wget; then
    wget -q -O "$output_path" "$url"
  else
    fail "Neither curl nor wget is installed."
    exit 1
  fi
}

resolve_latest_tag() {
  local release_json

  if have_tool curl; then
    release_json=$(curl -fsSL "$GITHUB_API/latest")
  else
    release_json=$(wget -qO- "$GITHUB_API/latest")
  fi

  echo "$release_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1
}

resolve_release_tag() {
  if [[ "$DRY_RUN" == true ]]; then
    if [[ "$VERSION" == "latest" ]]; then
      echo "latest"
    else
      echo "$VERSION"
    fi
    return
  fi

  if [[ "$VERSION" != "latest" ]]; then
    echo "$VERSION"
    return
  fi

  info "Resolving latest release for $REPO_OWNER/$REPO_NAME"
  local latest_tag
  latest_tag="$(resolve_latest_tag || true)"

  if [[ -z "$latest_tag" ]]; then
    fail "Could not determine the latest release tag."
    fail "Check that the repository exists and has at least one release."
    exit 1
  fi

  echo "$latest_tag"
}

echo ""
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║        Orbit Installer               ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prefix)
      INSTALL_PREFIX="$2"
      shift 2
      ;;
    --version)
      VERSION="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      fail "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
done

echo -e "${YELLOW}━━━ Step 1: Detecting platform ─━━━${NC}"
PLATFORM="$(detect_platform)"
OS="${PLATFORM%-*}"
ARCH="${PLATFORM#*-}"
info "Detected platform: $PLATFORM"

echo ""
echo -e "${YELLOW}━━━ Step 2: Checking dependencies ─━━${NC}"

if have_tool curl; then
  ok "Found curl: $(command -v curl)"
elif have_tool wget; then
  ok "Found wget: $(command -v wget)"
else
  fail "Please install curl or wget and try again."
  exit 1
fi

echo ""
echo -e "${YELLOW}━━━ Step 3: Resolving release ─━━━━━━${NC}"
RELEASE_TAG="$(resolve_release_tag)"
info "Release tag: $RELEASE_TAG"

ASSET_BASE_NAME="$APP_NAME-$OS-$ARCH"
ARCHIVE_NAME="$ASSET_BASE_NAME.tar.gz"
ARCHIVE_URL="$GITHUB_DL/$RELEASE_TAG/$ARCHIVE_NAME"
DIRECT_BINARY_URL="$GITHUB_DL/$RELEASE_TAG/$ASSET_BASE_NAME"

WORKDIR="$(mktemp -d)"
ARCHIVE_PATH="$WORKDIR/$ARCHIVE_NAME"
EXTRACT_DIR="$WORKDIR/extract"
mkdir -p "$EXTRACT_DIR"

cleanup() {
  rm -rf "$WORKDIR"
}

trap cleanup EXIT

echo ""
echo -e "${YELLOW}━━━ Step 4: Downloading asset ───────${NC}"
info "Primary asset: $ARCHIVE_URL"

SOURCE_BINARY=""

if [[ "$DRY_RUN" == true ]]; then
  info "[DRY RUN] Would download: $ARCHIVE_URL"
  info "[DRY RUN] Fallback binary: $DIRECT_BINARY_URL"
  info "[DRY RUN] Would install to: $INSTALL_PREFIX/$APP_NAME"
else
  if download_file "$ARCHIVE_URL" "$ARCHIVE_PATH"; then
    if tar -tzf "$ARCHIVE_PATH" >/dev/null 2>&1; then
      tar -xzf "$ARCHIVE_PATH" -C "$EXTRACT_DIR"
      SOURCE_BINARY="$EXTRACT_DIR/$ASSET_BASE_NAME"
      ok "Downloaded archive: $ARCHIVE_NAME"
    else
      warn "The archive download is not valid, falling back to the direct binary."
    fi
  else
    warn "Archive asset not available, falling back to the direct binary."
  fi

  if [[ -z "$SOURCE_BINARY" ]]; then
    DIRECT_PATH="$WORKDIR/$APP_NAME"
    if ! download_file "$DIRECT_BINARY_URL" "$DIRECT_PATH"; then
      fail "Could not download $ARCHIVE_NAME or the fallback binary."
      fail "Check the release page: https://github.com/$REPO_OWNER/$REPO_NAME/releases"
      exit 1
    fi

    SOURCE_BINARY="$DIRECT_PATH"
    ok "Downloaded binary: $ASSET_BASE_NAME"
  fi

  chmod +x "$SOURCE_BINARY"
fi

echo ""
echo -e "${YELLOW}━━━ Step 5: Installing ───────────────${NC}"

if [[ "$DRY_RUN" == true ]]; then
  info "[DRY RUN] Skipping installation step"
else
  if [[ ! -d "$INSTALL_PREFIX" ]]; then
    warn "Directory $INSTALL_PREFIX does not exist. Creating it..."
    if ! mkdir -p "$INSTALL_PREFIX"; then
      fail "Cannot create $INSTALL_PREFIX. Try using sudo or a different prefix."
      exit 1
    fi
  fi

  if cp "$SOURCE_BINARY" "$INSTALL_PREFIX/$APP_NAME"; then
    chmod +x "$INSTALL_PREFIX/$APP_NAME"
    ok "Installed to: $INSTALL_PREFIX/$APP_NAME"
  else
    fail "Failed to install to $INSTALL_PREFIX/$APP_NAME. Try using sudo."
    exit 1
  fi

  case ":$PATH:" in
    *:"$INSTALL_PREFIX":*)
      ok "$INSTALL_PREFIX is in your PATH"
      ;;
    *)
      warn "$INSTALL_PREFIX is not in your PATH. Add it to your shell profile:"
      warn "  export PATH=\"\$PATH:$INSTALL_PREFIX\""
      ;;
  esac
fi

echo ""
echo -e "${GREEN}╔══════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    Orbit installed successfully     ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════╝${NC}"
echo ""
info "Run 'orbit --help' to verify the installation"
info "Installed: $RELEASE_TAG ($PLATFORM)"
info "Source:    https://github.com/$REPO_OWNER/$REPO_NAME"
