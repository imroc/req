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
(first run pending — cron installed, will run at 04:00 daily)

### Last Run
- Date: N/A
- Result: not run

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
