# Loop State

This file is the persistent state for Loop Engineering automated maintenance loops.
All loops read from here and write back after execution. State is external, not in model context.

## Dependency Upgrade Loop

### Run History
(first run pending)

### Pending
- [ ] First run: scan all upgradable dependencies

### Last Run
- Date: N/A
- Result: not run

---

## CI Fix Loop

### Run History
(first run pending)

### Pending Human Action
(none)

### Last Run
- Date: N/A
- Result: not run

---

## Issue Triage Loop

### Run History
(first run pending)

### Last Run
- Date: N/A
- Result: not run

---

## PR Review Loop

### Run History
(first run pending)

### Last Run
- Date: N/A
- Result: not run

---

## Upstream Sync Tracking Loop

### Current Baselines
- Go stdlib net/http: 2024-09-10 (ad6ee2)
- golang.org/x/net/http2: 2024-09-06 (3c333c)
- quic-go: v0.57.1

### Run History
(first run pending)

### Last Run
- Date: N/A
- Result: not run

---

## Issue Analysis Snapshot (manual, 2026-06-27)

### Priority — Needs Immediate Action

| # | Title | Type | Notes |
|---|------|------|------|
| #482 | Adapt to quic-go v0.59.0 (ConnectionTracingID removed) | **modified-code sync·urgent** | quic-go v0.59.0 removed `ConnectionTracingID`/`ConnectionTracingKey`, breaks compilation. PR #485 exists, **review cautiously** — contributor may not understand modified-code implications. Affects `internal/http3/transport.go`, `client.go`, `conn.go`. Cannot merge directly |
| #489 | Security: custom auth header leak on cross-domain redirect | **security·high** | `SetCommonHeader` auth headers (X-API-Key etc.) not stripped on cross-domain redirect. CWE-200, CVSS 7.4. Root cause: `middleware.go:528-540`, `client.go:334` |
| #485 | PR: upgrade quic-go to v0.59.0 | **PR review·caution** | Corresponds to #482. 260+/124- lines. Removes deprecated API, replaces hijacker callbacks with RawClientConn, fixes SupportsDatagrams check. **Must manually verify modified-code compatibility** |

### quic-go Related (Caution Required)

| # | Title | Notes |
|---|------|------|
| #460 | Suggest using x/net quic instead of quic-go | Author replied: switch when stdlib matures. Long-term tracking |
| #457 | HTTP/3 AddConn and RoundTrip race | `internal/http3/transport.go` `t.newClientConn` init race, panic under high concurrency. Has repro stacktrace and fix suggestion |
| #372 | Does http3 support SetTLSFingerprint? | Feature question, involves HTTP/3 + TLS fingerprint |

### Bug/Performance

| # | Title | Notes |
|---|------|------|
| #495 | dialConn memory overusage | Unlimited connection pool, suggest MaxConnsPerHost default |
| #433 | Large file upload buffers entire file in memory | `middleware.go:122-123` multipart CreatePart triggers full buffering, 670MB file uses 1.7GB |
| #397 | concurrent map read and map write | `middleware.go:537` parseRequestHeader concurrent read/write of client headers map |
| #436 | Retry limit reached but no error returned | Unexpected behavior |
| #419 | Keep-alive long connection SetTimeout incorrectly disconnects | |
| #416 | ParallelDownload does not close output file | |
| #376 | HTTP/2 frequent errors | 9 comments, may involve modified `internal/http2/` |

### Feature Requests

| # | Title | Notes |
|---|------|------|
| #475 | Read retryOption in middleware | |
| #473 | Socks4 proxy support | |
| #459/#454 | utls fingerprint update to 133 / Chrome133, Firefox133/135 | utls version upgrade |
| #431 | jsonrpc2 support | |
| #425 | zstd content encoding support | |
| #406 | Response body size limit | |
| #404 | Generate cURL debug code | |
| #394 | GraphQL request support | |
| #369 | SSE (Server-Sent Events) support | |

### Pending PR Review (non-quic-go)

| PR | Title | Notes |
|----|------|------|
| #491 | fix: retry on GOAWAY errors (HTTP/2 cached conn) | Involves modified HTTP/2, verify |
| #486 | fix: SetCookieJarFactory return http.CookieJar | Corresponds to #415 |
| #478/#477 | JA3 support / ClientHelloSpec setting | TLS fingerprint related |
| #472 | chrome headers accept add application/json | Corresponds to #471, small change |
| #465 | Unmarshal should check error by status-code |
