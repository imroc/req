# req 项目上下文

本项目是 `imroc/req`（v3），一个 Go HTTP 客户端库，支持 HTTP/1.1、HTTP/2、HTTP/3，可自动嗅探协议并选择最优版本。

## 基本信息

- **模块路径**: `github.com/imroc/req/v3`
- **Go 版本**: 1.24+（CI 测试 1.24.x 和 1.25.x）
- **主分支**: `master`
- **包管理**: Go modules（`go.mod` / `go.sum`）

## 架构：魔改标准库与第三方库

本项目对以下上游代码进行了**魔改（源码内置并修改）**，而非直接依赖：

1. **Go 标准库 net/http** — 魔改后放在根目录（`transport.go`、`transfer.go`、`textproto_reader.go`、`http.go`、`http_request.go`、`response.go` 等），保留 Go Authors 版权头。支持 dump 内容、middleware、HTTP/3 接入等。
2. **Go 标准库 HTTP/2**（golang.org/x/net/http2）— 魔改后放在 `internal/http2/`。
3. **quic-go** — 魔改后放在 `internal/http3/`，支持 HTTP/3 和协议自动嗅探。

**这意味着不能用 `go get -u` 直接合并上游更新**。每次更新需要：
- 跟踪 Go 标准库 net/http 和 quic-go 的上游变更
- 手动比对 diff，将上游改动同步到 req 中对应的魔改文件，同时保留 req 的定制逻辑
- 固定更新到 Go 标准库 HTTP 最新稳定版 + quic-go 最新稳定版
- 提交信息惯例：`merge upstream net/http: <日期>(<commit短hash>)`、`merge upstream http2: <日期>(<hash>)`、`port quic-go <版本>`

## 常用命令

```bash
# 运行全部测试（与 CI 一致）
go test ./... -coverprofile=coverage.txt

# 运行单个包测试
go test ./internal/...

# 查看可升级的常规依赖（非魔改的）
go list -u -m all

# 同步上游（人工操作，非自动）
# 1. 对照 Go 标准库 src/net/http 最新 commit
# 2. 逐文件比对 diff，同步到根目录魔改文件
# 3. 对照 quic-go 最新 release，同步到 internal/http3/
```

## 关键依赖

- **quic-go**（魔改于 `internal/http3/`）— HTTP/3 支持，版本敏感，升级需重点回归测试。contributor 提交的 quic-go 相关 PR/issue 需谨慎审查，通常考虑不到魔改兼容性
- `github.com/refraction-networking/utls` — TLS 指纹模拟
- `golang.org/x/net`、`golang.org/x/text`、`golang.org/x/crypto` — Go 官方扩展库
- `github.com/andybalholm/brotli`、`github.com/klauspost/compress` — 压缩

## 架构要点

- 根目录核心文件：`client.go`、`request.go`、`response.go`、`transport.go`（魔改自标准库）、`middleware.go`
- `internal/http2/` — 魔改自 golang.org/x/net/http2
- `internal/http3/` — 魔改自 quic-go
- `internal/` 其他 — 内部工具（digest、dump、compress、charsets 等）
- `pkg/` — 公共辅助包
- `http2/` — HTTP/2 帧定义（priority、setting）
- `examples/` — 使用示例

## CI

- `.github/workflows/ci.yml` — push/PR 到 master 时运行 `go test ./...`
- 忽略 `**.md` 变更

## 维护约定（Loop Engineering）

本项目采用 Loop Engineering 理念进行自动化维护。循环配置位于 `.codebuddy/`，总体设计见 [loop-engineering.md](./loop-engineering.md)。

**运行方式**：通过系统 cron + `codebuddy -p`（headless）每天凌晨自动运行，无需保持会话。安装：`.codebuddy/scripts/install-cron.sh`，手动触发：`.codebuddy/scripts/run-loop.sh <类型>`，日志目录由 `LOOP_LOG_DIR` 环境变量控制（默认 `/tmp/loop-logs/`）。

- 所有循环状态持久化到 `STATE.md`（项目根目录），不依赖模型上下文记忆
- 每个循环遵循五阶段：Discover → Plan → Execute → Verify → Iterate
- 实现与验证分离：不同 agent 负责修复和审查
- 循环必须有明确的停止条件，设置最大迭代次数
- **涉及 GitHub 操作时（issue、PR、CI run、workflow、release 等），必须加载 `gh-cli` 技能**获取命令参考，确保使用正确的命令语法。对应 agent 在 `skills` 字段中声明 `gh-cli`
- **循环可自行做日常维护决断**（升级依赖、修 CI、出 PR 等），但**禁止打 tag 发版**。`git tag`、`gh release create` 等发版操作始终由人决定，自动化只到出 PR 为止
- **所有 git commit message 必须用英文**
