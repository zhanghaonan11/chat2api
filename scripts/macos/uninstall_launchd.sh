#!/bin/bash
set -euo pipefail

LABEL="com.shan.chat2api"
PLIST="$HOME/Library/LaunchAgents/$LABEL.plist"

launchctl bootout "gui/$(id -u)" "$PLIST" >/dev/null 2>&1 || true
rm -f "$PLIST"

if launchctl print "gui/$(id -u)/$LABEL" >/dev/null 2>&1; then
  echo "$LABEL is still registered"
  exit 1
fi

echo "$LABEL uninstalled"
