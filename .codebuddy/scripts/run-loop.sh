#!/usr/bin/env bash
#
# Loop Engineering Runner — called by system cron
# Each invocation starts a fresh codebuddy headless session,
# reads STATE.md for context, executes the loop, writes back STATE.md.
#
# Usage: ./run-loop.sh <loop_type>
#   loop_type: ci-fix | issue-triage | pr-review | dependency-upgrade | upstream-sync | all
#
# When "all" is specified, runs all 5 loops sequentially:
#   upstream-sync -> dependency-upgrade -> ci-fix -> issue-triage -> pr-review
#
# Crontab example (Beijing time, UTC+8):
#   0 1 * * * <repo-root>/.codebuddy/scripts/run-loop.sh all

set -euo pipefail

# cron has a minimal PATH — set full PATH so codebuddy, gh, go, etc. are found
export PATH="/root/.linuxbrew/bin:/root/.linuxbrew/sbin:/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:/root/.bin:/root/.local/bin:/root/.cargo/bin:/root/go/bin:/root/.krew/bin:/root/.fzf/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

# Also load NVM, Cargo, Go env if available
export HOME="${HOME:-/root}"
export GOPATH="${GOPATH:-/root/go}"
[ -f /root/.cargo/env ] && . /root/.cargo/env 2>/dev/null || true

LOOP_TYPE="${1:?Usage: run-loop.sh <ci-fix|issue-triage|pr-review|dependency-upgrade|upstream-sync|all>}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
LOG_DIR="${LOOP_LOG_DIR:-/tmp/loop-logs}"
mkdir -p "$LOG_DIR"

cd "$REPO_DIR"

# "all" mode: run all loops sequentially
if [[ "$LOOP_TYPE" == "all" ]]; then
  echo "=== Running all loops sequentially at $(date) ==="
  for lt in upstream-sync dependency-upgrade ci-fix issue-triage pr-review; do
    echo "--- Starting: $lt ---"
    "$0" "$lt" || echo "WARNING: $lt exited with error, continuing..."
    echo "--- Finished: $lt ---"
    sleep 10  # brief pause between loops to avoid API rate limits
  done
  echo "=== All loops completed at $(date) ==="
  exit 0
fi

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="$LOG_DIR/${LOOP_TYPE}-${TIMESTAMP}.log"

# Pull latest master before running
git pull --ff-only origin master >> "$LOG_FILE" 2>&1 || true

case "$LOOP_TYPE" in
  ci-fix)
    PROMPT='Check for failed CI runs: gh run list --status failure --limit 3 --workflow ci.yml -R imroc/req. If there are failures, for each: use gh run view <id> --log-failed to get logs, analyze the failure, and fix auto-fixable issues. Create a branch fix/ci-<run-id>, make the fix, run go build ./... && go vet ./... && go test ./... to verify. If tests pass, create a PR with gh pr create. If not fixable, note it. Finally, update the "CI Fix Loop" section in STATE.md with the date and result, then commit and push STATE.md to master.'
    ;;
  issue-triage)
    PROMPT='Triage recent issues: run gh issue list -R imroc/req --state open --limit 10 --json number,title,labels,createdAt. For issues without labels, fetch details with gh issue view, classify as bug/enhancement/security/question/performance, and apply labels with gh issue edit --add-label. For quic-go or HTTP/3 related issues, add the quic-go label. Post a brief initial response if appropriate. Finally, update the "Issue Triage Loop" section in STATE.md with the date and result, then commit and push STATE.md to master.'
    ;;
  pr-review)
    PROMPT='Review open PRs: run gh pr list -R imroc/req --state open --limit 5 --json number,title,files. For each PR, fetch diff with gh pr diff and review the changes. Pay extra attention to PRs touching internal/http3/ or internal/http2/ — these are modified upstream code and need careful review. Post review comments with gh pr review. Finally, update the "PR Review Loop" section in STATE.md with the date and result, then commit and push STATE.md to master.'
    ;;
  dependency-upgrade)
    PROMPT='Check for dependency upgrades: run go list -u -m all and look for upgradable modules. Upgrade non-sensitive deps with go get -u ./... && go mod tidy. For sensitive deps (utls, x/net, x/crypto, x/text), upgrade individually and test. Do NOT modify files in internal/http3/ or internal/http2/ — those are modified upstream code. Run go build ./... && go vet ./... && go test ./... after upgrade. If tests pass, create branch chore/upgrade-deps-<date>, commit, and create PR with gh pr create. Finally, update the "Dependency Upgrade Loop" section in STATE.md with the date and result, then commit and push STATE.md to master.'
    ;;
  upstream-sync)
    PROMPT='Check upstream changes: read STATE.md for current baselines (Go stdlib net/http, golang.org/x/net/http2, quic-go). Check https://github.com/quic-go/quic-go/releases for new quic-go releases. Check https://github.com/golang/go/commits/master/src/net/http for stdlib changes. Check https://github.com/golang/net/commits/master/http2 for http2 changes. Generate a sync report summarizing what changed. If breaking changes or security fixes are found, open an issue with gh issue create. Finally, update the "Upstream Sync Tracking Loop" section in STATE.md with new baselines, then commit and push STATE.md to master.'
    ;;
  *)
    echo "Unknown loop type: $LOOP_TYPE" | tee -a "$LOG_FILE"
    echo "Valid types: ci-fix, issue-triage, pr-review, dependency-upgrade, upstream-sync, all"
    exit 1
    ;;
esac

echo "=== Loop: $LOOP_TYPE at $(date) ===" >> "$LOG_FILE"

# Run codebuddy in headless mode (-y bypasses permissions for non-interactive use)
timeout 600 codebuddy -p -y \
  "$PROMPT" >> "$LOG_FILE" 2>&1 || true

EXIT_CODE=$?
echo "=== Exit code: $EXIT_CODE at $(date) ===" >> "$LOG_FILE"

# Keep only last 30 days of logs
find "$LOG_DIR" -name "*.log" -mtime +30 -delete 2>/dev/null || true

exit $EXIT_CODE
