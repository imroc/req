#!/usr/bin/env bash
#
# Install cron job for Loop Engineering
# Runs all 5 loops sequentially once daily at 01:00 (Beijing time, UTC+8).
#
# The crontab is written to a persistent path ($HOME/.local/share/loop-crontab)
# so it survives container restarts. rc.local loads it on boot.
#
# Usage: ./install-cron.sh
# To uninstall: ./install-cron.sh --remove

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_SCRIPT="${SCRIPT_DIR}/run-loop.sh"
MARKER="# loop-engineering-req"
PERSIST_DIR="${HOME}/.local/share"
PERSIST_FILE="${PERSIST_DIR}/loop-crontab"

# Remove existing loop cron jobs from current crontab
remove_existing() {
  crontab -l 2>/dev/null | grep -v "$MARKER" | crontab - 2>/dev/null || true
  rm -f "$PERSIST_FILE"
  echo "Removed loop cron jobs."
}

if [[ "${1:-}" == "--remove" ]]; then
  remove_existing
  exit 0
fi

# Check script exists
if [[ ! -x "$RUN_SCRIPT" ]]; then
  echo "Error: $RUN_SCRIPT not found or not executable"
  exit 1
fi

# Check codebuddy is in PATH
if ! command -v codebuddy &>/dev/null; then
  echo "Error: codebuddy not found in PATH"
  exit 1
fi

# Check gh is authenticated
if ! gh auth status &>/dev/null 2>&1; then
  echo "Error: gh not authenticated. Run: gh auth login"
  exit 1
fi

remove_existing

# Build the crontab content: keep existing non-loop entries + add loop entry
EXISTING=$(crontab -l 2>/dev/null | grep -v "$MARKER" || true)
CRONTAB_CONTENT="${EXISTING}
${MARKER}
0 1 * * * ${RUN_SCRIPT} all ${MARKER}"

# Write to persistent path (survives container restart, loaded by rc.local)
mkdir -p "$PERSIST_DIR"
echo "$CRONTAB_CONTENT" > "$PERSIST_FILE"

# Apply immediately
echo "$CRONTAB_CONTENT" | crontab -

echo "Installed loop cron job (daily, Beijing time):"
echo "  01:00  all (upstream-sync -> dependency-upgrade -> ci-fix -> issue-triage -> pr-review)"
echo ""
echo "Persistent crontab: $PERSIST_FILE (auto-loaded by rc.local on container restart)"
echo "Logs: \$LOOP_LOG_DIR (default: /tmp/loop-logs/)"
echo "Uninstall: ${SCRIPT_DIR}/install-cron.sh --remove"
echo ""
echo "Current crontab:"
crontab -l | grep "$MARKER"
