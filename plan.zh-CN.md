# MemForge 编码计划

[English](plan.md)

## 1. 项目目标

MemForge 是为 AI 编码代理提供的本地优先项目记忆层。

核心目标：

- 为 Codex / Claude Code / Cursor / Gemini CLI 提供长期项目记忆
- 不污染用户仓库
- 支持人工记忆持久化与检索
- 为 AI 编码流程编译相关上下文
- 维持低 Token 成本和高检索精度
- 保持完全本地化、人工可控

---

## 2. 产品定位

MemForge 不是：

- 通用知识库
- 聊天记录存档
- 个人 AI 记忆系统
- 通用 RAG 平台

MemForge 是：

> 面向 AI 编码代理的本地项目上下文编译器。

---

## 3. 技术栈

```txt
语言：Go 1.26.3+
CLI：github.com/spf13/cobra
配置：github.com/spf13/viper（文件 + 环境变量 + flag）
数据库：modernc.org/sqlite（纯 Go SQLite + FTS5）
存储：Markdown（真值源）+ SQLite（索引）
ID：github.com/oklog/ulid/v2
Tokenizer：github.com/pkoukk/tiktoken-go（cl100k_base；离线降级：字符数/4）
MCP：stdio MCP Server（v0.3）
LLM Provider：可插拔；OpenAI 兼容 / Anthropic / Ollama（v0.2+）
```

依赖策略：

- 生产代码新增依赖必须在 PR 中给出具体理由，包含二进制体积、离线行为和供应链影响评估。
- 工具、harness、CI 仅允许使用 Go 标准库，除非记录具体理由。

---

## 4. AI 工程治理

本项目沿用 `aitok` 的治理形态，按 memforge 语义去噪。代理必须遵循 `AGENTS.md` 中的协议以及 `docs/zh-CN/harness-engineering.md` 中的 harness/PR 规则。

隐私与范围红线：

- 默认无网络上传、无远程同步、无遥测。命令启动期唯一的网络行为是低频的 GitHub Release 元数据版本检查，必须支持 `--no-version-check` 或 `MEMFORGE_NO_VERSION_CHECK=1` 关闭。
- 记忆只落在用户本地数据目录。CLI 不得向工作目录写入，不得自动扫描、哈希或指纹化仓库文件内容。
- LLM Provider 调用为 opt-in。`init`、`remember`、`search`、`context`、`before` 不得作为副作用触发任何 LLM 请求。只有用户显式调用相应命令时才允许把记忆内容发送给 Provider。
- 生产包不得 import `harness/` 或 `tools/`。

迭代前自我约束流程（对齐 aitok §3，已去噪）：

1. 需求分类——feature / bugfix / refactor / harness-tooling / analysis-only；写明用户可见结果、目标平台、发版判定。
2. 栈适配——优先使用 Go 标准库与现有包边界；新增依赖需要具体理由和影响评估。
3. 验收和回归方案——实现前先写可观测的验收标准；至少覆盖一个失败或边界场景。
4. 范围守护——预先指出预期改动的文件或目录；不做无关重构和格式化。
5. 交付就绪——迭代过程中跑最小有针对性的检查；交付前按 §16 跑回归矩阵。

发版策略：工程或流程性变更不需要发版；feature 和 bugfix 必须进入发版流程或显式延期。

---

## 5. MVP 范围（v0.1）

第一版只关注核心工作流：

```txt
init
remember
search
context
before
reindex
```

v0.1 不包含：

```txt
向量检索
复杂 UI
仓库自动扫描
自动记忆持久化
Git diff 摘要
会话自动抽取
MCP server
LLM Provider 调用
```

---

## 6. 项目结构

```txt
memforge/
  cmd/
    memforge/
      main.go              # CLI 入口，串联 Cobra 命令
  internal/
    cli/                   # init/remember/search/context/before/reindex 各一个文件
    project/               # detect.go、hash.go、meta.go
    memory/                # model.go、markdown.go、store.go、validator.go
    index/                 # sqlite.go、migrations.go、fts.go、search.go
    compiler/              # compiler.go、ranker.go、budget.go、tokenizer.go
    config/                # config.go（Viper 封装，处理优先级）
    buildinfo/             # 通过 -ldflags 注入版本元数据
  tools/
    commitlint/            # Go 原生提交信息校验
    validate-pr-body/      # Go 原生 PR 模板校验
    version/               # 读取 VERSION，校验 tag
  harness/
    architecture_test.go   # 仅做结构与契约传感器
  docs/
    harness-engineering.md
    github-automation.md
    iteration-notes/
    zh-CN/
  .githooks/
    commit-msg
  .github/
    pull_request_template.md
    workflows/
  AGENTS.md、CLAUDE.md、README.md、CONTRIBUTING.md（含 zh-CN 镜像）
  Makefile、VERSION、go.mod、.gitignore
```

