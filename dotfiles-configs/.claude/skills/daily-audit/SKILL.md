---
name: daily-audit
description: "Daily hardware/security/availability check across the VPS container, worker-server and worker-pc, plus a best-effort Claude usage note. Checks: worker-server reachability; homeobs and its dependent services; a worker-pc wake/ping/shutdown cycle; a permissions/security baseline diff on all three hosts; a Prometheus health pass; liveness of every token in /run/secrets. Reports to Telegram via goteleout, then announces the fact of sending with `say` on worker-server. Trigger phrases: 'ежедневный аудит', 'daily audit', '/daily-audit'. Runs unattended on a daily schedule; also runnable on demand."
---

# /daily-audit

## Usage

```
/daily-audit
```

Meant to run once a day, unattended, via a scheduled deployment. Also safe to run on demand — every
check below is non-destructive except the worker-pc wake/shutdown cycle, which is itself designed to
be safe (see Hardware → C).

## Goal

Catch drift and outages on the home infrastructure before the user notices them: a host going
unreachable, a service dying, a permission or listening port that shouldn't be there, a token that
quietly expired. One short daily report beats finding out during an incident.

## State

A baseline file persists the last known-good snapshot between runs, since this skill may run in a
fresh container each time:

```
worker-server:/srv/aibot/daily-audit/baseline.json
```

Read it at the start of the run, diff today's findings against it, then **overwrite it with today's
snapshot at the end of the run regardless of what was found** — the file is "yesterday's state", not an
approved-changes list. The report itself is the audit trail; call out every delta there, prominently
for the security-sensitive fields (sudoers, authorized_keys, group membership), quietly for
expected churn (LAN device count, metric values). If the file doesn't exist yet (first run), create it
without reporting deltas and say so in the report.

## Checks

Run every check even if an earlier one fails — a reachability failure on one host must not skip the
checks on the others. Record failures as findings, not as reasons to stop.

### A. Reachability

`ssh -o ConnectTimeout=8 -o BatchMode=yes worker-server 'echo alive'`. This is the always-on host;
treat unreachable as the top-line finding of the report.

### B. Known services (currently: homeobs)

On worker-server:

```bash
for s in homeobs prometheus prometheus-alertmanager prometheus-blackbox-exporter prometheus-node-exporter grafana-server; do
  systemctl is-active "$s"
done
systemctl --failed --no-legend --no-pager
curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:9105/metrics
curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:3000/api/health
```

Add new hosts' service lists here as they're set up — this check is meant to grow with the
infrastructure, not stay homeobs-only forever.

### C. worker-pc wake / ping / shutdown

Follow the lifecycle rules in the root `CLAUDE.md` (worker-pc section) — do not skip or loosen them
for this job:

1. `ssh -o ConnectTimeout=6 -o BatchMode=yes worker-pc 'echo already-on'`. If it answers, a human may
   be using it — **do not wake or power it off**; just record "already on, skipped the cycle" and move
   on to the next check.
2. If unreachable: send Wake-on-LAN from worker-server (`worker-pc-facts` memory has the MAC), poll
   SSH every 5s for up to 2 minutes.
3. Once up, `ssh worker-pc "ping -c 5 -i 0.3 <worker-server's current LAN IP>"` — read the IP from
   worker-server's own `ip -4 -br addr` that run, don't hardcode it, it has moved before.
4. `ssh worker-pc 'sudo systemctl poweroff'`, then confirm with 3 failed connection attempts ~5s
   apart. Only power off because this run woke it — never because it merely answered a ping.

### D. Permissions / security baseline

On each of: this container, worker-server, worker-pc (worker-pc only if step C woke it — don't wake it
just for this). Gather, then diff against the baseline file:

```bash
id
sudo -n -l 2>&1            # expect "a password is required" on the container; a fixed NOPASSWD list on the others
ss -tlnH 2>/dev/null | awk '{print $4}' | sort -u   # listening ports; container has no ss — note that, don't treat as a failure
wc -l ~/.ssh/authorized_keys 2>&1
```

Known-good baseline as of 2026-10-15 (seed the first baseline.json from this if none exists):

