# AGENTS

[English](AGENTS.md)

此文件是本仓库 AI 编码代理的执行指南。面向用户的过程更新和交付说明默认使用 zh-CN，除非用户另有要求。代码、命令名、测试名、上游文档与引用原文保持 as-is。

## 1) 项目使命

MemForge 是面向 AI 编码代理（Codex、Claude Code、Cursor、Gemini CLI）的本地优先项目记忆层。
它把结构化的项目记忆保存在用户本机，用 SQLite + FTS5 建立索引，并在 Token 预算内编译成 agent-ready 上下文。
虽然 CLI 也面向人类用户，但产品与工程决策必须优先保证 AI 代理与自动化流程的可靠调用。

当前代码库形态：

- Go 1.26.3+
- 单二进制 CLI：`cmd/memforge`
- CLI 命令：`internal/cli`
- 项目检测：`internal/project`
- 记忆模型与 Markdown I/O：`internal/memory`
- SQLite + FTS5 索引：`internal/index`
- 上下文编译器：`internal/compiler`
- 配置（Viper）：`internal/config`
- 构建元数据：`internal/buildinfo`
- Harness 传感器：`harness`

迭代记录：

- 随仓库版本化的迭代交接记录放在 `docs/iteration-notes/`，中文镜像在 `docs/zh-CN/iteration-notes/`。
- 继续推进 plan.md 中任意里程碑前，若存在对应迭代笔记应先阅读。

除非明确要求，否则不在范围内：

- 网络上传、远程同步、遥测或云端统计
- 自动扫描、哈希或指纹化用户源文件
- 在 `init`、`remember`、`search`、`context`、`before`、`reindex` 期间触发 LLM 调用
- 常驻后台服务或守护进程
- 基于向量的检索（计划在 v0.4，opt-in）

## 2) 开发命令

- 格式与静态检查：`make check`
- 启用本地 git hooks：`make setup`
- 全量测试：`make test`
- 指定包测试：`make test-packages PKGS="./internal/index ./internal/compiler"`
- Harness-only 测试：`make test-harness`
- Vet：`make vet`
- 构建：`make build`
- 用项目本地 Go cache 运行 CLI：`make run ARGS="search auth"`
- 完整本地验证：`make validate`
- PR 元数据校验：`make validate-pr-body`
- 提交信息检查：`make commitlint COMMIT_MSG_FILE=<commit-msg-file>`
- 沙箱内 agent session 避免直接使用 `go test`、`go vet`、`go build`、`go run`；统一使用 Makefile 目标，让 Go build/cache 输出固定落在 `.cache/memforge`。指定包测试使用 `make test-packages PKGS="./internal/index ./internal/compiler"`，不要直接运行 raw `go test ./internal/index ./internal/compiler`。

## 3) 迭代前自我约束流程

每次产品或 Harness 迭代在编辑文件前做简短内部 Review：

1. 需求分类
- 把请求归入 feature、bugfix、refactor、harness/tooling、analysis-only。
- 写明用户可见结果、目标平台、影响区域。
- 写明发版判定：工程或流程优化不需要发版；feature 与 bugfix 必须提示或继续进入发版流程，除非用户明确延后。
- 请求歧义时，在动手前写下最稳妥的具体假设；只有当错误假设可能造成数据或产品风险时才追问。

2. 栈适配
- 优先使用 Go 标准库与现有包边界。
- 新增依赖需要具体理由，并说明对二进制体积、离线行为、供应链风险的影响。
- 本地记忆存储必须保持 Markdown 为权威表示。SQLite 是索引，不是真值源。

3. 验收与回归方案
- 实现前写下可观测的验收标准。
- 至少覆盖一个失败或边界场景：缺存储目录、坏 frontmatter、ULID 冲突、超大内容、reindex 后 FTS 触发器重放、跨 fork 的项目哈希碰撞。
- 把每条验收标准映射到单元测试、harness 测试、CLI 烟测、手动证据或显式 N/A。

4. 范围守护
- 列出预期改动的文件或目录。
- 不做无关重构、格式化、依赖升级、发版工作流变动。
- harness、CI、文档与生产代码的职责必须分离。

5. 交付就绪
- 迭代过程中跑最小有针对性的检查。
- 交付前根据改动面跑下方 §6 的验证矩阵。
- 总结残余风险和未能跑完的验证。

## 4) 编码规则