---

## 7. 本地存储布局

默认存储位置解析顺序：

```txt
1. MEMFORGE_HOME（环境变量，绝对路径）
2. $XDG_DATA_HOME/memforge
3. ~/.local/share/memforge（macOS 与 Linux 通用兜底）
```

根目录下的布局：

```txt
projects/
  {project_hash}/
    meta.json
    schema_version
    memories/
      manual.md
      constraints.md
      conventions.md
      decisions.md
      bugfixes.md
      api-contracts.md
      agent-instructions.md
    index.sqlite
    cache/
```

项目哈希确定方式：

```txt
1. 优先读取 `git config --get remote.origin.url`，规范化（小写 scheme/host、去凭据、去末尾 `.git`）。
2. 否则使用 git toplevel 绝对路径。
3. 否则使用当前工作目录的规范绝对路径。
4. 对所选标识做 SHA256；小写十六进制；取前 16 位。
```

可选的项目级配置：`.memoryrc`（TOML），可不存在。项目配置只承载代理可见的旋钮（预算、kind 权重、默认 tag），不包含任何记忆内容。

配置优先级：`flag > env (MEMFORGE_*) > 项目 .memoryrc > 用户 $XDG_CONFIG_HOME/memforge/config.toml > 默认值`。

SQLite 运行规则：

- 打开时设置 `PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA busy_timeout=5000`。
- Schema 版本记录在 `schema_version`（单行 TEXT）。每次 CLI 启动检查迁移；`init` 与 `reindex` 必须可重复执行。

---

## 8. 记忆模型

```go
type MemoryKind string

const (
  KindManual           MemoryKind = "manual"
  KindConstraint       MemoryKind = "constraint"
  KindConvention       MemoryKind = "convention"
  KindDecision         MemoryKind = "decision"
  KindBugfix           MemoryKind = "bugfix"
  KindAPIContract      MemoryKind = "api-contract"
  KindAgentInstruction MemoryKind = "agent-instruction"
)
```

Kind 语义（必须落到 `docs/memory-format.md`）：

- `manual`——用户写的高优先自由记忆。
- `constraint`——代理必须遵守的不变量（"X 必须保持框架无关"）。
- `convention`——仓库代码风格约定。
- `decision`——已落定的架构决策。
- `bugfix`——具体故障留下的经验。
- `api-contract`——稳定的接口或 schema 承诺。
- `agent-instruction`——给 AI 代理的操作说明；优先级低于 `manual`，用于流程类提示。

记忆结构：

```go
type Memory struct {
  ID         string     // ULID，可按时间排序，26 字符
  ProjectID  string
  Kind       MemoryKind
  Title      string
  Content    string
  Tags       []string
  Source     string     // 自由文本，例如 "cli"、"session:2026-05-19"、"import"
  Confidence float64    // 0.0–1.0
  UsageCount int
  CreatedAt  time.Time
  UpdatedAt  time.Time
}
```

Markdown 文件格式（同一 kind 一个文件，每条记忆是一个块）：

```md
<!-- memforge:memory id=01HQR... kind=decision -->
---
title: Repository layer must remain framework-agnostic
tags: [architecture, repository]
source: cli
confidence: 1.0
created_at: 2026-05-19T10:00:00Z
updated_at: 2026-05-19T10:00:00Z
---

正文（Markdown）。

<!-- /memforge:memory -->
```

真值源：Markdown 是权威表示。SQLite 可由 `memforge reindex` 完全重建，不能持有 Markdown frontmatter 或正文之外的信息。

写入原子性：`remember` 先写 Markdown 块（fsync 目录），再在同一个 SQL 事务里 upsert 行并触发 FTS 同步。SQLite 失败时下一次 CLI 启动通过重读 Markdown 自愈。

---

## 9. SQLite Schema

