# Loop State

本文件是 Loop Engineering 自动维护循环的持久化状态。
所有循环从这里读取上次状态，执行后写回。不依赖模型上下文记忆。

## 依赖升级循环

### 运行历史
（首次运行，暂无记录）

### 待处理
- [ ] 首次运行：全量扫描可升级依赖

### 最近一次运行
- 日期：N/A
- 结果：未运行

---

## Issue 分流状态（人工分析，2026-06-27）

### 需优先处理

| # | 标题 | 类型 | 说明 |
|---|------|------|------|
| #482 | Adapt to quic-go v0.59.0 (ConnectionTracingID 移除) | **魔改同步·紧急** | quic-go v0.59.0 移除了 `ConnectionTracingID`/`ConnectionTracingKey`，导致编译失败。已有 PR #485，**需谨慎审查**——contributor 可能未完全考虑魔改兼容性。涉及 `internal/http3/transport.go`、`client.go`、`conn.go`。这是 quic-go 魔改同步，不能直接 merge |
| #489 | Security: 自定义 auth header 跨域重定向泄露 | **安全漏洞·高** | `SetCommonHeader` 设置的自定义 auth header（X-API-Key 等）在跨域重定向时不被剥离。CWE-200, CVSS 7.4。根因在 `middleware.go:528-540` 和 `client.go:334` |
| #485 | PR: upgrade quic-go to v0.59.0 | **PR 审查·谨慎** | 对应 #482。改动 260+/124-，涉及移除 deprecated API、替换 hijacker 回调为 RawClientConn、修复 SupportsDatagrams 检查。**必须人工逐文件比对魔改逻辑** |

### quic-go 相关（需谨慎）

| # | 标题 | 说明 |
|---|------|------|
| #460 | 建议用 x/net 的 quic 替换 quic-go | 作者已回复：等标准库成熟后切换。长期跟踪 |
| #457 | HTTP/3 AddConn 和 RoundTrip 竞态 | `internal/http3/transport.go` 中 `t.newClientConn` 初始化竞态，高并发下 panic。有复现堆栈和修复建议 |
| #372 | http3 是否支持 SetTLSFingerprint | 功能咨询，涉及 HTTP/3 + TLS 指纹 |

### Bug/性能问题

| # | 标题 | 说明 |
|---|------|------|
| #495 | dialConn 内存暴涨 | 无限连接池导致，建议设 MaxConnsPerHost。可考虑设默认上限 |
| #433 | 大文件上传全量缓冲到内存 | `middleware.go:122-123` multipart CreatePart 触发全量缓冲，670MB 文件吃 1.7GB 内存 |
| #397 | concurrent map read and map write | `middleware.go:537` parseRequestHeader 并发读写 client headers map。请求过程中修改 client header 导致 |
| #436 | 重试达上限不返回 error | 行为不符合预期 |
| #419 | Keep-Alive 长连接下 SetTimeout 错误断开 | |
| #416 | ParallelDownload 未关闭输出文件 | |
| #376 | HTTP/2 经常触发错误 | 9 条评论，可能涉及魔改的 `internal/http2/` |

### 功能请求

| # | 标题 | 说明 |
|---|------|------|
| #475 | 中间件中读取 retryOption | |
| #473 | Socks4 代理支持 | |
| #459/#454 | utls 指纹更新到 133 / 支持 Chrome133、Firefox133/135 | utls 版本升级 |
| #431 | jsonrpc2 支持 | |
| #425 | zstd 内容编码支持 | |
| #406 | response body size limit | |
| #404 | 生成 cURL 调试代码 | |
| #394 | graphql 请求支持 | |
| #369 | SSE (Server-Sent Events) 支持 | |

### 待审查 PR（非 quic-go）

| PR | 标题 | 说明 |
|----|------|------|
| #491 | fix: retry on GOAWAY errors (HTTP/2 缓存连接) | 涉及魔改 HTTP/2，需验证 |
| #486 | fix: SetCookieJarFactory 返回 http.CookieJar | 对应 #415 |
| #478/#477 | JA3 支持 / ClientHelloSpec 设置 | TLS 指纹相关 |
| #472 | chrome headers accept 添加 application/json | 对应 #471，小改动 |
| #465 | Unmarshal 应根据 status-code 检查 error |

