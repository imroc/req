#!/usr/bin/env bash
#
# Pre-push test gate — runs full test suite before allowing push.
# Install: ln -s .codebuddy/scripts/pre-push-check.sh .git/hooks/pre-push
# Or add to .codebuddy/scripts/install-cron.sh to install hook.
#
# Exits non-zero if any check fails, blocking the push.

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Pre-push test gate ===${NC}"

# 1. go build
echo -e "${YELLOW}Running go build ./...${NC}"
if ! go build ./... 2>&1; then
	echo -e "${RED}FAIL: go build${NC}"
	exit 1
fi
echo -e "${GREEN}go build: OK${NC}"

# 2. go vet
echo -e "${YELLOW}Running go vet ./...${NC}"
if ! go vet ./... 2>&1; then
	echo -e "${RED}FAIL: go vet${NC}"
	exit 1
fi
echo -e "${GREEN}go vet: OK${NC}"

# 3. go test (short mode for speed, full mode can be run separately)
echo -e "${YELLOW}Running go test ./... -short${NC}"
if ! go test ./... -short -timeout 120s 2>&1; then
	echo -e "${RED}FAIL: go test${NC}"
	exit 1
fi
echo -e "${GREEN}go test: OK${NC}"

# 4. go mod tidy check (only if go.mod changed)
if git diff --cached --name-only HEAD 2>/dev/null | grep -q "go.mod\|go.sum"; then
	echo -e "${YELLOW}Checking go mod tidy...${NC}"
	cp go.mod /tmp/go.mod.bak
	cp go.sum /tmp/go.sum.bak
	go mod tidy
	if ! diff -q go.mod /tmp/go.mod.bak >/dev/null || ! diff -q go.sum /tmp/go.sum.bak >/dev/null; then
		echo -e "${RED}FAIL: go mod tidy produced changes. Run 'go mod tidy' and commit.${NC}"
		mv /tmp/go.mod.bak go.mod
		mv /tmp/go.sum.bak go.sum
		exit 1
	fi
	echo -e "${GREEN}go mod tidy: OK${NC}"
fi

echo -e "${GREEN}=== All checks passed ===${NC}"
exit 0