```sql
CREATE TABLE schema_version (version TEXT NOT NULL);

CREATE TABLE memories (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  tags_json TEXT NOT NULL DEFAULT '[]',
  tags_flat TEXT NOT NULL DEFAULT '',   -- 与 tags_json 同步的空格分隔镜像，供 FTS 使用
  source TEXT,
  confidence REAL NOT NULL DEFAULT 1.0,
  usage_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX memories_kind_idx ON memories(kind);
CREATE INDEX memories_updated_idx ON memories(updated_at);

CREATE VIRTUAL TABLE memories_fts USING fts5(
  title,
  content,
  tags_flat,
  content='memories',
  content_rowid='rowid',
  tokenize = 'unicode61 remove_diacritics 2'
);

CREATE TRIGGER memories_ai AFTER INSERT ON memories BEGIN
  INSERT INTO memories_fts(rowid, title, content, tags_flat)
  VALUES (new.rowid, new.title, new.content, new.tags_flat);
END;

CREATE TRIGGER memories_ad AFTER DELETE ON memories BEGIN
  INSERT INTO memories_fts(memories_fts, rowid, title, content, tags_flat)
  VALUES ('delete', old.rowid, old.title, old.content, old.tags_flat);
END;

CREATE TRIGGER memories_au AFTER UPDATE ON memories BEGIN
  INSERT INTO memories_fts(memories_fts, rowid, title, content, tags_flat)
  VALUES ('delete', old.rowid, old.title, old.content, old.tags_flat);
  INSERT INTO memories_fts(rowid, title, content, tags_flat)
  VALUES (new.rowid, new.title, new.content, new.tags_flat);
END;
```

Tag 双重持久化：`tags_json` 用于结构化往返，`tags_flat` 是空格分隔的镜像参与 FTS5 评分。每次 upsert 由 `tags_json` 重建 `tags_flat`。

---

## 10. CLI 命令

### 通用 flag

每个命令都支持：

- `--format text|json`（默认 `text`，`json` 是自动化契约）。
- `--no-version-check`，同时尊重 `MEMFORGE_NO_VERSION_CHECK=1`。
- 退出码契约：
  - `0` 成功
  - `1` 用户可见错误（缺项目、校验失败）
  - `2` 调用非法（错误的 flag）
  - `3` 内部错误（文件系统、SQLite）

JSON 输出必须是完整的机器可读对象，stdout 不混入人类可读警告；版本提示等放 stderr。

### init

```bash
memforge init [--root PATH]
```

职责：

```txt
识别当前项目（git root 或 cwd）
计算项目哈希（§7）
创建存储目录与 meta.json
创建 kind 分组的 Markdown 文件（空文件）
打开 SQLite，执行迁移，初始化 FTS5
```

幂等：重复执行 `init` 不得破坏已有记忆，只补齐缺失文件并迁移 schema。

### remember

```bash
memforge remember \
  --kind decision \
  --title "Repository layer is framework-agnostic" \
  --tag architecture --tag repository \
  --source cli \
  --confidence 1.0 \
  "正文，或用 - 从 stdin 读取，或用 --from FILE 从文件读取"
```

内容来源（三选一）：

```txt
位置参数
--from FILE
-          （从 stdin 读取）
```

职责：

```txt
校验 kind、tag、content
生成 ULID
将 Markdown 块追加到对应 kind 文件
在单个事务内 upsert SQLite 行
`--format json` 时输出 {id, kind, title}
```

### search

```bash
memforge search [--kind k1,k2] [--tag t1] [--limit 20] [--format json] QUERY
```

行为：

```txt
带前缀 token 的 FTS5 MATCH
执行排序器（§11）
返回 Title、Kind、Tags、Snippet、Score
```

### context

```bash
memforge context --budget 3000 [--format json] [--kinds k1,k2]
```

职责：

```txt
载入选中 kind 的全部记忆（默认全部）
在 kind 内部排序（§11）
按 kind 权重比例分配 Token 预算
按 kind 分组输出 agent-ready 的 Markdown
```

### before

```bash
memforge before [--budget 3000] [--format json] "任务描述"
```

`before` 是任务条件检索，相对 `context` 的区别：

```txt
1. 用任务描述作为查询执行检索
2. 把任务匹配的记忆与高优先级记忆合并
   （manual + constraint 在子预算内强制保留）
3. 输出 Markdown，以任务作为首个段落
```

### reindex

```bash
memforge reindex
```

从 Markdown 重建 `index.sqlite`。用于写入失败或人工修改 Markdown 后的恢复。

---

## 11. 上下文编译器

目标：

> 在受限 Token 预算下产出价值最高的项目上下文。

