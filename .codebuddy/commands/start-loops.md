---
description: Start all loop engineering maintenance loops for the req project
allowed-tools: Read, Bash, Grep, Edit, WebFetch
---

# Start Loop Engineering

Start all maintenance loops for the req project using CodeBuddy's built-in cron.

## Loops to Start

Use the CronCreate tool to create the following scheduled tasks:

### 1. CI Fix Loop (every 15 minutes)
- Schedule: every 15 minutes
- Prompt: "Use the ci-fixer agent to check for failed CI runs with `gh run list --status failure --limit 3 --workflow ci.yml -R imroc/req`. If there are failures, use gh run view to get logs, apply ci-triage skill to classify, fix auto-fixable issues on a branch fix/ci-<run-id>, run go build && go vet && go test, then use code-reviewer agent to review. If approved, create PR with gh pr create. Update STATE.md."

### 2. Issue Triage Loop (every 2 hours)
- Schedule: every 2 hours
- Prompt: "Use the issue-triager agent to triage recent issues. Use `gh issue list -R imroc/req --state open --limit 10` to find untriaged issues. For each, fetch with gh issue view, classify, apply labels with gh issue edit --add-label, post initial response if needed. Flag quic-go related issues with extra caution. Update STATE.md."

### 3. PR Review Loop (every 1 hour)
- Schedule: every 1 hour
- Prompt: "Use the pr-reviewer agent to review open PRs. Use `gh pr list -R imroc/req --state open --limit 5` to find unreviewed PRs. For each, fetch diff with gh pr diff, apply pr-review skill checklist. Post review with gh pr review. Be extra cautious with PRs touching internal/http3/ or internal/http2/. Update STATE.md."

### 4. Dependency Upgrade Loop (daily)
- Schedule: daily at 03:00
- Prompt: "Use the dependency-upgrader agent to check for dependency upgrades. Run `go list -u -m all` to find upgradable deps. Upgrade non-sensitive deps with go get -u, run go test ./... after. For sensitive deps (utls, x/net, x/crypto, x/text), upgrade individually and test. Do NOT touch modified code in internal/http3/ or internal/http2/. Create PR with gh pr create if tests pass. Update STATE.md."

### 5. Upstream Sync Tracking Loop (daily)
- Schedule: daily at 02:00
- Prompt: "Use the upstream-tracker agent to check upstream changes. Check Go stdlib net/http, golang.org/x/net/http2, and quic-go for new releases/commits since last sync. Generate a sync report. If breaking changes found, open an issue with gh issue create. Update STATE.md upstream sync baselines."

## Execution

For each loop above, use CronCreate with the appropriate schedule and prompt.
After creating all loops, list them with CronList to confirm.
