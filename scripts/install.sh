#!/usr/bin/env bash
set -euo pipefail

REPO="MagnumGOYB/memforge"
MEMFORGE_VERSION="${MEMFORGE_VERSION:-latest}"
MEMFORGE_INSTALL_DIR="${MEMFORGE_INSTALL_DIR:-$HOME/.local/bin}"

need() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "memforge install: missing required command: $1" >&2
		exit 2
	fi
}

detect_asset() {
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"
	arch="$(uname -m)"

	case "$os-$arch" in
	darwin-arm64) printf 'memforge-darwin-arm64' ;;
	darwin-x86_64 | darwin-amd64) printf 'memforge-darwin-amd64' ;;
	linux-arm64 | linux-aarch64) printf 'memforge-linux-arm64' ;;
	linux-x86_64 | linux-amd64) printf 'memforge-linux-amd64' ;;
	*)
		echo "memforge install: unsupported platform: $os-$arch" >&2
		echo "Download a release asset manually or install with Go." >&2
		exit 2
		;;
	esac
}

latest_asset_url() {
	curl -fsSL "https://api.github.com/repos/MagnumGOYB/memforge/releases/latest" |
		sed -n 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^"]*\/'"$asset"'\)".*/\1/p' |
		head -n 1
}

need curl
need sed
need uname
need install

asset="$(detect_asset)"
if [ "$MEMFORGE_VERSION" = "latest" ]; then
	url="$(latest_asset_url)"
	if [ -z "$url" ]; then
		echo "memforge install: could not resolve latest release asset for $asset" >&2
		exit 3
	fi
	tag="latest"
else
	case "$MEMFORGE_VERSION" in
	v*) tag="$MEMFORGE_VERSION" ;;
	*) tag="v$MEMFORGE_VERSION" ;;
	esac
	url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
fi
tmp_dir="${TMPDIR:-/tmp}/memforge-install.$$"
tmp_bin="$tmp_dir/$asset"

cleanup() {
	rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

mkdir -p "$tmp_dir" "$MEMFORGE_INSTALL_DIR"
echo "Downloading MemForge ${tag} asset ${asset}..." >&2
curl -fsSL "$url" -o "$tmp_bin"
install -m 0755 "$tmp_bin" "$MEMFORGE_INSTALL_DIR/memforge"

echo "Installed memforge to $MEMFORGE_INSTALL_DIR/memforge" >&2
if [ -x "$MEMFORGE_INSTALL_DIR/memforge" ]; then
	"$MEMFORGE_INSTALL_DIR/memforge" --version
fi
