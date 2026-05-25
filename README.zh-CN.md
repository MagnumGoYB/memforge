# MemForge

[English](README.md)

![MemForge 项目概览](README-overview-zh-CN.webp)

MemForge 是面向 Claude Code、Codex、Cursor、Gemini CLI 等 AI coding agents 的本地优先项目记忆层。

它把结构化项目记忆保存在用户本机，用 SQLite + FTS5 建立索引，并在 token budget 内编译为可直接提供给 agent 的上下文。Markdown 始终是真值源；SQLite 索引可完全重建。

## 它是什么

MemForge 是一个面向 AI coding workflow 的本地项目上下文编译器。

它的目标是在不污染工作仓库的前提下，为 agent 和开发者提供可持续积累的项目记忆。

## 它不是什么

- 不是通用知识库
- 不是聊天记录归档系统
- 不是云端记忆服务
- 不是仓库自动扫描器
- 不是常驻后台守护进程

## 当前状态

当前仓库已经具备可用的本地优先 MVP 命令链路。

目前已实现：

- `version`
- `help`
- `init`
- `remember`
- `search`
- `context`
- `before`
- `after`
- `reindex`
- `mcp`
- `diff-summary`
- hybrid `search --hybrid`
- `after --adapter` session adapters
- `debug paths`
- governance、harness 与 GitHub workflow 护栏

## 安装

根据使用方式选择一条路径。

### 先更新宿主应用

在更新 MemForge 插件之前，先把运行它的宿主应用更新到最新版：

```bash
codex update
claude update
```

如果你还在用 Claude Code 的旧命令名，`claude upgrade` 也可以。若你使用的是 Codex.app，请先通过它自己的应用更新通道把桌面应用更新到最新版，然后再重新安装或刷新插件包。

### 1. Claude Code 插件安装

这是 Claude Code 用户推荐路径。通过 Claude Code marketplace 或 release catalog 安装插件：

```txt
/plugin marketplace add <marketplace-or-release-catalog>
/plugin install memforge@<marketplace-name>
/reload-plugins
```

发布版插件包包含各平台的 `memforge` runtime binary，并通过 `bin/run-memforge-mcp-launcher.sh` 启动 MCP；该 wrapper 会先从常见安装路径寻找 Node，再启动 `bin/memforge-mcp-launcher.js`。用户不需要先运行 `go install`，也不需要把单独的 `memforge` CLI 放到 `PATH` 上。

### 2. Codex 插件安装

> **Codex 官方插件商店状态：** OpenAI 的自助插件发布尚未开放（截至 2026 年 5 月，OpenAI 文档标注为 "coming soon"）。MemForge 目前未出现在 Codex 官方插件浏览器（`/plugins`）中。当公开发布通道开放且 MemForge 被收录后，用户可直接从 Codex 插件浏览器安装。

在此之前，测试验收和对外分发应优先使用 Git marketplace/catalog。Codex.app 内升级只支持 Git marketplace/catalog 安装。验收优先走 Git marketplace 安装路径，才能覆盖真实用户的自动升级行为。

**A. Git marketplace/catalog（验收与升级优先路径）**

仓库提供 Codex marketplace entry：`.agents/plugins/marketplace.json` 与 `marketplaces/codex/marketplace.json`。可以从仓库 URL 添加 Git marketplace/catalog，或从指向 Git-backed entry 的 checkout 添加：

```bash
codex plugin marketplace add https://github.com/MagnumGOYB/memforge
codex plugin add memforge@memforge-codex-marketplace
```

这是推荐验收路径，因为 Codex.app 可以基于 Git marketplace/catalog source 做应用内升级。

**B. 本地/私有 marketplace 或 Codex host 插件包**

当 Codex host 支持本地/私有 marketplace 或 plugin package 安装时，通过该 host 安装打包后的 `dist/plugins/codex/memforge` bundle。打包 bundle 包含各平台的 `memforge` runtime，并通过 `bin/run-memforge-mcp-launcher.sh` 启动；用户不应需要预先安装 CLI。

