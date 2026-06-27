#!/usr/bin/env bash
#
# Start Loop Engineering — call this on container/host startup
# Starts cron daemon (if not running) and installs loop crontab.
#
# Usage: .codebuddy/scripts/start-loops.sh
#
# To auto-start on container boot, add to your entrypoint or profile:
#   /data/git/req/.codebuddy/scripts/start-loops.sh
# Or add to ~/.bashrc / /etc/profile.d/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# 1. Ensure cron daemon is running
if ! pgrep -x cron >/dev/null 2>&1; then
    echo "Starting cron daemon..."
    if command -v systemctl &>/dev/null; then
        systemctl start cron 2>/dev/null || systemctl start crond 2>/dev/null || cron -f &
    elif command -v crond &>/dev/null; then
        crond
    else
        echo "Error: cannot start cron daemon" >&2
        exit 1
    fi
    echo "Cron daemon started."
else
    echo "Cron daemon already running."
fi

# 2. Install loop crontab
echo "Installing loop crontab..."
"$SCRIPT_DIR/install-cron.sh"

# 3. Verify
echo ""
echo "Loop Engineering is active."
echo "Cron daemon: $(pgrep -x cron >/dev/null && echo 'running' || echo 'NOT running')"
echo "Crontab:"
crontab -l 2>/dev/null | grep "loop-engineering-req" || echo "  (no loop entries found)"
