#!/usr/bin/env bash
#
# Install cron jobs for Loop Engineering
# All loops run once daily during nighttime (Beijing time, UTC+8),
# staggered to avoid API rate limits.
#
# Schedule (Beijing time):
#   01:00  upstream-sync     (check upstream changes first)
#   02:00  dependency-upgrade
#   03:00  ci-fix            (fix any CI failures)
#   04:00  issue-triage      (triage new issues)
#   05:00  pr-review         (review open PRs)
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

# Add new cron jobs (Beijing time = UTC+8)
# Cron uses system local timezone — ensure server is in CST/Asia-Shanghai
(
  crontab -l 2>/dev/null || true
  echo "$MARKER"
  echo "0 1 * * * ${RUN_SCRIPT} upstream-sync $MARKER"
  echo "0 2 * * * ${RUN_SCRIPT} dependency-upgrade $MARKER"
  echo "0 3 * * * ${RUN_SCRIPT} ci-fix $MARKER"
  echo "0 4 * * * ${RUN_SCRIPT} issue-triage $MARKER"
  echo "0 5 * * * ${RUN_SCRIPT} pr-review $MARKER"
) | crontab -

echo "Installed 5 loop cron jobs (daily, Beijing time):"
echo "  01:00  upstream-sync"
echo "  02:00  dependency-upgrade"
echo "  03:00  ci-fix"
echo "  04:00  issue-triage"
echo "  05:00  pr-review"
echo ""
echo "Logs: \$LOOP_LOG_DIR (default: /tmp/loop-logs/)"
echo "Uninstall: ${SCRIPT_DIR}/install-cron.sh --remove"
echo ""
echo "Current crontab:"
crontab -l | grep "$MARKER"
