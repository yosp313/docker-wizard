#!/usr/bin/env sh
set -euf

REPO_SLUG=${REPO_SLUG:-"yosp313/docker-wizard"}
VERSION=${VERSION:-""}
REF=${REF:-""}
INSTALL_ROOT=${INSTALL_ROOT:-"$HOME/.docker-wizard"}
LINK_DIR=${LINK_DIR:-"$HOME/.local/bin"}
ALLOW_MAIN_FALLBACK=${ALLOW_MAIN_FALLBACK:-""}

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
  normalized_version=${VERSION#v}
  normalized_version=${normalized_version#V}
  REF="v$normalized_version"
fi

if [ -z "$REF" ]; then
  printf '%s\n' "Resolving latest release from $REPO_SLUG"
  if latest_json=$(curl -fsSL -H "Accept: application/vnd.github+json" "https://api.github.com/repos/$REPO_SLUG/releases/latest"); then
    REF=$(printf '%s' "$latest_json" | tr -d '\n' | sed -n 's/.*"tag_name":"\([^"]*\)".*/\1/p')
  fi
  if [ -z "$REF" ]; then
    if [ "$ALLOW_MAIN_FALLBACK" = "1" ]; then
      printf '%s\n' "No releases found, falling back to main because ALLOW_MAIN_FALLBACK=1" >&2
      REF="main"
    else
      printf '%s\n' "Unable to resolve latest release from $REPO_SLUG." >&2
      printf '%s\n' "Publish a release first, or set REF=<tag>." >&2
      printf '%s\n' "For unreleased installs, set REF=main (or ALLOW_MAIN_FALLBACK=1)." >&2
      exit 1
    fi
  fi
fi

case "$REF" in
  main|master)
    ARCHIVE_URL="https://github.com/$REPO_SLUG/archive/refs/heads/$REF.tar.gz"
    ;;
  v*)
    ARCHIVE_URL="https://github.com/$REPO_SLUG/archive/refs/tags/$REF.tar.gz"
    ;;
  V*)
    REF="v${REF#V}"
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
top_dir=$(tar -tzf "$tmpdir/src.tar.gz" | sed -n '1p' | cut -d/ -f1)
if [ -z "$top_dir" ]; then
  printf '%s\n' "failed to read archive" >&2
  exit 1
fi

tar -xzf "$tmpdir/src.tar.gz" -C "$tmpdir"
src_dir="$tmpdir/$top_dir"
if [ ! -d "$src_dir" ]; then
  printf '%s\n' "failed to extract source" >&2
  exit 1
fi

BIN_DIR="$INSTALL_ROOT/bin"
CONFIG_DIR="$BIN_DIR/config"
mkdir -p "$BIN_DIR" "$CONFIG_DIR" "$LINK_DIR"

printf '%s\n' "Building docker-wizard"
build_flags=""
if [ -n "$REF" ]; then
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
