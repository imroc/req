# Loop Engineering Design

This project uses Loop Engineering for automated maintenance. Core idea: shift from "prompting AI step by step" to "designing self-running loops". Humans become rule-makers, AI systems run autonomously to achieve goals.

## Four-Layer Architecture

| Layer | Purpose | Implementation |
|-------|---------|---------------|
| Prompt | How to ask | Each agent's system prompt |
| Context | What AI sees | `CODEBUDDY.md`, `STATE.md` |
| Harness | AI work environment | skills (project knowledge), allowed-tools (permission constraints) |
| Loop | What to do after each step | System cron + `codebuddy -p` (headless), or GitHub Actions |

## Execution Mode

This project supports two execution modes. **System cron is the primary mode** for environments where the CodeBuddy API is only reachable from an internal network (e.g. TKE private deployment).

### Mode 1: System Cron + Headless (Primary)

Each loop runs as a system cron job that invokes `codebuddy -p` (headless mode). Every run is a fresh, isolated session that reads `STATE.md` for context, executes the loop, and writes back `STATE.md`.

**Advantages:**
- No expiration — runs permanently, no need to re-create
- No need to keep a session open
- Each run is isolated (fresh context, no drift)
- Works with TKE internal API (runs locally)

**Install:**
```bash
cd <repo-root>
.codebuddy/scripts/install-cron.sh
```

**Uninstall:**
```bash
.codebuddy/scripts/install-cron.sh --remove
```

**Manual trigger:**
```bash
.codebuddy/scripts/run-loop.sh <loop-type>
# loop-type: ci-fix | issue-triage | pr-review | dependency-upgrade | upstream-sync
```

**View logs:**
```bash
ls $LOOP_LOG_DIR  # default: /tmp/loop-logs/
cat $LOOP_LOG_DIR/ci-fix-*.log
```

**Schedule (daily, Beijing time, staggered to avoid API rate limits):**

| Time | Loop | Script |
|------|------|--------|
| 01:00 | upstream-sync | `run-loop.sh upstream-sync` |
| 02:00 | dependency-upgrade | `run-loop.sh dependency-upgrade` |
| 03:00 | ci-fix | `run-loop.sh ci-fix` |
| 04:00 | issue-triage | `run-loop.sh issue-triage` |
| 05:00 | pr-review | `run-loop.sh pr-review` |

**Run flow:**
```
system cron (daily, e.g. 03:00)
  → run-loop.sh ci-fix
    → git pull --ff-only origin master      # sync latest
    → codebuddy -p -y "..."                  # headless session
      → read STATE.md                        # get last state
      → execute loop logic (via agent)       # discover → plan → execute → verify
      → write STATE.md                       # persist state
      → git push                             # push state changes
    → 10 min timeout protection
    → log to $LOOP_LOG_DIR (default: /tmp/loop-logs/)
```

### Mode 2: GitHub Actions (Optional, for public API)

For environments with a public CodeBuddy API endpoint. Gated behind `CODEBUDDY_ENABLED` repo variable (default: off).

**Enable:**
```bash
# Set repo variable
gh variable set CODEBUDDY_ENABLED -R imroc/req -b "true"
# Set API key secret
gh secret set CODEBUDDY_API_KEY -R imroc/req
```

**Disable:**
```bash
gh variable set CODEBUDDY_ENABLED -R imroc/req -b "false"
```

Workflows are in `.github/workflows/*-loop.yml`.

### Mode 3: Interactive Session (For Ad-hoc Use)

You can also trigger loops directly in an interactive CodeBuddy session:

```
> Use the ci-fixer agent to check for failed CI runs
> Use the issue-triager agent to triage recent issues
> Use the pr-reviewer agent to review PR #485
> Use the dependency-upgrader agent to upgrade dependencies
> Use the upstream-tracker agent to check upstream changes
```

Slash commands (session-level, expire in 3 days):
- `/start-loops` — create all 5 loops as session cron jobs
- `/stop-loops` — delete all session cron jobs
- `/loop-status` — show loop status and last run results

## Implemented Loops

### 1. Upstream Sync Tracking (daily 01:00)
- **Goal**: Track Go stdlib net/http, golang.org/x/net/http2, and quic-go upstream changes
- **Five phases**: Discover (check baselines) → Fetch upstream changes → Identify affected files → Generate report → Open issue if needed
- **Stop conditions**: Report generated for all 3 sources / Critical change found
- **Files**: `skills/upstream-sync/SKILL.md`, `agents/upstream-tracker.md`
- **Note**: Only tracks and reports. Sync of modified code is manual human work.

