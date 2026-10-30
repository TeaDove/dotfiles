# Project overview

Personal dotfiles + a small Go CLI that installs them.

## Layout

- `u.go` + `internal/cli/` — the `u` CLI (built with `github.com/urfave/cli/v3`).
  Subcommands live in `internal/cli/commands/` (logs, netscan, netserve, git, code,
  watch, …) plus top-level utilities (uuid, sha, md5, locate-by-ip). `u install`
  (`internal/cli/install.go`) installs the configs.
- `dotfiles-configs/` — the actual config files (`.zshrc`, `.bashrc`, `.tmux.conf`,
  `.config/fish`, `.codex`, …). `u install` copies them into `$HOME`; files listed in
  `mergeConfigs` are merged instead of overwritten.
- `extra/` — OS bootstrap scripts (`osx-install.sh`, `ubuntu-install.sh`,
  `data-analytics-install.sh`) and helper `bin/`.
- `devcontainer/` — Docker sandbox (`make cbox-build`).
- `aibot-vps/` — Docker setup for a VPS running Codex Remote Control
  (`codex remote-control --no-daemon`): the container is only a point of presence, real work happens over SSH on
  remote hosts.

## Common commands

- `make install` — `go install u.go` then `u install`.
- `make test` — run the Go test suite.

## After changes

- After any change to the configs in `dotfiles-configs/` or to the Go code, run `make install` so the installed
  `u` binary and the configs in `$HOME` match the repo.
