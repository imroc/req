#!/usr/bin/env bash
#
# Install cron job for Loop Engineering
# Runs all 5 loops sequentially once daily at 01:00 (Beijing time, UTC+8).
# Loops run in order: upstream-sync -> dependency-upgrade -> ci-fix -> issue-triage -> pr-review
#
# Usage: ./install-cron.sh
# To uninstall: ./install-cron.sh --remove

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_SCRIPT="${SCRIPT_DIR}/run-loop.sh"
MARKER="# loop-engineering-req"

# Remove existing loop cron jobs
remove_existing() {
  crontab -l 2>/dev/null | grep -v "$MARKER" | crontab - 2>/dev/null || true
  echo "Removed existing loop cron jobs."
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

# Add cron job — runs all loops sequentially at 01:00 Beijing time
# Cron uses system local timezone — ensure server is in CST/Asia-Shanghai
(
  crontab -l 2>/dev/null || true
  echo "$MARKER"
  echo "0 1 * * * ${RUN_SCRIPT} all $MARKER"
) | crontab -

echo "Installed loop cron job (daily, Beijing time):"
echo "  01:00  all (upstream-sync -> dependency-upgrade -> ci-fix -> issue-triage -> pr-review)"
echo ""
echo "Logs: \$LOOP_LOG_DIR (default: /tmp/loop-logs/)"
echo "Uninstall: ${SCRIPT_DIR}/install-cron.sh --remove"
echo ""
echo "Current crontab:"
crontab -l | grep "$MARKER"
