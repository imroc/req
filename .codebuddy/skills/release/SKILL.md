---
name: release
description: Creates a new GitHub release for req. Handles tag creation, release notes generation from commit history, and gh release create. Use whenever the user asks to publish a release, cut a version, or create a tag/release for req.
allowed-tools: Read, Bash, Grep
---

# Release Manager

Creates a new GitHub release for the req project. This involves creating a git tag, generating release notes from commits since the last release, and publishing via `gh`.

## Prerequisites

- The `gh-cli` skill must be loaded for GitHub release commands. If not already loaded, load it first.
- Working tree should be clean (no uncommitted changes to tracked files).
- Current branch should be `master`.

## Workflow

### 1. Determine the version

The user specifies the version (e.g. `v3.58.0`). Confirm it follows the existing `vMAJOR.MINOR.PATCH` convention.

### 2. Find the previous release

```bash
git tag --sort=-v:refname | head -5
```

The most recent tag is the baseline for generating release notes. For example, if releasing `v3.58.0`, the previous tag is `v3.57.0`.

### 3. Collect commits since the previous tag

```bash
git log <previous_tag>..HEAD --oneline --no-merges
```

Review each commit to categorize it. Read full commit messages for context:

```bash
git log <hash> -1 --format='%B'
```

### 4. Categorize and filter commits

Group commits into these categories:

- **New Features** — `feat:` commits that add user-facing functionality
- **Bug Fixes** — `fix:` commits that resolve issues
- **Dependencies** — `chore: upgrade` or dependency-related commits
- **Tests** — `test:` commits (only mention if significant)
- **Internal** — `ci:`, `docs:`, loop engineering infrastructure, STATE.md updates

Filter out internal-only commits that don't affect library users:
- `.codebuddy/` infrastructure (loop engineering, cron scripts)
- STATE.md updates
- Documentation about internal tooling

When a commit references an issue or PR (e.g. `Closes #123`, `Closes PR #456 by @user`), include the reference in the release notes. Credit external contributors.

### 5. Write release notes

Use this template (in English):

```markdown
## New Features

- **Brief description** — details. Closes #123.

## Bug Fixes

- **Brief description** — details. Fixes #456.

## Dependencies

- **Brief description** — details. Addresses #789.

## Tests

- Brief description of significant test additions.
```

Omit empty sections. Keep descriptions concise but informative — users read these to decide whether to upgrade.

### 6. Create the tag

```bash
git tag -a <version> -m "<version> Release"
git push origin <version>
```

### 7. Create the GitHub release

Write the release notes to a temp file, then use `gh release create`:

```bash
gh release create <version> \
  --title "<version> Release" \
  --notes-file /tmp/<version>-release-notes.md
```

### 8. Verify

```bash
gh release view <version>
```

Confirm the release URL is returned and the notes render correctly.

## Rules

- Release notes must be in English
- Always credit external contributors (e.g. "Closes PR #123 by @username")
- Filter out internal infrastructure commits that don't affect library users
- The previous release format can be checked with:
  ```bash
  gh release view <previous_tag> --json body,name,tagName
  ```
- Never create a release without the user explicitly confirming the version number
- If the working tree has uncommitted changes, warn the user before proceeding
