# Contributing

[English](CONTRIBUTING.md)

感谢参与 `memforge`。本仓库是面向 AI coding agents 的本地优先 Go CLI 与项目记忆层，因此所有贡献都必须保护离线/隐私边界、稳定的自动化行为，以及“不污染用户仓库”的约束。

## 必跑检查

提交前运行：

```bash
make check
make test
make test-harness
make build
```

当 PR 元数据、CI、workflow 或仓库护栏发生变化时，还要运行：

```bash
make validate-pr-body
```

若变更范围更广，优先运行：

```bash
make validate
```

## 贡献约束

- 每个行为变更都需要测试、harness 覆盖，或明确的手工证据。
- Parser 与 markdown 相关改动在相关场景下必须覆盖 malformed 或 missing input。
- 保持默认离线。
- 除非用户明确要求，不要加入上传、同步、遥测或后台服务。
- 不要把记忆写入用户仓库。
- Markdown 始终是真值源；SQLite 必须能通过 `memforge reindex` 重建。
- CLI 自动化行为必须保持稳定，尤其是 JSON 输出、stderr/stdout 分离与 exit codes。

## 仓库工作流

- 运行一次 `make setup` 以启用仓库 `commit-msg` hook。
- 提交信息必须符合 `make commitlint COMMIT_MSG_FILE=<commit-msg-file>` 校验的仓库格式。
- Pull request 必须完整填写 `.github/pull_request_template.md` 中要求的各个 section。
- Feature 与 bugfix 变更必须标记 merge 后需要发版，或明确记录用户批准的延期。

## 治理文档同步

当修改 harness、CI、PR workflow 或校验脚本时，需同步更新对应治理文档：

- Agent 指令变更时更新 `AGENTS.md` / `AGENTS.zh-CN.md`
- 更新 `docs/harness-engineering.md` / `docs/zh-CN/harness-engineering.md`
- 更新 `docs/github-automation.md` / `docs/zh-CN/github-automation.md`
- 若 PR 元数据规则变化，更新 `.github/pull_request_template.md` / `.github/pull_request_template.zh-CN.md`
