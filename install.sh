#!/usr/bin/env sh
set -euf

REPO_SLUG=${REPO_SLUG:-"yosp313/docker-wizard"}
VERSION=${VERSION:-""}
REF=${REF:-""}
INSTALL_ROOT=${INSTALL_ROOT:-"$HOME/.docker-wizard"}
LINK_DIR=${LINK_DIR:-"$HOME/.local/bin"}

if [ -z "$REPO_SLUG" ]; then
  printf '%s\n' "REPO_SLUG is required." >&2
  exit 1
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf '%s\n' "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd tar
require_cmd go

if [ -n "$VERSION" ]; then
  REF="V$VERSION"
fi

if [ -z "$REF" ]; then
  printf '%s\n' "Resolving latest release from $REPO_SLUG"
  if latest_json=$(curl -fsSL "https://api.github.com/repos/$REPO_SLUG/releases/latest"); then
    REF=$(printf '%s' "$latest_json" | tr -d '\n' | sed -n 's/.*"tag_name":"\([^"]*\)".*/\1/p')
  fi
  if [ -z "$REF" ]; then
    printf '%s\n' "No releases found, falling back to main" >&2
    REF="main"
  fi
fi

case "$REF" in
  main|master)
    ARCHIVE_URL="https://github.com/$REPO_SLUG/archive/refs/heads/$REF.tar.gz"
    ;;
  V*)
    ARCHIVE_URL="https://github.com/$REPO_SLUG/archive/refs/tags/$REF.tar.gz"
    ;;
  *)
    ARCHIVE_URL="https://github.com/$REPO_SLUG/archive/refs/tags/$REF.tar.gz"
    ;;
esac

tmpdir=$(mktemp -d)
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

printf '%s\n' "Downloading $ARCHIVE_URL"
curl -fsSL "$ARCHIVE_URL" -o "$tmpdir/src.tar.gz"
mkdir -p "$tmpdir/src"
tar -xzf "$tmpdir/src.tar.gz" -C "$tmpdir/src" --strip-components=1

src_dir="$tmpdir/src"
if [ ! -d "$src_dir" ]; then
  printf '%s\n' "failed to extract source" >&2
  exit 1
fi

BIN_DIR="$INSTALL_ROOT/bin"
CONFIG_DIR="$BIN_DIR/config"
mkdir -p "$BIN_DIR" "$CONFIG_DIR" "$LINK_DIR"

printf '%s\n' "Building docker-wizard"
build_flags=""
if [ -n "$VERSION" ]; then
  build_flags="-ldflags=-X=main.version=V$VERSION"
elif [ -n "$REF" ]; then
  build_flags="-ldflags=-X=main.version=$REF"
fi
(cd "$src_dir" && go build $build_flags -o "$BIN_DIR/docker-wizard" .)

if [ ! -f "$src_dir/config/services.json" ]; then
  printf '%s\n' "missing config/services.json in source" >&2
  exit 1
fi

cp "$src_dir/config/services.json" "$CONFIG_DIR/services.json"
ln -sf "$BIN_DIR/docker-wizard" "$LINK_DIR/docker-wizard"

printf '%s\n' "Installed to $BIN_DIR/docker-wizard"
printf '%s\n' "Config at $CONFIG_DIR/services.json"
printf '%s\n' "Make sure $LINK_DIR is in your PATH"
