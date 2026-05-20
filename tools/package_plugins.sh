#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
DIST=${DIST:-$ROOT/dist}
VERSION=${VERSION:-$(cat "$ROOT/VERSION")}

PLUGINS_DIR="$DIST/plugins"
CLAUDE_PACKAGE="$PLUGINS_DIR/claude-code/memforge"
CODEX_PACKAGE="$PLUGINS_DIR/codex/memforge"

platforms=(
  darwin-arm64
  darwin-amd64
  linux-arm64
  linux-amd64
  windows-arm64
  windows-amd64
)

missing_binary() {
  printf 'missing built binary: %s\n' "$1" >&2
  exit 2
}

copy_runtime() {
  local platform=$1
  local source_name="memforge-$platform"
  local target_name="memforge"

  if [[ $platform == windows-* ]]; then
    source_name="$source_name.exe"
    target_name="memforge.exe"
  fi

  local source="$DIST/bin/$source_name"
  [[ -f $source ]] || missing_binary "$source"

  for package in "$CLAUDE_PACKAGE" "$CODEX_PACKAGE"; do
    local target_dir="$package/bin/$platform"
    mkdir -p "$target_dir"
    cp "$source" "$target_dir/$target_name"
    chmod +x "$target_dir/$target_name" 2>/dev/null || true
  done
}

rm -rf "$PLUGINS_DIR"
mkdir -p "$PLUGINS_DIR/claude-code" "$PLUGINS_DIR/codex"

cp -R "$ROOT/plugins/claude-code/memforge" "$PLUGINS_DIR/claude-code/"
cp -R "$ROOT/plugins/codex/memforge" "$PLUGINS_DIR/codex/"

for platform in "${platforms[@]}"; do
  copy_runtime "$platform"
done

claude_zip="$DIST/plugins/memforge-claude-code-plugin-v$VERSION.zip"
codex_zip="$DIST/plugins/memforge-codex-plugin-v$VERSION.zip"

(
  cd "$PLUGINS_DIR/claude-code"
  zip -qr "$claude_zip" memforge
)

(
  cd "$PLUGINS_DIR/codex"
  zip -qr "$codex_zip" memforge
)

printf '%s\n' "$claude_zip"
printf '%s\n' "$codex_zip"