- **container**: uid 1000(claude), no sudo, no `ss`/`netstat` available.
- **worker-server**: uid 1001(teadove-aibot), groups `teadove-aibot,audio,users`, sudo limited to
  `mintdrivers-remove-live-media`, `mintdrivers-load-broadcom-modules`, `mint-refresh-cache`,
  `dpkg_lock_check.sh`, `systemctl restart homeobs`, `systemctl reboot` — nothing else. Listening:
  `22` (0.0.0.0, expected — SSH), `631` (loopback, cups), `9090/9093/9100/9105/9115` (loopback,
  Prometheus stack), `53` (loopback, local resolver), `3000` (**all interfaces** — Grafana, expected,
  it's the one LAN-visible service). 2 authorized_keys entries.
- **worker-pc**: uid 1001(teadove-aibot), groups `teadove-aibot,users`, sudo limited to the same first
  four plus `systemctl poweroff`, `systemctl reboot`. Listening: `22` (0.0.0.0), `631` (loopback),
  `53` (loopback) — nothing else; it has no always-on services. 1 authorized_keys entry.

Flag as a finding, not a silent baseline update, anything in: sudo -l output, authorized_keys count,
group membership, or a newly-listening port on a non-loopback address. ufw/iptables rules are **not**
checkable with current permissions on any host — don't claim they were checked; note the gap once in
the report rather than every day (see Known gaps).

### E. Prometheus health

On worker-server, against `127.0.0.1:9090`:

```bash
curl -s 127.0.0.1:9090/api/v1/targets | jq -r '.data.activeTargets[] | select(.health!="up") | .labels.job'
curl -s 127.0.0.1:9090/api/v1/alerts | jq -r '.data.alerts[] | select(.state=="firing") | .labels.alertname'
```

Real-time anomalies already page the user instantly via Alertmanager → homeobs → goteleout — this
check is a complementary daily sanity pass (every target up, nothing silently firing), not a
replacement for it. Don't re-derive "is this value normal" from first principles each day; report
target/alert counts and any `health != "up"` or `state == "firing"` entries, nothing more speculative.

### F. Token liveness

Run from the container, which holds `/run/secrets/*` (never print the values):

| Token | Check |
|---|---|
| `github_token` | `ssh worker-server '~/bin/gh auth status'` (already configured there) |
| `openai_api_key` | `curl -s -o /dev/null -w '%{http_code}' https://teadove.space:7999/proxy/openai/v1/models -H "Authorization: Bearer $(cat /run/secrets/openai_api_key)"` — expect `200` |
| `yandex_disk_token` | `uv run --with "yadisk[sync-defaults]" python <script>` on worker-server calling `client.check_token()` (see `yandex-disk-access` memory for the exact invocation) |
| `telegram_bot_token` | No separate probe — this report's own successful delivery via goteleout **is** the liveness check. If delivery fails, that's the finding (see Output). |

## Output

Format: a short Telegram message via `goteleout` from worker-server (same pattern as every other
notification in this project) — a few lines per section above, problems first, then a one-line
all-clear summary for whatever had no findings. Keep it scannable, not a wall of raw command output.

### Claude usage (best effort)

If a `<total_tokens left>` style system signal is present in the run, report it as a rough session
budget signal. There is no API for personal Claude plan usage/limits, and the Admin API's usage/cost
endpoints need an Admin API key this project doesn't have — say so plainly rather than silently
omitting the section every time; don't speculate about remaining quota beyond what's actually
observable.

### Audio

Last step, every run, unattended:

```bash
ssh worker-server '~/bin/say "Daily audit report sent to Telegram"'
```

This is a standing, explicitly user-approved exception to the "speaker only on request" rule (see that
memory) — scoped to exactly this one call in this one skill. Do not add `say` anywhere else without a
fresh explicit ask in that request.

## Known gaps

State these once per report when relevant, don't pad the report with them daily if nothing changed:

- No firewall (ufw/iptables) visibility on any host with current permissions.
- No read access to the system journal (only this skill's own service-user-scoped journal entries are
  visible) — deeper incident diagnosis after a finding here still needs a human to grant
  `systemd-journal` group membership or do it themselves.
- Claude usage/cost reporting is best-effort only (see Output → Claude usage).
