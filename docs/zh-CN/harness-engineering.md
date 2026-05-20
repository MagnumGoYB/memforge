# Harness 工程

[English](../harness-engineering.md)

Harness 由两部分组成：在编辑前引导代理的前馈指南，以及在编辑后捕获漂移的反馈传感器。

对 MemForge 来说，harness 刻意保持轻量：Go 测试、Makefile、PR 元数据校验、CI 闸门和精炼的代理指南。它在不引入任何后台服务的前提下，守住离线、不污染仓库的记忆契约。

## 前馈指南

- `AGENTS.md` 与 `AGENTS.zh-CN.md`：仓库使命、编码约束、验证矩阵、隐私边界、交付规则。
- `README.md`：面向用户的 CLI 用法与安装路径。
- `CONTRIBUTING.md`：贡献者验证与离线优先规则。
- `Makefile`：本地标准命令：`make setup`、`make check`、`make test`、`make test-packages`、`make test-harness`、`make vet`、`make build`、`make validate`、`make validate-pr-body`、`make commitlint`。
- `tools/commitlint` 与 `.githooks/commit-msg`：仓库原生 Go 提交信息校验，强制 `{emoji} {type}{scope}: {subject}` 格式，不依赖 Node/npm。每个 type 严格对应一个 emoji：`✨ feat`、`🐛 fix`、`📝 docs`、`👷 ci`、`💄 style`、`♻️ refactor`、`🔖 release`、`⚡️ perf`、`✅ test`、`🔧 chore`、`🏗️ build`。`make setup` 启用本地 hook。
- `.github/pull_request_template.md`：PR 模板，覆盖需求分类、验收标准、测试证据、Validation、回滚与残余风险。
- 发版策略：工程或流程优化无需发版；feature 与 bugfix 必须跟进发版或显式延期。
- `.github/workflows/ci.yml`：与本地一致的托管校验。

## 反馈传感器

- `make test`：覆盖项目检测、记忆模型、Markdown I/O、SQLite/FTS、编译器和 CLI 命令的单测与集成测试。
- `make test-packages PKGS="./internal/index ./internal/compiler"`：指定包测试，复用与全量测试相同的仓库本地 Go cache。
- `make test-harness`：仓库结构传感器，校验代理文档、Makefile 命令、CI 闸门、PR 模板、离线/隐私约束，以及仓库受控文档的专有名词大小写（`docs/casing-rules.txt`）。大小写传感器会跳过生成缓存、vendored dependencies、`node_modules` 与 Git metadata，即使它们位于嵌套 worktree 目录下。
- `make vet`：静态分析。
- `make build`：单二进制构建检查。
- `make validate-pr-body`：可执行的 PR 元数据闸门。
- `make setup`：一次性本地设置，运行 `git config core.hooksPath .githooks`。
- `make commitlint COMMIT_MSG_FILE=<commit-msg-file>`：单条提交信息闸门，`make setup` 后经 `.githooks/commit-msg` 串联。
- `make commitlint-range COMMIT_RANGE=<base..head>`：PR 范围提交闸门，PR CI 同步校验。
- `.cache/memforge/`：仓库本地、git 忽略的 Go build/module cache，让代理验证固定在本仓库 checkout，而不是临时 `/tmp` 路径。
- 沙箱内 agent session 不要直接跑 raw `go test`、`go vet`、`go build`、`go run`。统一使用 Makefile 目标，让 Go 缓存不会回落到 `~/Library/Caches/go-build`。

## 代理工作流契约

- 在编辑前完成需求分类：feature、bugfix、refactor、harness/tooling、analysis-only。
- 实现前锁定可观测验收标准。
- 交付前记录发版判定。
- 行为有改动且条件允许时，先加一条失败测试、harness 传感器或显式手工验证清单。
- 改动严格限定在指定文件或目录。
- 交付前把每条验收标准映射到证据。
- 写明跳过的校验与残余风险。

## MemForge 专属护栏

- 默认不加网络上传、同步、遥测、远程报告。命令启动期唯一允许的网络行为是 GitHub Release 版本检查，必须支持 `--no-version-check` 或 `MEMFORGE_NO_VERSION_CHECK=1` 关闭。
- 不向用户仓库写入。记忆只存于 `MEMFORGE_HOME` 或 `$XDG_DATA_HOME/memforge`。
- 不自动扫描、哈希、指纹化用户源文件。
- Markdown 是真值源；SQLite 必须可由 `memforge reindex` 完全重建。
- LLM Provider 调用是 opt-in。MVP 命令（`init`、`remember`、`search`、`context`、`before`、`reindex`）不得触发网络 LLM 调用。
- CLI 自动化输出保持稳定：`--format json` 是契约。

## 修改 Harness

修改 harness、CI、PR 工作流或校验脚本时：

1. 更新可执行脚本/测试/workflow。
2. 行为有变时同步 `AGENTS.md` 与 `AGENTS.zh-CN.md`。
3. 更新本文档与 `docs/harness-engineering.md`。
4. PR 元数据规则变化时跑 `make check`、`make test-harness`、`make validate-pr-body`。

提交工作流规则变化时，同步 `tools/commitlint`、`.githooks/commit-msg`、`Makefile`、`.github/workflows/pr.yml`、`AGENTS.md`、`AGENTS.zh-CN.md`、本文档与 `docs/harness-engineering.md`。

发版判定规则变化时，同步 `tools/validate-pr-body`、`.github/pull_request_template.md`、`.github/pull_request_template.zh-CN.md`、`.github/workflows/pr-review.yml`、`AGENTS.md`、`AGENTS.zh-CN.md`、本文档、`docs/harness-engineering.md`，以及 `docs/github-automation.md` / `docs/zh-CN/github-automation.md`。

## GitHub 自动化覆盖

- `docs/github-automation.md` 与 `docs/zh-CN/github-automation.md` 记录 PR Review checklist、Dependabot auto-merge 与 CI 流程。
- `.github/workflows/pr.yml` 校验真实 PR 元数据与 PR 范围内所有提交信息。
- `.github/workflows/pr-review.yml` 用 `issues: write` 与 `pull-requests: write` 权限创建或更新 PR Issue 评论里的 Review checklist。
- `.github/CODEOWNERS` 把高风险区域路由给维护者；仓库 Review 治理保持 GitHub 原生，不引入付费 AI Review 自动化。
- `.github/workflows/ci.yml` 在 push 与 pull request 时执行 `make validate` 与 `make test-harness`。