Token 计算：tiktoken `cl100k_base`。当 tokenizer 模块不可用时回落到 `utf8.RuneCountInString(content)/4 + 8`，并在 stderr 输出告警。

评分公式（归一化，避免任何单项主导）：

```txt
score =
    0.40 * bm25_norm        // FTS5 BM25，候选集内做 min-max 归一化
  + 0.20 * kind_weight_norm // kind 权重 / 100
  + 0.15 * recency_score    // 0.5 ^ (距 updated_at 的天数 / 30)
  + 0.15 * usage_score      // min(1, log10(1 + usage_count) / 2)
  + 0.10 * confidence       // 0.0–1.0
```

默认 kind 权重（可在 `.memoryrc` 覆盖）：

```txt
manual:            100
constraint:         90
convention:         80
decision:           70
bugfix:             60
api-contract:       55
agent-instruction:  50
```

预算分配：

```txt
1. 计算 per-kind 配额 = (kind_weight / sum_kind_weights) * total_budget
2. 按分数倒序贪心填满配额
3. 未用完的配额回流给下一个权重最高的 kind
4. 在贪心填充前，先把 manual + constraint 按子预算 30% 强制纳入
```

---

## 12. 开发阶段

每个阶段都必须遵循 §4 的自我约束流程和 §16 的回归矩阵。除纯文档迭代外，测试必须先于实现。

### Phase 1 — 项目脚手架

任务：

- 初始化 Go module
- 用 `internal/buildinfo` 串起 Cobra 根命令和 version 子命令
- 用 Viper 实现 §7 的配置优先级
- 激活本轮交付的 Makefile、.githooks 与 harness 传感器

验收：

```bash
make build
./memforge --version
./memforge help
make test-harness
```

### Phase 2 — 项目检测与存储引导

任务：

- 实现 `internal/project`（detect、hash、meta）
- 实现 `internal/index/sqlite.go`（迁移、WAL、busy_timeout）
- 实现 `internal/memory/markdown.go`（frontmatter 解析与写入）
- `memforge init` 写出 `meta.json`、kind 分组 Markdown、`index.sqlite`

验收：

```bash
memforge init
test -f "$(memforge debug paths --format json | jq -r .meta)"
```

### Phase 3 — remember

任务：

- 实现 `--kind`、`--title`、`--tag`、`--source`、`--confidence`、`--from`、stdin
- Markdown 写入 → SQLite upsert 原子化
- JSON 输出契约

验收：

```bash
memforge remember --kind decision --title "..." "Body"
memforge remember --kind constraint --from notes.md
echo "Body via stdin" | memforge remember --kind manual --title "Pipe" -
```

### Phase 4 — search

任务：

- 带前缀 token 的 FTS5 MATCH 构造器
- 排序与摘要生成
- JSON 与文本两种输出

验收：

```bash
memforge search "repository framework"
memforge search --format json --kind constraint,decision "auth"
```

### Phase 5 — context

任务：

- 接入 tokenizer（tiktoken 主路径 + 字符/4 兜底）
- kind 间预算分配器
- 按 kind 分组的 Markdown 上下文输出

验收：

```bash
memforge context --budget 3000
memforge context --budget 1500 --kinds constraint,decision --format json
```

### Phase 6 — before

任务：

- 任务条件检索
- 高优先（manual + constraint）子预算与任务匹配结果合并
- 以任务为首段的 Markdown 输出

验收：

```bash
memforge before "Refactor auth middleware"
```

### Phase 7 — reindex 与恢复

任务：

- 解析所有 Markdown 块并从零重建 SQLite
- 检测并报告孤儿（无 Markdown 块的行）与幽灵（无行的 Markdown 块）

验收：

```bash
rm "$STORAGE/projects/$HASH/index.sqlite"
memforge reindex
memforge search "repository framework"   # 仍然能命中
```

---

## 13. v0.2 路线图

```txt
after 命令（从 session 日志抽取候选记忆）
LLM Provider 抽象（OpenAI 兼容、Anthropic、Ollama）
记忆去重与合并建议
随时间衰减的 confidence
```

示例：

```bash
memforge after --from session.md
```

默认行为：

```txt
生成候选记忆
要求人工确认
持久化已确认记忆
```

LLM 调用为 opt-in。`init`、`remember`、`search`、`context`、`before` 期间永不触发。

---

## 14. v0.3 路线图

