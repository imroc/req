---
name: upstream-sync
description: Tracks Go stdlib net/http and quic-go upstream changes against req's modified copies. Generates diff reports for manual sync. Used proactively for upstream tracking.
allowed-tools: Read, Bash, Grep, WebFetch
---

# Upstream Sync Tracking Expert

req modifies Go stdlib net/http, golang.org/x/net/http2, and quic-go source code internally. This skill tracks upstream changes and generates reports for manual sync.

## Modified Code Locations

| Upstream | req Location | Sync Commit Convention |
|----------|-------------|----------------------|
| Go stdlib `src/net/http/` | Root files (`transport.go`, `transfer.go`, `textproto_reader.go`, `http.go`, `http_request.go`, `response.go`, etc.) | `merge upstream net/http: <date>(<hash>)` |
| `golang.org/x/net/http2` | `internal/http2/` | `merge upstream http2: <date>(<hash>)` |
| `github.com/quic-go/quic-go` | `internal/http3/` | `port quic-go <version>` |

## Workflow

1. **Check current upstream baselines**
   - Read `go.mod` for current quic-go version
   - Check git log for last sync commits:
     ```bash
     git log --oneline --all | grep -E "merge upstream|port quic-go" | head -5
     ```

2. **Check Go stdlib upstream** (via WebFetch)
   - Fetch Go source `src/net/http/` commit history since last sync
   - URL: `https://github.com/golang/go/commits/master/src/net/http`
   - Identify commits that affect files modified in req

3. **Check quic-go upstream**
   - Fetch latest quic-go releases: `https://github.com/quic-go/quic-go/releases`
   - Compare with current version in go.mod
   - Identify breaking changes (removed APIs, renamed types)

4. **Check golang.org/x/net/http2 upstream**
   - Fetch `https://github.com/golang/net/commits/master/http2`
   - Identify commits since last sync

5. **Generate sync report** covering:
   - Current baseline: Go stdlib hash, http2 hash, quic-go version
   - New upstream changes since last sync (categorized: bug fix, security, feature, breaking)
   - Affected files in req that need manual sync
   - Risk assessment (especially for quic-go breaking changes)
   - Recommended action items

## Rules

- This skill only **tracks and reports** — never modifies modified-code files directly
- Sync of modified code is a manual human process
- Always note breaking changes in quic-go (they affect `internal/http3/`)
- Report Go version requirements (new stdlib changes may require newer Go)
- Write all reports in English
