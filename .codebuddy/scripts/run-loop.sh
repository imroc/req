#!/usr/bin/env bash
#
# Loop Engineering Runner — called by system cron
# Each invocation starts a fresh codebuddy headless session,
# reads STATE.md for context, executes the loop, writes back STATE.md.
#
# Usage: ./run-loop.sh <loop_type>
#   loop_type: ci-fix | issue-triage | pr-review | dependency-upgrade | upstream-sync
#
# Crontab example (Beijing time, UTC+8):
#   0 3 * * * /data/git/req/.codebuddy/scripts/run-loop.sh ci-fix
#   0 4 * * * /data/git/req/.codebuddy/scripts/run-loop.sh issue-triage
#   0 5 * * * /data/git/req/.codebuddy/scripts/run-loop.sh pr-review
#   0 2 * * * /data/git/req/.codebuddy/scripts/run-loop.sh dependency-upgrade
#   0 1 * * * /data/git/req/.codebuddy/scripts/run-loop.sh upstream-sync

set -euo pipefail

LOOP_TYPE="${1:?Usage: run-loop.sh <ci-fix|issue-triage|pr-review|dependency-upgrade|upstream-sync>}"
REPO_DIR="/data/git/req"
LOG_DIR="/tmp/loop-logs"
mkdir -p "$LOG_DIR"

cd "$REPO_DIR"

TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="$LOG_DIR/${LOOP_TYPE}-${TIMESTAMP}.log"

# Pull latest master before running
git pull --ff-only origin master >> "$LOG_FILE" 2>&1 || true

case "$LOOP_TYPE" in
  ci-fix)
    PROMPT='Use the ci-fixer agent to check for failed CI runs. Run: gh run list --status failure --limit 3 --workflow ci.yml -R imroc/req. For each failure, use gh run view to get logs, apply ci-triage skill to classify. Fix auto-fixable issues on a branch fix/ci-<run-id>-<date>. Run go build ./... && go vet ./... && go test ./... to verify. Use code-reviewer agent to review changes. If approved, create PR with gh pr create. Update STATE.md CI Fix Loop section. Commit and push STATE.md to master.'
    TOOLS="Read,Bash,Grep,Edit"
    ;;
  issue-triage)
    PROMPT='Use the issue-triager agent to triage recent issues. Run: gh issue list -R imroc/req --state open --limit 10 --json number,title,labels,createdAt. For issues without type labels, fetch with gh issue view, classify by type (bug/enhancement/security/question/performance), apply labels with gh issue edit --add-label. Flag quic-go related issues with quic-go label and note modified-code caveat. Post initial response if appropriate. Update STATE.md Issue Triage Loop section. Commit and push STATE.md to master.'
    TOOLS="Read,Bash,Grep,WebFetch"
    ;;
  pr-review)
    PROMPT='Use the pr-reviewer agent to review open PRs. Run: gh pr list -R imroc/req --state open --limit 5 --json number,title,files. For each PR, fetch diff with gh pr diff, apply pr-review skill checklist. Be extra cautious with PRs touching internal/http3/ or internal/http2/ or modified stdlib files. Post review with gh pr review. Update STATE.md PR Review Loop section. Commit and push STATE.md to master.'
    TOOLS="Read,Bash,Grep"
    ;;
  dependency-upgrade)
    PROMPT='Use the dependency-upgrader agent to check for dependency upgrades. Run: go list -u -m all 2>/dev/null | grep "\[". Upgrade non-sensitive deps with go get -u ./... && go mod tidy. For sensitive deps (utls, x/net, x/crypto, x/text), upgrade individually and test. Do NOT touch modified code in internal/http3/ or internal/http2/. Run go build ./... && go vet ./... && go test ./... after upgrade. If tests pass, create branch chore/upgrade-deps-<date>, commit, and create PR with gh pr create. Update STATE.md Dependency Upgrade Loop section. Commit and push STATE.md to master.'
    TOOLS="Read,Bash,Grep,Edit"
    ;;
  upstream-sync)
    PROMPT='Use the upstream-tracker agent to check upstream changes. Read STATE.md for current baselines. Check Go stdlib net/http (https://github.com/golang/go/commits/master/src/net/http), golang.org/x/net/http2 (https://github.com/golang/net/commits/master/http2), and quic-go releases (https://github.com/quic-go/quic-go/releases) for new changes since last sync. Generate a sync report. If breaking changes or security fixes found, open an issue with gh issue create. Update STATE.md Upstream Sync Tracking Loop section with new baselines. Commit and push STATE.md to master.'
    TOOLS="Read,Bash,Grep,WebFetch"
    ;;
  *)
    echo "Unknown loop type: $LOOP_TYPE" | tee -a "$LOG_FILE"
    echo "Valid types: ci-fix, issue-triage, pr-review, dependency-upgrade, upstream-sync"
    exit 1
    ;;
esac

echo "=== Loop: $LOOP_TYPE at $(date) ===" >> "$LOG_FILE"

# Run codebuddy in headless mode
timeout 600 codebuddy -p -y \
  --allowedTools "$TOOLS" \
  "$PROMPT" >> "$LOG_FILE" 2>&1 || true

EXIT_CODE=$?
echo "=== Exit code: $EXIT_CODE at $(date) ===" >> "$LOG_FILE"

# Keep only last 30 days of logs
find "$LOG_DIR" -name "*.log" -mtime +30 -delete 2>/dev/null || true

exit $EXIT_CODE
