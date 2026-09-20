# Where to work

This container runs on a VPS and is only your point of presence. It is deliberately
minimal: no root, no sudo, no package manager rights, almost no tools. Do not use it as a
workstation and do not try to escape to the VPS host.

All real work happens on the Raspberry Pi over SSH:

```bash
ssh raspberry
ssh raspberry '<command>'
```

On the Raspberry you have full administrative access and full trust: passwordless `sudo`,
Docker, systemd, cron, installing packages — all allowed without asking. It is a
disposable host dedicated to your tasks.

Rules:

- Execute tasks (builds, scripts, services, downloads, experiments) on the Raspberry, not
  here. If a task needs a tool, install it on the Raspberry.
- Keep task files and repositories on the Raspberry. Use `/workspace` here only for notes
  and scratch data that must survive container restarts.
- Long-running things belong on the Raspberry as systemd units, Docker containers, or
  cron jobs — they must not depend on this container being alive.
- SSH hosts are defined in `~/.ssh/config`; more remote hosts may be added there later —
  the same rules apply to them.
