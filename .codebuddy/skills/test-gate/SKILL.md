---
name: test-gate
description: Full test suite runner. Runs go build, go vet, go test ./... before any change is pushed. Used to ensure no breaking changes. Used proactively before commits and pushes.
allowed-tools: Read, Bash
---

# Test Gate

You are responsible for ensuring code changes do not break existing tests.

## Workflow

Run these checks in order. If any fails, stop and report the error.

1. **Build check**
   ```bash
   go build ./...
   ```

2. **Vet check**
   ```bash
   go vet ./...
   ```

3. **Full test suite** (always run, not just short mode)
   ```bash
   go test ./... -timeout 120s
   ```

4. **Go mod tidy check** (if go.mod or go.sum was modified)
   ```bash
   cp go.mod /tmp/go.mod.bak && cp go.sum /tmp/go.sum.bak
   go mod tidy
   diff go.mod /tmp/go.mod.bak && diff go.sum /tmp/go.sum.bak
   ```

5. **Coverage check** (informational, not blocking)
   ```bash
   go test ./... -coverprofile=/tmp/coverage.out -timeout 120s
   go tool cover -func=/tmp/coverage.out | tail -1
   ```

## Rules

- All checks must pass before any commit or push
- If tests fail, fix the issue before proceeding
- Never use --no-verify to skip checks
- Never skip tests for "quick fixes" — always run the full suite
- Report coverage changes if coverage drops significantly
