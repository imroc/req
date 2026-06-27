# Loop Engineering Design

This project uses Loop Engineering for automated maintenance. Core idea: shift from "prompting AI step by step" to "designing self-running loops". Humans become rule-makers, AI systems run autonomously to achieve goals.

## Four-Layer Architecture

| Layer | Purpose | Implementation |
|-------|---------|---------------|
| Prompt | How to ask | Each agent's system prompt |
| Context | What AI sees | `CODEBUDDY.md`, `STATE.md` |
| Harness | AI work environment | skills (project knowledge), allowed-tools (permission constraints) |
| Loop | What to do after each step | GitHub Actions triggers, agent five-phase iteration |

## Implemented Loops

### 1. Dependency Upgrade Loop (closed-loop)
- **Trigger**: Weekly Monday 03:00 UTC, or manual
- **Goal**: Safely upgrade all upgradable Go dependencies (non-modified-code only)
- **Five phases**: Discover (`go list -u`) → Risk-classify → Upgrade → Test → Rollback-retry
- **Stop conditions**: All deps processed / 3 consecutive test failures / 5 iterations max
- **Files**:
  - skill: `skills/dependency-upgrade/SKILL.md`
  - agent: `agents/dependency-upgrader.md`
  - workflow: `.github/workflows/dependency-upgrade-loop.yml`
- **Separation**: dependency-upgrader executes → code-reviewer reviews → PR only after approval

### 2. CI Fix Loop (closed-loop)
- **Trigger**: On CI workflow failure, or manual
- **Goal**: Auto-fix fixable CI failures
- **Five phases**: Discover (`gh run list`) → Classify (ci-triage) → Fix (ci-fix) → Local verify → Retry
- **Stop conditions**: All fixable failures resolved / 3 fix attempts / No failed CI runs
- **Files**:
  - skills: `skills/ci-triage/SKILL.md`, `skills/ci-fix/SKILL.md`
  - agents: `agents/ci-fixer.md`, `agents/code-reviewer.md`
  - workflow: `.github/workflows/ci-fix-loop.yml`
- **Separation**: ci-fixer fixes → code-reviewer reviews → PR only after approval

### 3. Issue Triage Loop (closed-loop)
- **Trigger**: On new issue, or daily 04:00 UTC batch
- **Goal**: Classify, label, and respond to issues; flag quic-go/modified-code issues for caution
- **Five phases**: Discover (untriaged issues) → Classify → Apply labels → Post response → Verify
- **Stop conditions**: All recent issues triaged / 10 issues processed / Needs human judgment
- **Files**:
  - skill: `skills/issue-triage/SKILL.md`
  - agent: `agents/issue-triager.md`
  - workflow: `.github/workflows/issue-triage-loop.yml`
- **Special**: quic-go issues get `quic-go` label and modified-code caveat note

### 4. PR Review Loop (closed-loop)
- **Trigger**: On new PR or PR update
- **Goal**: Review PRs with special caution for modified stdlib, quic-go, and HTTP/2 code
- **Five phases**: Discover (open PRs) → Fetch diff → Review checklist → Post review → Verify
- **Stop conditions**: All open PRs reviewed / 5 PRs reviewed / Needs human judgment
- **Files**:
  - skill: `skills/pr-review/SKILL.md`
  - agent: `agents/pr-reviewer.md`
  - workflow: `.github/workflows/pr-review-loop.yml`
- **Special**: PRs touching `internal/http3/` never auto-approved; quic-go version compatibility required

### 5. Upstream Sync Tracking Loop (semi-closed-loop)
- **Trigger**: Weekly Monday 02:00 UTC
- **Goal**: Track Go stdlib net/http, golang.org/x/net/http2, and quic-go upstream changes; generate sync reports
- **Five phases**: Discover (check baselines) → Fetch upstream changes → Identify affected files → Generate report → Open issue if needed
- **Stop conditions**: Report generated for all 3 sources / Critical change found
- **Files**:
  - skill: `skills/upstream-sync/SKILL.md`
  - agent: `agents/upstream-tracker.md`
  - workflow: `.github/workflows/upstream-sync-loop.yml`
- **Note**: This loop only tracks and reports. Sync of modified code is manual human work — cannot `go get -u` modified code.

## State Management

`STATE.md` (project root) is the shared state file for all loops. This is the core Loop Engineering principle — **state is external, not in model context**.

## Cost Control

- Each loop has max iteration limits (dep upgrade 5, CI fix 3, issue triage 10, PR review 5)
- Uses `model: inherit` to reuse main session model
- Loops run only when needed (event-triggered or scheduled), not always-on
- GitHub Actions provides execution environment, no extra service cost

## Autonomy Boundary

- **Loops can make daily maintenance decisions autonomously**: upgrade deps, fix CI, triage issues, review PRs — no step-by-step human confirmation needed
- **No tagging or releasing**: `git tag`, `gh release create` are always human decisions. Automation stops at PR creation.
- code-reviewer and pr-reviewer agents block any PR that includes release operations
- All commit messages in English

## How to Add a New Loop

1. Create a skill directory and `SKILL.md` under `skills/` (encode project knowledge)
2. Create an agent under `agents/` (define responsibilities, stop conditions, tool permissions)
3. Create a trigger workflow under `.github/workflows/` (scheduled or event-triggered)
4. Add a section to `STATE.md`
5. Reuse `code-reviewer` agent for implementation/review separation
6. Agents involving GitHub operations must declare `gh-cli` in `skills` field
7. Agent safety constraints must include no-tag/no-release rule
8. All skills must be project-level (under `.codebuddy/skills/`), never global
9. Set explicit max iteration limits (turn limits) to prevent infinite loops
10. All commit messages in English

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

| Loop | Trigger | Value |
|------|---------|-------|
| Test coverage improvement | Scheduled | Find low-coverage files, add tests |
| CHANGELOG generation | Tag/scheduled | Summarize changes, generate release notes |
| Security scan | Scheduled | Check known vulnerable dependencies |