### 2. Dependency Upgrade (daily 02:00)
- **Goal**: Safely upgrade all upgradable Go dependencies (non-modified-code only)
- **Five phases**: Discover (`go list -u`) → Risk-classify → Upgrade → Test → Rollback-retry
- **Stop conditions**: All deps processed / 3 consecutive test failures / 5 iterations max
- **Files**: `skills/dependency-upgrade/SKILL.md`, `agents/dependency-upgrader.md`
- **Separation**: dependency-upgrader executes → code-reviewer reviews → PR only after approval

### 3. CI Fix (daily 03:00)
- **Goal**: Auto-fix fixable CI failures
- **Five phases**: Discover (`gh run list`) → Classify (ci-triage) → Fix (ci-fix) → Local verify → Retry
- **Stop conditions**: All fixable failures resolved / 3 fix attempts / No failed CI runs
- **Files**: `skills/ci-triage/SKILL.md`, `skills/ci-fix/SKILL.md`, `agents/ci-fixer.md`, `agents/code-reviewer.md`
- **Separation**: ci-fixer fixes → code-reviewer reviews → PR only after approval

### 4. Issue Triage (daily 04:00)
- **Goal**: Classify, label, and respond to issues; flag quic-go/modified-code issues for caution
- **Five phases**: Discover (untriaged issues) → Classify → Apply labels → Post response → Verify
- **Stop conditions**: All recent issues triaged / 10 issues processed / Needs human judgment
- **Files**: `skills/issue-triage/SKILL.md`, `agents/issue-triager.md`
- **Special**: quic-go issues get `quic-go` label and modified-code caveat note

### 5. PR Review (daily 05:00)
- **Goal**: Review PRs with special caution for modified stdlib, quic-go, and HTTP/2 code
- **Five phases**: Discover (open PRs) → Fetch diff → Review checklist → Post review → Verify
- **Stop conditions**: All open PRs reviewed / 5 PRs reviewed / Needs human judgment
- **Files**: `skills/pr-review/SKILL.md`, `agents/pr-reviewer.md`
- **Special**: PRs touching `internal/http3/` never auto-approved; quic-go version compatibility required

## State Management

`STATE.md` (project root) is the shared state file for all loops. This is the core Loop Engineering principle — **state is external, not in model context**. Each cron run reads `STATE.md` at the start and writes back at the end.

## Cost Control

- Each loop runs once daily (not always-on), 10 min max per run
- Staggered schedule (01:00–05:00) to avoid API rate limits
- Headless mode (`-p`) uses minimal tokens per run
- 30-day log retention (older logs auto-deleted)

## Autonomy Boundary

- **Loops can make daily maintenance decisions autonomously**: upgrade deps, fix CI, triage issues, review PRs — no step-by-step human confirmation needed
- **No tagging or releasing**: `git tag`, `gh release create` are always human decisions. Automation stops at PR creation.
- code-reviewer and pr-reviewer agents block any PR that includes release operations
- All commit messages in English

## How to Add a New Loop

1. Create a skill directory and `SKILL.md` under `.codebuddy/skills/` (encode project knowledge)
2. Create an agent under `.codebuddy/agents/` (define responsibilities, stop conditions, tool permissions)
3. Add a case to `.codebuddy/scripts/run-loop.sh` with the loop's prompt and tools
4. Add a cron entry in `.codebuddy/scripts/install-cron.sh`
5. Add a section to `STATE.md`
6. Reuse `code-reviewer` agent for implementation/review separation
7. Agents involving GitHub operations must declare `gh-cli` in `skills` field
8. Agent safety constraints must include no-tag/no-release rule
9. All skills must be project-level (under `.codebuddy/skills/`), never global
10. Set explicit max iteration limits (turn limits) to prevent infinite loops
11. All commit messages in English

## Inner Loop vs Outer Loop

This project follows the two-layer loop model from Loop Engineering:

- **Inner Loop** = each agent's ReAct-style execution (think → act → observe within one task)
- **Outer Loop** = the orchestration layer that manages task lifecycle (Discover → Plan → Execute → Verify → Iterate)

The code-reviewer agent provides **adversarial verification** — a different agent/perspective reviews the executor's work, avoiding the self-confirmation bias of single-agent self-review.

## Cognitive Surrender Risk

Loop Engineering can accelerate "understanding debt" — the gap between the codebase's real state and the developer's mental model grows as AI makes more changes autonomously. Mitigations:

- Regularly review AI-generated changes (diff review before merge)
- Maintain STATE.md as a shared source of truth
- Treat AI as a collaborator, not authority — question every change
- Key changes (quic-go port, stdlib sync, security fixes) always require human review

## Future Expandable Loops

| Loop | Value |
|------|-------|
| Test coverage improvement | Find low-coverage files, add tests |
| CHANGELOG generation | Summarize changes, generate release notes |
| Security scan | Check known vulnerable dependencies |