对于非 Git marketplace 安装，`check_update` MCP tool 只负责检测新版并返回重装指引，不承诺 Codex.app 自动升级。

**C. 直接 MCP 注册（推荐 CLI fallback）**

```bash
go install github.com/MagnumGOYB/memforge/cmd/memforge@latest
codex mcp add memforge -- memforge --no-version-check mcp
```

这使用 `PATH` 上的 `memforge` binary，不是 packaged plugin runtime path。

**D. 本地 marketplace discovery（仅开发/调试）**

Codex CLI 0.132 暴露 marketplace 管理和 plugin install/remove 命令。在运行 `make package-plugins` 后，从打包 checkout：

```bash
codex plugin marketplace add "$PWD"
codex plugin add memforge@memforge-codex-marketplace
```

这会从 `dist/plugins/codex/memforge` 安装打包 bundle；它不应依赖 `PATH` 上的单独 `memforge` binary。

当前 Claude Code 与 Codex 分发细节见 `docs/plugin-distribution.md`。

### 手动更新插件

打包插件会自带 `memforge` runtime。只更新单独安装的 CLI，不会更新已经安装的插件包。

如果是 Codex Git marketplace/catalog 安装，直接从同一个 marketplace entry 重新安装插件：

```bash
codex plugin remove memforge
codex plugin add memforge@memforge-codex-marketplace
```

如果是 Codex 本地 checkout marketplace 安装，先刷新 checkout、重新打包，再重装：

```bash
cd /path/to/memforge
git pull
make package-plugins
codex plugin remove memforge
codex plugin marketplace add "$PWD"
codex plugin add memforge@memforge-codex-marketplace
```

如果是 Claude Code 本地 checkout marketplace 安装，刷新 checkout、重新打包、重装后 reload plugins：

```bash
cd /path/to/memforge
git pull
make package-plugins
claude plugin marketplace add "$PWD"
claude plugin install memforge@memforge-marketplace
```

然后在 Claude Code 内执行：

```txt
/reload-plugins
```

如果是直接 MCP 注册，则更新 `PATH` 上的 CLI binary：

```bash
MEMFORGE_VERSION=latest \
  curl -fsSL https://raw.githubusercontent.com/MagnumGOYB/memforge/main/scripts/install.sh | bash
```

### 3. CLI 安装

使用 curl 安装最新 release binary：

```bash
curl -fsSL https://raw.githubusercontent.com/MagnumGOYB/memforge/main/scripts/install.sh | bash
```

脚本默认安装到 `$HOME/.local/bin`。需要时可以覆盖安装目录或版本：

```bash
MEMFORGE_INSTALL_DIR="$HOME/bin" MEMFORGE_VERSION="latest" \
  curl -fsSL https://raw.githubusercontent.com/MagnumGOYB/memforge/main/scripts/install.sh | bash
```

curl 安装脚本支持 macOS 与 Linux 的 arm64/amd64。Windows 用户请手动下载 release asset，或使用下面的 Go install 路径。

Go install 要求：

- Go 1.26.3 或更新版本

安装最新发布的 module 版本：

```bash
go install github.com/MagnumGOYB/memforge/cmd/memforge@latest
```

从本地 checkout 安装：

```bash
git clone https://github.com/MagnumGOYB/memforge.git
cd memforge
make build
go install ./cmd/memforge
```

面向 agent 和自动化调用时，建议显式设置本地存储根目录：

```bash
export MEMFORGE_HOME="$HOME/.local/share/memforge"
```

## 快速开始

构建 CLI：

```bash
make build
```

运行当前命令：

```bash
make run ARGS="version"
MEMFORGE_HOME=/absolute/path make run ARGS="init --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="remember --kind decision --title 'Repository layer is framework-agnostic' --format json --no-version-check 'Body'"
MEMFORGE_HOME=/absolute/path make run ARGS="search --format json --no-version-check 'repository framework'"
MEMFORGE_HOME=/absolute/path make run ARGS="context --budget 3000 --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="before --budget 3000 --format json --no-version-check 'Refactor repository layer'"
MEMFORGE_HOME=/absolute/path make run ARGS="after --from /absolute/path/session.md --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="after --from /absolute/path/session.md --approve all --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="reindex --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="mcp --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="search --hybrid --format json --no-version-check 'repository framework'"
MEMFORGE_HOME=/absolute/path make run ARGS="after --adapter claude-code --from /absolute/path/session.jsonl --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="diff-summary --from /absolute/path/numstat.txt --format json --no-version-check"
MEMFORGE_HOME=/absolute/path make run ARGS="debug paths --format json --no-version-check"
```

