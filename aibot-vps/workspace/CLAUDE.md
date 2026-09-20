# Where to work

This container runs on a VPS and is only your point of presence. It is deliberately
minimal: no root, no sudo, no package manager rights, almost no tools. Do not use it as a
workstation and do not try to escape to the VPS host.

All real work happens on the Raspberry Pi over SSH:

```bash
ssh raspberry
ssh raspberry '<command>'
```

On the Raspberry you connect as user `teadove` and have full administrative access and
full trust: `sudo`, Docker, systemd, cron, installing packages — all allowed without
asking. It is a disposable host dedicated to your tasks.

`sudo` on the Raspberry asks for a password. It is stored in this container at
`/run/secrets/raspberry_sudo_pass`. Never print or echo the password; feed it to
`sudo -S` over stdin:

```bash
ssh raspberry 'sudo -S -p "" whoami' < /run/secrets/raspberry_sudo_pass
```

For a sequence of root commands, prefer a single `sudo -S` invocation running a script,
or refresh the sudo timestamp first (`sudo -S -v`) and then run plain `sudo` commands
within the same ssh session.

Rules:

- Execute tasks (builds, scripts, services, downloads, experiments) on the Raspberry, not
  here. If a task needs a tool, install it on the Raspberry.
- Keep task files and repositories on the Raspberry. Use `/workspace` here only for notes
  and scratch data that must survive container restarts.
- Long-running things belong on the Raspberry as systemd units, Docker containers, or
  cron jobs — they must not depend on this container being alive.
- SSH hosts are defined in `~/.ssh/config`; more remote hosts may be added there later —
  the same rules apply to them.

# GitHub

A GitHub token for a dedicated bot account is stored in this container at
`/run/secrets/github_token`. Never print or echo it, and never embed it in remote URLs
(it would leak into `.git/config` and shell history). Set it up on the Raspberry once via
the `gh` CLI (install it there first if missing):

```bash
ssh raspberry 'gh auth login --with-token && gh auth setup-git' < /run/secrets/github_token
```

After that, plain `git` and `gh` on the Raspberry are authenticated persistently. If you
commit, set the bot account's `user.name`/`user.email` in git config on the Raspberry
first.
