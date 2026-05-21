# 插件分发

[English](../plugin-distribution.md)

本仓库提供面向 Claude Code 与 Codex 的 MemForge 插件包。发布包面向普通用户的安装方式是 marketplace 或 release bundle；用户不需要先把 `memforge` CLI 安装到 `PATH`。

## Claude Code

Claude Code 支持 plugin marketplaces。本仓库包含：

```txt
dist/plugins/claude-code/memforge/.claude-plugin/plugin.json
dist/plugins/claude-code/memforge/.mcp.json
.claude-plugin/marketplace.json
```

### 普通用户安装

普通用户路径是通过 marketplace 或 release plugin 安装：

```txt
/plugin marketplace add <marketplace-or-release-catalog>
/plugin install memforge@<marketplace-name>
/reload-plugins
```

发布版插件包会在插件的 `bin/<platform>/` 目录内包含对应平台的 `memforge` runtime binary。用户在安装 Claude Code marketplace 包之前，不需要运行 `go install`，也不需要把单独安装的 `memforge` binary 放到 `PATH` 上。

`memforge-marketplace` entry 指向 `dist/plugins/claude-code/memforge` 下的打包 bundle，所以从源码 checkout 添加或刷新 marketplace 前需要先运行 `make package-plugins`：

```bash
make package-plugins
claude plugin marketplace add "$PWD"
claude plugin install memforge@memforge-marketplace
```

插件的 MCP 配置通过 bundled Node launcher 启动 server：

```json
{
  "command": "node",
  "args": ["${CLAUDE_PLUGIN_ROOT}/bin/memforge-mcp-launcher.js"],
  "env": {
    "MEMFORGE_PLUGIN_ROOT": "${CLAUDE_PLUGIN_ROOT}"
  }
}
```

Launcher 会解析当前平台，选择 bundled runtime，并从该 runtime 启动 stdio MCP server。

### 本地开发 smoke

从 source checkout 开发时，仍然可以直接在本仓库 smoke 插件：

```bash
make build
go install ./cmd/memforge
claude --plugin-dir ./plugins/claude-code/memforge
```

这个 source-checkout 流程仅用于本地开发和调试。它不是普通用户安装路径，也不应成为 marketplace/release 安装的前置条件。

## Codex

> **Codex 官方插件商店状态：** OpenAI 的自助插件发布尚未开放（截至 2026 年 5 月，OpenAI 文档标注为 "coming soon"）。MemForge 目前未出现在 Codex 官方插件浏览器（`/plugins`）中。当公开发布通道开放且 MemForge 被收录后，用户可直接从 Codex 插件浏览器安装，无需任何手动配置。

Codex 支持 plugin manifest 与 marketplace/catalog 安装流程。本仓库提供本地/私有分发包：

```txt
dist/plugins/codex/memforge/.codex-plugin/plugin.json
dist/plugins/codex/memforge/.mcp.json
.agents/plugins/marketplace.json
```

打包后的 Codex plugin bundle 同样包含各平台的 `memforge` runtime，并通过 MCP 配置使用 `bin/memforge-mcp-launcher.js`。当 Codex host 支持本地/私有 marketplace 或 plugin package 安装时，用户不应需要预先把 `memforge` CLI 安装到 `PATH`。

Codex MCP 配置使用相对插件根目录的 launcher 路径，不能依赖 Claude Code 专用环境变量：

```json
{
  "command": "node",
  "args": ["./bin/memforge-mcp-launcher.js"],
  "cwd": ".",
  "default_tools_approval_mode": "approve"
}
```

Codex CLI 0.132 暴露 marketplace 管理和 plugin install/remove 命令。`memforge-codex-marketplace` entry 指向 `dist/plugins/codex/memforge` 下的打包 bundle，所以从源码 checkout 添加或刷新 marketplace 前需要先运行 `make package-plugins`：

```bash
make package-plugins
codex plugin marketplace add "$PWD"
codex plugin add memforge@memforge-codex-marketplace
```

日常 CLI smoke 更简单的 fallback 是直接注册 MCP server：

```bash
go install ./cmd/memforge
codex mcp add memforge -- memforge --no-version-check mcp
```

直接 `codex mcp add` 可以避免额外保留一个本地 marketplace entry，但它只是开发/调试 fallback，不是 packaged plugin 的 runtime path。它使用 `PATH` 上的 `memforge` binary；打包后的 Codex plugin bundle 使用 bundled launcher。Codex MCP 配置设置 `default_tools_approval_mode` 为 `approve`，因此非交互 `codex exec` 可以完成 MemForge MCP tool call。

## Release 与 CI 打包

GitHub release workflow 会构建多平台 `memforge` binaries，运行仓库与 harness 校验，用 `tools/package_plugins.sh` 打包 MemForge Claude Code 和 Codex plugin zips，通过 Node launcher smoke bundled Claude runtime，并把 standalone binaries 与 plugin zip assets 上传到 release。

## Standalone CLI 安装

需要独立 `memforge` CLI 的用户，可以用 curl 安装最新 release binary：

```bash
curl -fsSL https://raw.githubusercontent.com/MagnumGOYB/memforge/main/scripts/install.sh | bash
```

安装脚本支持 `MEMFORGE_INSTALL_DIR` 与 `MEMFORGE_VERSION`：

```bash
MEMFORGE_INSTALL_DIR="$HOME/bin" MEMFORGE_VERSION="latest" \
  curl -fsSL https://raw.githubusercontent.com/MagnumGOYB/memforge/main/scripts/install.sh | bash
```

这个 standalone CLI 安装适合直接注册 MCP 或 shell 使用。它不是 Claude Code 或 Codex marketplace/release plugin 安装的前置条件，因为打包插件会使用 bundled runtime binaries。

## 真实 smoke 目标

仓库校验会检查 plugin manifests、launcher 配置、打包和 release workflow 结构。运行时 smoke 应验证：

1. marketplace/release plugin 可以在没有单独 `memforge` binary 位于 `PATH` 的情况下启动。
2. `bin/memforge-mcp-launcher.js` 会为当前平台选择 bundled runtime。
3. bundled runtime 能响应 MCP `tools/list`。
4. `save_memory` 可以持久化 memory。
5. `search_memory` 可以检索到它。