## 项目配置

`context`、`before` 以及 MCP context tools 会读取用户配置文件和项目根目录下的可选 TOML 配置：

```toml
default_budget = 1200

[kind_weights]
decision = 99
bugfix = 95
constraint = 100
```

项目 `.memoryrc` 会覆盖 `$XDG_CONFIG_HOME/memforge/config.toml`。CLI `--budget` 和 MCP tool 的 `budget` 参数会覆盖两者。配置不存储 memory 内容；memory 仍只保存在 `MEMFORGE_HOME` 或 `$XDG_DATA_HOME/memforge` 下。

## AI agent 调用约定

对自动化和 agent 调用，优先使用：

```bash
memforge --no-version-check <command> --format json
```

约定：

- JSON payload 只写 stdout。
- 人类可读 warning 写 stderr。
- `--no-version-check` 是自动化契约的一部分。
- `context`、`before`、`compile_context` 与 `get_project_context` 的 JSON 返回会包含 `estimated_tokens` 与 `usage.estimated_tokens`，用于表示 MemForge 对上下文大小的本地估算。

## 本地优先与隐私边界

- 记忆不会写入用户仓库。
- 存储路径解析到 `MEMFORGE_HOME` 或 `$XDG_DATA_HOME/memforge`。
- 默认行为是本地优先、默认离线。
- Markdown 是规范存储格式。
- SQLite 只是可重建的索引层。
- MVP 命令不会触发 opt-in LLM/provider 调用。
- `after` 是 proposal-first：它只从显式提供的 session 文件提取 candidate memories，只有传入 `--approve` 才会持久化。
- enabled MCP plugin 可以让 agent 在 thread 执行期间通过 `upsert_project_memory` 创建或更新稳定 project memories；这仍需要明确 tool call，不会自动扫描仓库，也不会调用远程 provider。
- provider-backed extraction 是 opt-in，且仅限 `after`；其他命令保持本地/离线。
- Hybrid search 必须显式使用 `search --hybrid`，默认使用本地 deterministic embeddings。
- Session adapters 与 diff summaries 都只是对显式提供的文件或本地 git 输出做本地转换。

## 开发

请优先使用仓库 Makefile 目标，让本地和 agent 校验时的 Go cache 固定在 `.cache/memforge`：

```bash
make setup
make check
make test
make test-packages PKGS="./internal/index ./internal/compiler"
make test-harness
make vet
make build
make validate
make validate-pr-body
```

## 开源项目文档

- 贡献指南：`CONTRIBUTING.md` / `CONTRIBUTING.zh-CN.md`
- Agent 执行指南：`AGENTS.md` / `AGENTS.zh-CN.md`
- Claude Code 入口：`CLAUDE.md` / `CLAUDE.zh-CN.md`
- Harness 工程：`docs/harness-engineering.md` / `docs/zh-CN/harness-engineering.md`
- GitHub 自动化：`docs/github-automation.md` / `docs/zh-CN/github-automation.md`
- Memory 格式：`docs/memory-format.md` / `docs/zh-CN/memory-format.md`
- MCP server：`docs/mcp.md` / `docs/zh-CN/mcp.md`
- Agent integrations：`docs/integrations.md` / `docs/zh-CN/integrations.md`
- Plugin distribution：`docs/plugin-distribution.md` / `docs/zh-CN/plugin-distribution.md`
- 开发计划：`plan.md` / `plan.zh-CN.md`

## 模块路径

```bash
go install github.com/MagnumGOYB/memforge/cmd/memforge@latest
```

若在当前 checkout 内本地开发：

```bash
go install ./cmd/memforge
```
