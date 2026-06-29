#!/usr/bin/env bash
#
# Start Loop Engineering — call this on container/host startup
# Starts cron daemon (if not running) and loads persistent crontab.
#
# Usage: .codebuddy/scripts/start-loops.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PERSIST_FILE="/root/.local/share/loop-crontab"

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

# 2. Load persistent crontab if it exists
if [ -f "$PERSIST_FILE" ]; then
    echo "Loading crontab from $PERSIST_FILE..."
    crontab "$PERSIST_FILE"
    echo "Crontab loaded."
else
    echo "No persistent crontab found at $PERSIST_FILE"
    echo "Run: ${SCRIPT_DIR}/install-cron.sh to create one."
fi

# 3. Verify
echo ""
echo "Loop Engineering status:"
echo "  Cron daemon: $(pgrep -x cron >/dev/null && echo 'running' || echo 'NOT running')"
echo "  Crontab:"
crontab -l 2>/dev/null | grep "loop-engineering-req" || echo "  (no loop entries found)"
