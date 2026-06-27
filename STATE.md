# Loop State

This file is the persistent state for Loop Engineering automated maintenance loops.
All loops read from here and write back after execution. State is external, not in model context.

## Dependency Upgrade Loop

### Run History
(first run pending — cron installed, will run at 02:00 daily)

### Last Run
- Date: N/A
- Result: not run

---

## CI Fix Loop

### Run History
(first run pending — cron installed, will run at 03:00 daily)

### Last Run
- Date: N/A
- Result: not run

---

## Issue Triage Loop

### Run History
- 2026-06-27: First run — triaged 10 open issues, applied 14 labels, created 4 labels (security, performance, quic-go, tls-fingerprint), posted 3 comments (#496, #481, #464)

### Last Run
- Date: 2026-06-27
- Result: success — 10 issues triaged

### Triage Details (2026-06-27)
| # | Title | Type | Extra | Comment |
|---|---|---|---|---|
| #496 | gin dependency conflict (quic-go v0.59.0 API break) | bug | quic-go | Yes — explained modified-code root cause, workaround |
| #495 | dialConn memory overusage | bug | performance | No — already self-answered |
| #494 | client vs request timeout precedence | question | — | No — already answered |
| #484 | OnAfterResponse full-iteration design question | question | — | No — design question for maintainer |
| #481 | TLS fingerprint with Transport only | question | tls-fingerprint | Yes — pointed to SetTLSHandshake |
| #475 | Expose retryOption in middleware | enhancement | — | No |
| #473 | Socks4 proxy support | enhancement | — | No — has working code example |
| #468 | Disable auto header injection | question | — | No — answered by maintainer |
| #464 | ForceHTTP1 + TLS random fingerprint | question | tls-fingerprint | Yes — explained they work together |
| #460 | Replace quic-go with x/net QUIC | enhancement | quic-go | No — active discussion |

### Labels Created This Run
- `security` (b60205) — Security vulnerability or concern
- `performance` (fbca04) — Performance-related issue
- `quic-go` (5319e7) — Involves quic-go, HTTP/3, or internal/http3 modified code
- `tls-fingerprint` (1d76db) — Involves TLS fingerprinting or utls

### Needs Human Attention
- **#496** — Compilation failure: gin v1.12.0 pulls quic-go v0.59.0, but internal/http3/ uses removed `quic.ConnectionTracingID` API. High priority, breaks user builds.
- **#484** — Inconsistent error handling between client-level and request-level afterResponse middleware loops (request-level stops on first error, client-level doesn't). May warrant code change.

---

## PR Review Loop

### Run History
(first run pending — cron installed, will run at 05:00 daily)

### Last Run
- Date: N/A
- Result: not run

---

## Upstream Sync Tracking Loop

### Current Baselines
- Go stdlib net/http: 2024-09-10 (ad6ee2)
- golang.org/x/net/http2: 2024-09-06 (3c333c)
- quic-go: v0.59.0
- utls: v1.8.2

### Run History
(first run pending — cron installed, will run at 01:00 daily)

### Last Run
- Date: N/A
- Result: not run

---

## Manual Actions Log (2026-06-27)

### Resolved
- #482: quic-go v0.59.0 ported (commit 8b018a5, 29f0e6d)
- #489: SensitiveHeadersRedirectPolicy added (commit b144240)
- #471/#472: Chrome accept header add application/json (commit 1e34ab9)
- #415/#486: SetCookieJarFactory returns http.CookieJar (commit d702aad)
- #491: GOAWAY retry on cached HTTP/2 connections (commit 5754208)
- #459: utls upgraded to v1.8.2 security update (commit dcc8516)
- PR #485 closed (superseded by 8b018a5)
- PR #472 closed (merged as 1e34ab9)
- PR #486 closed (merged as d702aad)
- PR #491 closed (merged as 5754208)

### Still Open
- PR #478/#477: JA3 / ClientHelloSpec support — needs careful review, TLS fingerprint changes
- PR #465: Unmarshal status-code check — needs review
- #495: dialConn memory overusage — consider default MaxConnsPerHost
- #433: Large file upload memory buffering — middleware.go fix needed
- #397: concurrent map read/write in parseRequestHeader — needs locking
- #457: HTTP/3 AddConn/RoundTrip race condition
- #376: HTTP/2 frequent errors (9 comments)
- 50 open issues total, mostly feature requests and usage questions
