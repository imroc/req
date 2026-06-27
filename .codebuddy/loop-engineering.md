# Loop Engineering 设计说明

本项目使用 Loop Engineering 理念实现自动化维护。核心理念：从"逐次提示 AI"转为"设计自循环系统"，人变成规则制定者，AI 系统自运行达成目标。

## 四层架构

| 层 | 作用 | 本项目实现 |
|----|------|-----------|
| Prompt 层 | 怎么问 | 每个 agent 的系统提示 |
| Context 层 | 让 AI 看到什么 | `CODEBUDDY.md`、`STATE.md` |
| Harness 层 | AI 工作环境 | skills（项目知识）、allowed-tools（权限约束） |
| Loop 层 | 做完一步怎么办 | GitHub Actions 触发、agent 五阶段迭代 |

## 已实现的循环

### 1. 依赖升级循环（闭环）
- **触发**：每周一 03:00 UTC，或手动
- **目标**：安全升级所有可升级的 Go 依赖
- **五阶段**：发现(`go list -u`) → 分级 → 升级 → 测试 → 回滚重试
- **停止条件**：所有依赖处理完 / 测试连续失败 3 次 / 迭代 5 次
- **文件**：
  - skill: `skills/dependency-upgrade/SKILL.md`
  - agent: `agents/dependency-upgrader.md`
  - workflow: `.github/workflows/dependency-upgrade-loop.yml`
- **验证分离**：dependency-upgrader 执行升级 → code-reviewer 审查 → 通过才出 PR

### 2. CI 失败自修复循环（闭环）
- **触发**：CI workflow 失败时，或手动
- **目标**：自动修复可修复的 CI 失败
- **五阶段**：发现(`gh run list`) → 分类(ci-triage) → 修复(ci-fix) → 本地验证 → 重试
- **停止条件**：所有可修复失败已修复 / 修复尝试 3 次 / 无失败 CI
- **文件**：
  - skills: `skills/ci-triage/SKILL.md`、`skills/ci-fix/SKILL.md`
  - agents: `agents/ci-fixer.md`、`agents/code-reviewer.md`
  - workflow: `.github/workflows/ci-fix-loop.yml`
- **验证分离**：ci-fixer 修复 → code-reviewer 审查 → 通过才出 PR

## 状态管理

`STATE.md`（项目根目录）是所有循环的共享状态文件，每次循环迭代后更新。这是 Loop Engineering 的核心原则——**状态存于外部，不依赖模型上下文**。

## 成本控制

- 每个循环设最大迭代次数（依赖升级 5 次，CI 修复 3 次）
- 使用 `model: inherit` 复用主会话模型，避免重复加载
- 循环只在需要时运行（事件触发或定时），非常驻
- GitHub Actions 提供执行环境，无额外服务成本

## 自主权边界

- **循环可自行做日常维护决断**：升级依赖、修 CI、出 PR 等，无需人工逐步确认
- **禁止打 tag 发版**：`git tag`、`gh release create` 等发版操作始终由人决定。自动化只到出 PR 为止
- code-reviewer agent 对任何包含发版操作的改动一律 block

## 如何新增循环

1. 在 `skills/` 下创建新 skill 目录和 `SKILL.md`（编码项目知识）
2. 在 `agents/` 下创建 agent（定义职责、停止条件、工具权限）
3. 在 `.github/workflows/` 下创建触发工作流（定时或事件触发）
4. 在 `STATE.md` 添加对应章节
5. 复用 `code-reviewer` agent 实现验证分离
6. 涉及 GitHub 操作的 agent 在 `skills` 字段声明 `gh-cli`
7. agent 安全约束中写明禁止发版

## 未来可扩展的循环

| 循环 | 触发 | 价值 |
|------|------|------|
| Issue 分流 | 新 Issue | 自动分类、打标签、初步回复 |
| 测试覆盖提升 | 定时 | 找低覆盖文件、补测试 |
| CHANGELOG 生成 | tag/定时 | 汇总变更、生成发布说明 |
| 安全扫描 | 定时 | 检查已知漏洞依赖 |
