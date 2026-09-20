# Project overview

Personal dotfiles + a small Go CLI that installs them.

## Layout

- `u.go` + `internal/cli/` — the `u` CLI (built with `github.com/urfave/cli/v3`).
  Subcommands live in `internal/cli/commands/` (logs, netscan, netserve, git, code,
  watch, …) plus top-level utilities (uuid, sha, md5, locate-by-ip). `u install`
  (`internal/cli/install.go`) installs the configs and mirrors `.claude` configs into
  `.codex` (see `codexMappings`).
- `dotfiles-configs/` — the actual config files (`.zshrc`, `.bashrc`, `.tmux.conf`,
  `.config/fish`, `.claude`, …). `u install` copies them into `$HOME`; a few listed in
  `mergeConfigs` (e.g. `.claude/settings.json`) are merged instead of overwritten.
- `extra/` — OS bootstrap scripts (`osx-install.sh`, `ubuntu-install.sh`,
  `data-analytics-install.sh`) and helper `bin/`.
- `devcontainer/` — Docker sandbox (`make cbox-build`).
- `aibot-vps/` — Docker setup for a VPS running Claude Code in Remote Control mode
  (`claude rc`): the container is only a point of presence, real work happens over SSH on
  remote hosts (Raspberry Pi). Claude configs are bind-mounted from
  `dotfiles-configs/.claude`, credentials (SSH key, sudo password, GitHub/OpenAI/Telegram
  tokens) are passed via Docker secrets from the gitignored `secrets/`; agent-facing
  rules live in `aibot-vps/workspace/CLAUDE.md`, bootstrap commands in
  `aibot-vps/README.md`.

## Common commands

- `make install` — `go install u.go` then `u install`.
- `make test` — run the Go test suite.
