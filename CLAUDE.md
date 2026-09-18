# Project overview

Personal dotfiles + a small Go CLI that installs them.

## Layout

- `u.go` + `pkg/cli/` — the `u` CLI (built with `github.com/urfave/cli/v3`). Subcommands
  live in `pkg/cli/commands/` (logs, netscan, netserve, git, code, watch, …) plus
  top-level utilities (uuid, sha, md5, locate-by-ip). `u install` (`pkg/cli/install.go`)
  installs the configs.
- `dotfiles-configs/` — the actual config files (`.zshrc`, `.bashrc`, `.tmux.conf`,
  `.config/fish`, `.claude`, …). `u install` copies them into `$HOME`; a few listed in
  `mergeConfigs` (e.g. `.claude/settings.json`) are merged instead of overwritten.
- `extra/` — OS bootstrap scripts (`osx-install.sh`, `ubuntu-install.sh`,
  `data-analytics-install.sh`) and helper `bin/`.
- `devcontainer/` — Docker sandbox (`make cbox-build`).

## Common commands

- `make install` — `go install u.go` then `u install`.
- `make test` — run the Go test suite.
