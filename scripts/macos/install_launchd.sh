#!/bin/bash
set -euo pipefail

LABEL="com.shan.chat2api"
REPO_DIR="/Users/shan/github/2api/chat2api"
PLIST="$HOME/Library/LaunchAgents/$LABEL.plist"
RUNNER="$REPO_DIR/scripts/macos/run_chat2api.sh"
BIN="$REPO_DIR/target/chat2api"

export PATH="/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:$HOME/go/bin"

cd "$REPO_DIR"
mkdir -p "$HOME/Library/LaunchAgents" "$REPO_DIR/logs" "$REPO_DIR/target"
chmod +x "$RUNNER"
go build -o "$BIN" ./cmd

cat > "$PLIST" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>$LABEL</string>
  <key>ProgramArguments</key>
  <array>
    <string>$RUNNER</string>
  </array>
  <key>WorkingDirectory</key>
  <string>$REPO_DIR</string>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>ProcessType</key>
  <string>Background</string>
  <key>EnvironmentVariables</key>
  <dict>
    <key>ENV</key>
    <string>dev</string>
    <key>PATH</key>
    <string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:$HOME/go/bin</string>
  </dict>
  <key>StandardOutPath</key>
  <string>$REPO_DIR/logs/launchd.stdout.log</string>
  <key>StandardErrorPath</key>
  <string>$REPO_DIR/logs/launchd.stderr.log</string>
</dict>
</plist>
PLIST

plutil -lint "$PLIST"
launchctl bootout "gui/$(id -u)" "$PLIST" >/dev/null 2>&1 || true
launchctl bootstrap "gui/$(id -u)" "$PLIST"
launchctl kickstart -k "gui/$(id -u)/$LABEL"
launchctl print "gui/$(id -u)/$LABEL"
