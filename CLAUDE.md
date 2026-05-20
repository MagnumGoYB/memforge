# CLAUDE

[中文](CLAUDE.zh-CN.md)

@AGENTS.md

# Claude Code entrypoint

Use `AGENTS.md` as the main project instruction source.

Additional Claude Code notes for this repo:

- Prefer zh-CN for user-facing progress updates unless the user asks otherwise.
- Use `make check`, `make test`, and `make build` for the common local validation path.
- Use `make validate` before handoff when the change is broader than a small localized edit.
- Keep automation-friendly behavior stable: prefer `--format json` with `--no-version-check` for scripted or agent usage.
- Memories live under `MEMFORGE_HOME` or `$XDG_DATA_HOME/memforge`; never write into the user's repository.
