#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
LAUNCHER="$SCRIPT_DIR/memforge-mcp-launcher.js"

if [ -x /opt/homebrew/bin/node ]; then
  exec /opt/homebrew/bin/node "$LAUNCHER" "$@"
fi

if [ -x /usr/local/bin/node ]; then
  exec /usr/local/bin/node "$LAUNCHER" "$@"
fi

if command -v node >/dev/null 2>&1; then
  exec node "$LAUNCHER" "$@"
fi

echo "memforge MCP launcher: node runtime not found; install Node.js or add node to PATH" >&2
exit 1
