# CLAUDE

[English](CLAUDE.md)

@AGENTS.zh-CN.md

# Claude Code 入口

主要指令来源为 `AGENTS.zh-CN.md`。

Claude Code 在本仓库的附加说明：

- 面向用户的过程更新默认使用 zh-CN，除非用户另有要求。
- 常规本地校验路径：`make check`、`make test`、`make build`。
- 当改动不止于小范围本地编辑时，交付前跑 `make validate`。
- 自动化友好行为保持稳定：脚本或代理使用 `--format json` 配合 `--no-version-check`。
- 记忆只存于 `MEMFORGE_HOME` 或 `$XDG_DATA_HOME/memforge`，不得写入用户仓库。
