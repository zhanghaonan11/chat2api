#!/bin/bash
set -euo pipefail

REPO_DIR="/Users/shan/github/2api/chat2api"
BIN="$REPO_DIR/target/chat2api"

export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:$HOME/go/bin"
export ENV="${ENV:-dev}"

cd "$REPO_DIR"
mkdir -p "$REPO_DIR/logs" "$REPO_DIR/target"

if [ ! -x "$BIN" ]; then
  go build -o "$BIN" ./cmd
fi

exec "$BIN"