---

## CI 修复循环

### 运行历史
（首次运行，暂无记录）

### 待人工处理
（无）

### 最近一次运行
- 日期：N/A
- 结果：未运行

---

## Issue 分流状态（人工分析，2026-06-27）

### 需优先处理

| # | 标题 | 类型 | 说明 |
|---|------|------|------|
| #482 | Adapt to quic-go v0.59.0 (ConnectionTracingID 移除) | **魔改同步·紧急** | quic-go v0.59.0 移除了 `ConnectionTracingID`/`ConnectionTracingKey`，导致编译失败。已有 PR #485，**需谨慎审查**——contributor 可能未完全考虑魔改兼容性。涉及 `internal/http3/transport.go`、`client.go`、`conn.go`。这是 quic-go 魔改同步，不能直接 merge |
| #489 | Security: 自定义 auth header 跨域重定向泄露 | **安全漏洞·高** | `SetCommonHeader` 设置的自定义 auth header（X-API-Key 等）在跨域重定向时不被剥离。CWE-200, CVSS 7.4。根因在 `middleware.go:528-540` 和 `client.go:334` |
| #485 | PR: upgrade quic-go to v0.59.0 | **PR 审查·谨慎** | 对应 #482。改动 260+/124-，涉及移除 deprecated API、替换 hijacker 回调为 RawClientConn、修复 SupportsDatagrams 检查。**必须人工逐文件比对魔改逻辑** |

### quic-go 相关（需谨慎）

| # | 标题 | 说明 |
|---|------|------|
| #460 | 建议用 x/net 的 quic 替换 quic-go | 作者已回复：等标准库成熟后切换。长期跟踪 |
| #457 | HTTP/3 AddConn 和 RoundTrip 竞态 | `internal/http3/transport.go` 中 `t.newClientConn` 初始化竞态，高并发下 panic。有复现堆栈和修复建议 |
| #372 | http3 是否支持 SetTLSFingerprint | 功能咨询，涉及 HTTP/3 + TLS 指纹 |

### Bug/性能问题

| # | 标题 | 说明 |
|---|------|------|
| #495 | dialConn 内存暴涨 | 无限连接池导致，建议设 MaxConnsPerHost。可考虑设默认上限 |
| #433 | 大文件上传全量缓冲到内存 | `middleware.go:122-123` multipart CreatePart 触发全量缓冲，670MB 文件吃 1.7GB 内存 |
| #397 | concurrent map read and map write | `middleware.go:537` parseRequestHeader 并发读写 client headers map。请求过程中修改 client header 导致 |
| #436 | 重试达上限不返回 error | 行为不符合预期 |
| #419 | Keep-Alive 长连接下 SetTimeout 错误断开 | |
| #416 | ParallelDownload 未关闭输出文件 | |
| #376 | HTTP/2 经常触发错误 | 9 条评论，可能涉及魔改的 `internal/http2/` |

### 功能请求

| # | 标题 | 说明 |
|---|------|------|
| #475 | 中间件中读取 retryOption | |
| #473 | Socks4 代理支持 | |
| #459/#454 | utls 指纹更新到 133 / 支持 Chrome133、Firefox133/135 | utls 版本升级 |
| #431 | jsonrpc2 支持 | |
| #425 | zstd 内容编码支持 | |
| #406 | response body size limit | |
| #404 | 生成 cURL 调试代码 | |
| #394 | graphql 请求支持 | |
| #369 | SSE (Server-Sent Events) 支持 | |

### 待审查 PR（非 quic-go）

| PR | 标题 | 说明 |
|----|------|------|
| #491 | fix: retry on GOAWAY errors (HTTP/2 缓存连接) | 涉及魔改 HTTP/2，需验证 |
| #486 | fix: SetCookieJarFactory 返回 http.CookieJar | 对应 #415 |
| #478/#477 | JA3 支持 / ClientHelloSpec 设置 | TLS 指纹相关 |
| #472 | chrome headers accept 添加 application/json | 对应 #471，小改动 |
| #465 | Unmarshal 应根据 status-code 检查 error |
