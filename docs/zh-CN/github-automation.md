# GitHub 自动化

[English](../github-automation.md)

本仓库在 PR、Review、Dependabot auto-merge、Bug 报告、CI 上使用 GitHub 原生自动化。

## Pull Request 流程

- `.github/pull_request_template.md` 强制需求分类、验收标准、测试证据、Validation、回滚与残余风险。
- `.github/workflows/pr.yml` 用 `make validate-pr-body` 校验真实 PR 元数据，并用 `make commitlint-range` 校验 PR 范围内的每条提交信息。
- PR 元数据必须写明发版判定：feature 与 bugfix 必须跟进发版或显式延期；工程/流程类变更标注不需要发版。
- `.github/workflows/ci.yml` 在 push 与 pull request 上跑本地校验与 harness 闸门。

## Review 流程

- `.github/workflows/pr-review.yml` 在 PR 新建或更新时发布 checklist 评论。
- checklist workflow 使用 `issues: write` 与 `pull-requests: write`，让 `actions/github-script` 在分支保护下能创建或更新 PR Issue 评论。
- checklist 提醒 reviewer 检查离线/隐私边界、不污染仓库、Markdown 真值源、CLI 输出稳定性、发版判定、发版影响。
- `.github/CODEOWNERS` 把核心区域路由给维护者，覆盖 `internal/`、`harness/`、`tools/` 与 GitHub workflows。
- 仓库 Review 治理只用 GitHub 原生能力，不引入付费 AI Review 自动化。

## Bugfix 流程

- `.github/ISSUE_TEMPLATE/bug_report.yml` 收集 CLI 命令、操作系统、版本、预期行为、实际行为、脱敏日志、Validation 证据。
- Bugfix PR 条件允许时必须先加失败测试、fixture、harness 传感器或显式手动复现。
- PR 模板强制失败/边界覆盖与回滚说明。

## Dependabot 流程

- `.github/dependabot.yml` 每周打开分组 PR，覆盖 Go modules 与 GitHub Actions。
- `.github/workflows/dependabot-auto-merge.yml` 只对非 major 的 Dependabot 更新启用 GitHub auto-merge。

## CI 流程

- `.github/workflows/ci.yml` 在 Ubuntu 上跑 `make validate` 与 `make test-harness`，使用固定 Go 工具链版本。
- 工作流使用当前 Node 24 主版本 Action：`actions/checkout`、`actions/setup-go`、`actions/github-script`、`actions/upload-artifact`、`actions/download-artifact`、`softprops/action-gh-release`；harness 应验证必要 workflow 能力，而不是固定某个 action major。

## 发版流程

v0.1 暂不引入发版自动化。一旦加入，写入本节并同步 `docs/github-automation.md`。

即便没有发版自动化，发版判定策略依然生效：feature 与 bugfix 必须跟进发版或显式延期；工程/流程类变更标注不需要发版。
