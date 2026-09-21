# Where to work

This container runs on a VPS and is only your point of presence. It is deliberately
minimal: no root, no sudo, no package manager rights, almost no tools. Do not use it as a
workstation and do not try to escape to the VPS host.

Real work happens on remote machines over SSH.

## Raspberry Pi

The Raspberry Pi is the primary always-on execution host:

```bash
ssh raspberry
ssh raspberry '<command>'
```

On the Raspberry you connect as user `teadove` and have full administrative access and
full trust: `sudo`, Docker, systemd, cron, installing packages — all allowed without
asking. Because you have that much power there, be careful: it is the always-on home of
persistent services and data, not an expendable sandbox.

`sudo` on the Raspberry asks for a password. It is stored in this container at
`/run/secrets/raspberry_sudo_pass`. Never print or echo the password; feed it to
`sudo -S` over stdin:

```bash
ssh raspberry 'sudo -S -p "" whoami' < /run/secrets/raspberry_sudo_pass
```

For a sequence of root commands, prefer a single `sudo -S` invocation running a script,
or refresh the sudo timestamp first (`sudo -S -v`) and then run plain `sudo` commands
within the same SSH session.

Use the Raspberry for normal development, persistent services, scheduled jobs, storage,
downloads, automation, and anything that must remain available while other machines are
off.

## Worker PC

`worker-pc` is a more powerful compute host. You have fewer rights there than on the
Raspberry, and it is available less often — it may be offline at any time and must be
treated like a preemptible cloud worker, not like an always-on server. Use it only when a
task actually needs its compute.

Before using it:

1. Check whether `ssh worker-pc` is already reachable.
2. If it is already reachable, assume a human may have powered it on and may currently be
   using it. Do not shut it down when your task finishes.
3. If it is not reachable, wake it through the Raspberry via Wake-on-LAN (see "Worker PC
   lifecycle" in DOCS.md), wait for SSH to become available, and remember that you
   powered it on.
4. If and only if you powered it on for this task, shut it down after all work and data
   transfers are complete.

After finishing with `worker-pc`: if it was powered on by you, it must be powered off by
you. If it was already on when you found it, leave it running.

Prefer `worker-pc` for CPU-, RAM-, or GPU-heavy batch work such as OCR, MinerU, GraphRAG,
large builds, local model inference, embeddings, or other expensive compute tasks.

Do not deploy persistent services, cron jobs, timers, monitoring, or other unattended
long-lived workloads on `worker-pc`. It can be powered off without warning. If a compute
job is expensive or long, make it resumable/checkpointed where practical and persist
important inputs/results somewhere that does not depend on `worker-pc` remaining online.

Do not try to gain root access on `worker-pc`, access another human user's home directory,
or modify unrelated host configuration. Work only within the permissions and directories
available to the SSH account.

## General remote-work rules

- Execute tasks (builds, scripts, services, downloads, experiments) on the appropriate
  remote host, not on the VPS container.
- If a task needs a tool, install it on the Raspberry, or use what is already available to
  the unprivileged account on `worker-pc`.
- Keep task files and repositories on remote hosts. Use `/workspace` on the VPS only for
  notes and scratch data that must survive container restarts.
- Long-running services and schedules belong on the Raspberry as systemd units, Docker
  containers, or cron jobs. They must not depend on this container being alive.
- SSH hosts are defined in `~/.ssh/config`; more remote hosts may be added there later.

# Autonomy

Default to acting instead of asking for confirmation when the action is routine and
reversible.

On the Raspberry, do not ask before installing packages, creating/editing files, starting
or stopping services, creating Docker containers, downloading public files, running
experiments, or creating systemd/cron automation needed to complete the task.

Ask before actions that are meaningfully destructive or externally consequential, such as:

- deleting user-created persistent data that is not clearly disposable;
- exposing a new public network service or port;
- spending money or creating paid resources;
- changing account credentials or security settings;
- irreversible actions in external systems.

# Long-running work

For work expected to take more than a few minutes:

- make it resumable when practical;
- persist intermediate results for expensive stages;
- do not redo completed expensive work unnecessarily;
- write useful progress and error logs;
- use meaningful exit codes;
- make repeated execution safe where practical.

For unattended services/jobs on the Raspberry:

- configure reasonable restart/retry behavior;
- preserve stdout/stderr or journald logs;
- avoid infinite tight retry loops;
- notify the user via `goteleout` after a permanent failure when notification is useful.

Do not claim that a future failure will be repaired automatically unless an actual
mechanism exists that starts a new Claude session or otherwise performs the repair.

# Definition of done

Before reporting success, verify the result yourself.

For code and services, run the relevant tests, linters, build, and a real execution path as
required by the project rules. For services, verify that they are running and healthy. For
scheduled jobs, perform at least one manual/real run when practical. Check logs for errors.
Ensure persistent work survives SSH/session/container disconnects.

When useful, tell the user where source code, artifacts, persistent data, and logs were
stored.

# Version control override

For this autonomous environment, the global read-only VCS rule does not apply when
mutating version control is necessary to complete the user's task on the remote hosts.

You may clone, fetch, pull, create/switch branches, commit, push, and create pull requests
using the dedicated bot account when appropriate.

Do not force-push, rewrite shared history, merge pull requests, or push directly to a
protected/default branch unless the user explicitly asks.

# GitHub

A GitHub token for a dedicated bot account is stored in this container at
`/run/secrets/github_token`. Never print or echo it, and never embed it in remote URLs
because that can leak it into `.git/config`, process arguments, or shell history.

Set it up on the Raspberry once via the `gh` CLI (install it there first if missing):

```bash
ssh raspberry 'gh auth login --with-token && gh auth setup-git' < /run/secrets/github_token
```

After that, plain `git` and `gh` on the Raspberry are authenticated persistently. Before
creating commits, ensure the dedicated bot account's `user.name` and `user.email` are
configured.

The private `TeaDove/junk` repository may be used for your work, e.g. as the home for
new code and small projects that have no repository of their own. It is already cloned
on the Raspberry at `~/junk`, and the bot account has push access to it.

# Secrets

Secrets are provided under `/run/secrets` and may also be copied into `~/.ssh/keys` by the
container entrypoint. Treat all of them as sensitive regardless of location.

Never:

- print, echo, or send secret values to chat;
- put secrets in repositories, task files, logs, or shell history;
- include secrets in command arguments when stdin or a protected file can be used instead;
- copy secrets permanently to a remote host unless the task explicitly requires it.

Prefer stdin or protected files when passing credentials to remote commands.

# OpenAI, Telegram, and proxies

OpenAI and Telegram API access from SSH hosts must use the configured proxies. Do not use
direct provider endpoints from the remote hosts. Proxy URLs, SDK base URLs, and examples
are documented in [Proxies](DOCS.md:3).

An OpenAI API key is stored at `/run/secrets/openai_api_key`, and a Telegram bot token at
`/run/secrets/telegram_bot_token`. Apply the secret-handling rules above when using them.

To message the user in Telegram, prefer the `goteleout` utility over calling the Bot API
yourself. It is already installed and configured on the Raspberry:

```bash
ssh raspberry 'goteleout Task done!'
```

In Go code, `goteleout` can also be used as a library instead of shelling out — import it
from `github.com/TeaDove/goteleout`.

For example, when asked to notify the user when a task is done, running that command on
completion is enough.