加入 MCP server 支持（stdio）。

工具（每个工具必须在 `docs/mcp.md` 发布 JSON Schema）：

```txt
search_memory       { query: string, kinds?: string[], limit?: int, hybrid?: bool }
compile_context     { budget?: int, kinds?: string[] }
save_memory         { kind: string, title: string, content: string, tags?: string[] }
upsert_project_memory { kind: string, title: string, content: string, tags?: string[] }
list_constraints    { limit?: int }
get_project_context { task?: string, budget?: int }
```

代理工作流：

```txt
任务开始   -> compile_context | get_project_context
编码过程   -> search_memory
任务收尾   -> save_memory 保存明确的新记忆，或 upsert_project_memory 维护 agent 选择的稳定记忆
```

MCP server 必须复用 CLI 的相同护栏：无远程网络、不自动扫描、stdout 仅输出 JSON。Enabled plugin 可以让 agent 在线程执行期间自行决定调用 `upsert_project_memory`，但 CLI `after` 仍保持 proposal-first 与人工确认。

---

## 15. v0.4 路线图

```txt
Git diff 摘要
会话抽取适配器（Codex、Claude Code、Cursor）
混合检索（BM25 + 向量）
向量 Provider（本地优先，远程 opt-in）
上下文预算优化器
Claude Code 集成文档
Codex 集成文档
```

---

## 16. 回归验证矩阵

- 仅文档或 harness：`make check`、`make test-harness`；PR 规则变更时追加 `make validate-pr-body`。
- 记忆模型 / Markdown 解析：`make test`，覆盖坏 frontmatter、缺字段、重复 ID、超大内容。
- Index/SQLite：`make test`，覆盖迁移重放、WAL 恢复、FTS 触发器扇出、reindex 正确性。
- 编译器 / 排序器：`make test`，覆盖空仓、单一 kind、超大内容、tokenizer 降级路径。
- CLI：`make test`、`make build`，需要时 `make run ARGS="search ..."` 烟测。
- 依赖、CI 或发布配置：`make validate`，并写明二进制体积/离线/供应链影响。

---

## 17. CI 与提交协议

- CI 必须运行 `make validate` 和 `make test-harness`。
- PR 使用 `.github/pull_request_template.md`，必须包含 Summary、需求分类、验收标准、改动范围、发版判定、TDD / 测试证据、Validation、风险与回滚。
- feature 与 bugfix PR 必须勾选合并后发版或写明用户批准的延期。harness/工具、文档、CI 类 PR 标记不需要发版。
- 提交信息遵循 Go commitlint 格式 `{emoji} {type}{scope}: {subject}`。映射：`✨ feat`、`🐛 fix`、`📝 docs`、`👷 ci`、`💄 style`、`♻️ refactor`、`🔖 release`、`⚡️ perf`、`✅ test`、`🔧 chore`、`🏗️ build`。一次性执行 `make setup` 接入 `.githooks/commit-msg`，或直接 `make commitlint COMMIT_MSG_FILE=<commit-msg-file>`。PR CI 校验 PR 范围内的每一条提交。
- 跳过本地校验后不得声称完成。
- 只提交属于当前迭代的文件，禁止顺带提交无关脏文件。

---

## 18. 开发优先级

```txt
1. Phase 1 — 脚手架与 harness
2. Phase 2 — 项目检测与存储
3. Phase 3 — remember
4. Phase 4 — search
5. Phase 5 — context
6. Phase 6 — before
7. Phase 7 — reindex
8. MCP（v0.3）
9. LLM 摘要（v0.2）
10. 向量检索（v0.4）
```

---

## 19. MVP 验收标准

下列命令端到端可靠运行：

```bash
memforge init
memforge remember --kind decision --title "Repository layer is framework-agnostic" "Body"
memforge search "repository framework"
memforge before "Refactor repository layer"
memforge reindex
```

并且系统具备：

```txt
可靠持久化（Markdown 为真值源，SQLite 可重建）
按 §11 排序器精准召回相关记忆
不污染仓库（存储固定在 MEMFORGE_HOME / XDG_DATA_HOME 下）
在 Token 预算内生成 agent-ready 上下文
删除 index.sqlite 后可通过 reindex 恢复
```

---

## 20. 核心原则

```txt
正确性优先于复杂度
本地优先于云端
人工控制优先于全自动
工程记忆优先于聊天记忆
CLI 优先于 UI
Markdown 真值源优先于派生索引
```
