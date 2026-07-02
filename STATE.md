# Loop State

This file is the persistent state for Loop Engineering automated maintenance loops.
All loops read from here and write back after execution. State is external, not in model context.

## Dependency Upgrade Loop

### Run History
- 2026-07-03: Upgraded 7 dependencies (brotli v1.2.2, go-querystring v1.2.0, compress v1.19.0, x/net v0.56.0, x/crypto v0.53.0, x/text v0.38.0, x/sys v0.46.0). quic-go v0.60.0 and utls v1.8.2 unchanged (latest). All tests pass. PR #503 created.

### Last Run
- Date: 2026-07-03
- Result: success — 7 deps upgraded, all tests pass, PR #503 created

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
- 2026-06-27 (run 1): Triaged 10 open issues, applied 14 labels, created 4 labels (security, performance, quic-go, tls-fingerprint), posted 3 comments (#496, #481, #464)
- 2026-06-27 (run 2): Triaged 15 unlabeled issues (#459–#423), applied 38 labels, created 6 labels (http2, modified-stdlib, priority:critical, priority:high, priority:medium, priority:low), posted 1 comment (#457)
- 2026-06-29: Triaged 10 open issues, found 2 unlabeled (#406, #404), applied 4 labels (enhancement, priority:low), posted 2 comments (#406, #404). No quic-go/HTTP/3 issues found in this batch.
- 2026-06-29 (run 2): Triaged 10 open issues (#475–#404), all already labeled — 0 unlabeled, 0 labels applied. 2 quic-go related (#460, #457) already correctly tagged. Posted 2 initial responses (#475, #431) on previously unanswered issues.

### Last Run
- Date: 2026-06-29 (run 2)
- Result: success — all 10 issues already labeled, 0 labels applied, 2 comments posted (#475, #431), 2 quic-go related (#460, #457) confirmed labeled

### Triage Details (2026-06-27, run 2)
| # | Title | Type | Extra | Priority | Comment |
|---|---|---|---|---|---|
| #459 | utls fingerprint update to 133 | enhancement | tls-fingerprint | low | No |
| #457 | HTTP/3 AddConn/RoundTrip race | bug | quic-go | high | Yes — noted modified-code caveat, high priority |
| #454 | Support firefox133/chrome133 | enhancement | tls-fingerprint | low | No |
| #453 | Better performance than resty | question | — | low | No |
| #452 | Rotating proxy best practices | question | — | low | No |
| #446 | Win7 compatibility (quic-go) | question | quic-go | low | No |
| #445 | Proxy set but connects to localhost | bug | — | medium | No |
| #444 | Put without body: Content-Length: 0 | question | — | low | No |
| #437 | Switching proxy after connection not working | bug | modified-stdlib | medium | No |
| #436 | Retry limit reached but no error returned | question | — | low | No |
| #433 | Large file upload buffered in memory | bug | performance, modified-stdlib | high | No |
| #431 | jsonrpc2 support | enhancement | — | low | No |
| #425 | zstd content encoding support | enhancement | — | low | No |
| #424 | Content-Type charset not applied in headers | bug | — | medium | No |
| #423 | SetSuccessResult for raw response | question | — | low | No |

### Labels Created This Run (run 2)
- `http2` (0052cc) — Involves HTTP/2 or internal/http2 modified code
- `modified-stdlib` (c5def5) — Involves modified Go stdlib files (transport.go, transfer.go, etc.)
- `priority:critical` (b60205) — Security vulnerability or compilation failure affecting all users
- `priority:high` (d93f0b) — Data loss, memory leak, or panic in production
- `priority:medium` (fbca04) — Bug with available workaround
- `priority:low` (0e8a16) — Minor issue or feature request

### Needs Human Attention
- **#496** — Compilation failure: gin v1.12.0 pulls quic-go v0.59.0, but internal/http3/ uses removed `quic.ConnectionTracingID` API. High priority, breaks user builds.
- **#484** — Inconsistent error handling between client-level and request-level afterResponse middleware loops (request-level stops on first error, client-level doesn't). May warrant code change.
- **#457** — HTTP/3 race condition between AddConn and RoundTrip in modified `internal/http3/` code. Potential panic in high-traffic environments. Needs investigation of initOnce bypass.
- **#433** — Large file upload (100MB+) causes entire file to be buffered in memory (2x+ memory usage). Likely fix needed in middleware.go upload path.

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
- Go stdlib net/http: 2026-07-02 (aa44f96)
- golang.org/x/net/http2: 2026-07-01 (bd5f1dc)
- quic-go: v0.60.0
- utls: v1.8.2

### Run History
- 2026-07-03: Checked all 3 upstreams. Found 70+ commits since last baselines. 11 security fixes (CVE-2026-33814, CVE-2026-56853, header injection via trailers, sensitive header leakage on redirect, HTTP/2 DoS/deadlock fixes). quic-go v0.60.0 released (Go 1.25+, FIPS 140-3). x/net/http2 major architectural change (Go 1.27 wrapping). Opened issue #502 with full sync report.

### Last Run
- Date: 2026-07-03
- Result: success — issue #502 created with full sync report, baselines updated, STATE.md committed

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
- 65 open issues total (25 now labeled), mostly feature requests and usage questions