- 记忆只能落在 `MEMFORGE_HOME` 或 `$XDG_DATA_HOME/memforge`（兜底 `~/.local/share/memforge`）。CLI 不得向用户仓库写入。
- Markdown 是记忆的真值源。SQLite 可由 `memforge reindex` 完全重建，不得携带 frontmatter 或正文之外的信息。
- 不引入用量数据网络发送。命令启动期唯一的网络行为是低频 GitHub Release 元数据版本检查，必须支持 `--no-version-check` 或 `MEMFORGE_NO_VERSION_CHECK=1` 关闭。
- LLM Provider 集成是 opt-in。MVP 命令（`init`、`remember`、`search`、`context`、`before`、`reindex`）绝不触发网络 LLM 调用。
- CLI 输出必须稳定；JSON 字段变更需要测试。
- AI 代理把 `--format json` 加 `--no-version-check` 视作主要自动化契约。
- JSON 命令的 stdout 必须保持完整的机器可读 JSON 对象。人类可读警告与版本提示放 stderr 或错误返回。
- 退出码契约：`0` 成功；`1` 用户可见错误；`2` 调用非法；`3` 内部错误。
- Markdown/Table 报表保持可读且可脚本化。
- 生产包不得 import `harness/` 或 `tools/`。
- 项目哈希构造是确定性的，记录在 `internal/project/hash.go`，未升级存储 schema 版本前不得变更。

## 5) Harness 维护规则

- 反复出现的问题源于上下文缺失时，更新 `AGENTS.md` / `AGENTS.zh-CN.md` 或相关文档。
- 问题可以被确定性检测时，新增或精化 `harness` 测试。
- 修改 harness、CI、PR 工作流或校验脚本时，同步更新：
  - `docs/harness-engineering.md`
  - `docs/zh-CN/harness-engineering.md`
  - PR 规则有变时更新 `.github/pull_request_template.md`
- Harness 测试只校验仓库结构、工作流约束、文档与脚本的对齐，不测试产品业务逻辑。
- 文档中专有名词大小写由 `harness/casing_test.go` 配合 `docs/casing-rules.txt` 词表强制执行。向文档引入新专有名词时，须将其正确形式和错误变体加入词表。

## 6) 回归验证矩阵

- 仅文档或 harness：`make check`、`make test-harness`；PR 规则变更时追加 `make validate-pr-body`。
- 记忆模型 / Markdown 解析：`make test`，覆盖坏 frontmatter、缺字段、重复 ID、超大内容。
- Index/SQLite：`make test`，覆盖迁移重放、WAL 恢复、FTS 触发器扇出、reindex 正确性。
- 编译器 / 排序器：`make test`，覆盖空仓、单一 kind、超大内容、tokenizer 降级路径。
- CLI：`make test`、`make build`，需要时 `make run ARGS="search ..."` 烟测。
- 依赖、CI、发布配置：`make validate`，并写明二进制体积/离线/供应链影响。

## 7) CI 与提交协议

- CI 必须运行 `make validate` 和 `make test-harness`。
- PR 必须包含 Summary、需求分类、验收标准、改动范围、发版判定、TDD / 测试证据、Validation、风险与回滚。
- feature 与 bugfix PR 必须标注合并后发版或显式延期；harness/工具、文档、CI 类 PR 标注不需要发版。
- 提交信息遵循 Go commitlint 格式 `{emoji} {type}{scope}: {subject}`，emoji 与 type 严格对应：`✨ feat`、`🐛 fix`、`📝 docs`、`👷 ci`、`💄 style`、`♻️ refactor`、`🔖 release`、`⚡️ perf`、`✅ test`、`🔧 chore`、`🏗️ build`。一次性 `make setup` 接入 `.githooks/commit-msg`，或直接 `make commitlint COMMIT_MSG_FILE=<commit-msg-file>`。PR CI 校验 PR 范围内每一条提交。
- 跳过本地校验后不得声称完成。
- 只提交本次迭代涉及的文件，禁止顺带提交无关脏文件。

## 8) 开源文档与 GitHub 自动化

- 每个公开指南或政策文档都需要 zh-CN 镜像，例如 `README.md` + `README.zh-CN.md`、`CONTRIBUTING.md` + `CONTRIBUTING.zh-CN.md`。
- GitHub PR、Review、自动化工作流的变更同步写入 `docs/github-automation.md` 与 `docs/zh-CN/github-automation.md`。
- 仓库 Review 治理只使用 GitHub 原生能力。除非用户明确要求修改策略，否则不引入付费 AI Review 自动化。
