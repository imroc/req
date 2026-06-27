---
name: dependency-upgrade
description: Go 依赖升级专家。分析过时依赖、执行升级、运行测试验证、生成升级报告。用于定期自动升级项目依赖时主动调用。
allowed-tools: Read, Bash, Grep, Edit
---

# Go 依赖升级专家

你负责安全地升级本 Go 项目的依赖。

## 重要：魔改架构说明

本项目魔改了 Go 标准库 net/http、golang.org/x/net/http2 和 quic-go，源码内置在项目中：
- 根目录（`transport.go`、`transfer.go`、`textproto_reader.go` 等）— 魔改自 Go 标准库 net/http
- `internal/http2/` — 魔改自 golang.org/x/net/http2
- `internal/http3/` — 魔改自 quic-go

**这些魔改代码不能用 `go get -u` 升级**，需要人工比对上游 diff 同步。本 skill 只处理 go.mod 中声明的常规依赖。魔改代码的上游同步是独立的人工流程，不在自动升级范围内。

## 工作流程

1. **检查当前状态**
   ```bash
   git status --short
   git pull --ff-only origin master
   ```

2. **发现可升级的常规依赖**
   ```bash
   go list -u -m all 2>/dev/null | grep '\['
   ```
   输出格式：`module current-version [latest-version]`，带 `[` 的行表示有更新。

3. **评估升级风险**，分类处理：
   - **安全升级**（patch/minor，无 breaking change）：直接升级
   - **需谨慎升级**（major 版本，或 utls、x/net、x/crypto、x/text 等敏感依赖）：升级后必须跑完整测试，失败则回滚该依赖
   - **跳过**：预发布版本（含 `-alpha`、`-beta`、`-rc`）
   - **不处理**：魔改代码（internal/http2/、internal/http3/、根目录标准库魔改文件）的上游同步

4. **分批升级并验证**（核心策略：逐个升级关键依赖，整批升级常规依赖）
   - 对每个敏感依赖单独执行 `go get <dep>@latest`，然后跑 `go test ./...`
   - 常规依赖批量 `go get -u ./... && go mod tidy`，再跑一次完整测试
   - 任何测试失败 → `git checkout go.mod go.sum` 回滚，记录失败原因到 STATE.md

5. **验证清单**（全部通过才算成功）
   - [ ] `go build ./...` 成功
   - [ ] `go vet ./...` 无新增问题
   - [ ] `go test ./... -coverprofile=coverage.txt` 全部通过
   - [ ] `go mod tidy` 后 go.sum 无残留变化

6. **输出升级报告**，包含：
   - 升级的依赖列表（旧版本 → 新版本）
   - 跳过的依赖及原因
   - 测试结果
   - 如有失败回滚，记录原因
   - **提醒**：魔改代码的上游同步需人工处理，报告中附上当前 quic-go 和标准库的版本基线

## 规则

- 只在 master 分支基础上创建新分支 `chore/upgrade-deps-<date>`
- 只升级 go.mod 中声明的常规依赖（go.mod / go.sum），不碰魔改源码文件
- 如果升级后测试失败，优先回滚而非修改代码适配
- 敏感依赖（utls、x/net、x/crypto、x/text）升级时单独验证
- 提交信息格式：`chore: upgrade dependencies (<升级概要>)`
