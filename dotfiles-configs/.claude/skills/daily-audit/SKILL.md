---
name: daily-audit
description: "Daily hardware/security/availability check across the VPS container, worker-server and worker-pc. Checks: worker-server reachability; homeobs and its dependent services; a worker-pc wake/ping/shutdown cycle; a permissions/security baseline diff on all three hosts; a Prometheus health pass; liveness of every token in /run/secrets; a Grafana dashboard screenshot, split into readable chunks, uploaded to Yandex Disk. Every anomaly (reboot, outage, failed wake, failed step) is recorded, diagnosed and given a proposed fix. Reports to Telegram via goteleout with the screenshot links (falling back to curl from the VPS container if worker-server is unreachable), then announces the fact of sending with `say` on worker-server. Trigger phrases: 'ежедневный аудит', 'daily audit', '/daily-audit'. Runs unattended on a daily schedule; also runnable on demand."
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
treat unreachable as the top-line finding of the report. Retry twice ~30s apart before calling it down, and
also probe worker-pc's port: if both time out, the whole home WAN path is down, not one host.

Being reachable now is not enough. worker-server is a headless server (lid closed, next to the router) and
must never reboot on its own, so also run `journalctl --list-boots --no-pager` and `uptime -s`: any boot
that started after the previous baseline's `generated_at` is an anomaly (see Anomalies), even if every
other check is green.

### B. Known services

Don't hardcode a service list here — it will drift the moment something is added or removed. Discover
it fresh every run, on worker-server:

```bash
systemctl --failed --no-legend --no-pager                                  # any failed unit, unconditional
systemctl list-units --type=service --state=running --no-legend --no-pager # the current active set
```

Diff the running-service set against the baseline file (see State) the same way as the permissions
baseline below — flag additions/removals as findings instead of silently absorbing them, then let the
baseline advance. For whichever services turn out to be running, check their own documented health
endpoint if they have one — as of 2026-10-15 that's homeobs (`curl 127.0.0.1:9105/metrics`, see its
README) and Grafana (`curl 127.0.0.1:3000/api/health`); read `/srv/aibot/*/README.md` or the project's
own docs for any service added later rather than guessing at a URL.

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
ss -tlnH 2>/dev/null | awk '{print $4}' | sort -u   # listening ports; container has no ss, note that
wc -l ~/.ssh/authorized_keys 2>&1
sudo -n ufw status verbose 2>&1   # worker-server only (NOPASSWD since 2026-10-15); skip elsewhere, not a failure
```

Known-good baseline as of 2026-10-15 (seed the first baseline.json from this if none exists):

- **container**: uid 1000(claude), no sudo, no `ss`/`netstat` available.
- **worker-server**: uid 1001(teadove-aibot), groups `teadove-aibot,audio,users`, sudo limited to
  `mintdrivers-remove-live-media`, `mintdrivers-load-broadcom-modules`, `mint-refresh-cache`,
  `dpkg_lock_check.sh`, `systemctl restart homeobs`, `systemctl reboot`, `ufw status[ verbose|
  numbered]` — nothing else. Listening: `22` (0.0.0.0, expected — SSH), `631` (loopback, cups),
  `9090/9093/9100/9105/9115` (loopback, Prometheus stack), `53` (loopback, local resolver), `3000`
  (**all interfaces** — Grafana, expected, it's the one LAN-visible service). 2 authorized_keys
  entries. `ufw`: active, default deny (incoming) / allow (outgoing), only `22` and `3000` allowed in
  (IPv4 and IPv6) — matches the listening-port baseline above exactly; any other allowed port or a
  non-active status is a finding.
- **worker-pc**: uid 1001(teadove-aibot), groups `teadove-aibot,users`, sudo limited to the same first
  four plus `systemctl poweroff`, `systemctl reboot`. Listening: `22` (0.0.0.0), `631` (loopback),
  `53` (loopback) — nothing else; it has no always-on services. 1 authorized_keys entry.

Flag as a finding, not a silent baseline update, anything in: sudo -l output, authorized_keys count,
group membership, a newly-listening port on a non-loopback address, or a `ufw` allowed-port/status
change on worker-server. `ufw`/iptables are **not** checkable on the container or worker-pc with
current permissions — don't claim they were checked there; note that gap once in the report rather
than every day (see Known gaps).

### E. Prometheus health and trend anomalies

On worker-server, against `127.0.0.1:9090`:

```bash
curl -s 127.0.0.1:9090/api/v1/targets | jq -r '.data.activeTargets[] | select(.health!="up") | .labels.job'
curl -s 127.0.0.1:9090/api/v1/alerts | jq -r '.data.alerts[] | select(.state=="firing") | .labels.alertname'
```

Real-time threshold alerts already page the user instantly via Alertmanager → homeobs → goteleout — the
two queries above are a complementary daily sanity pass (every target up, nothing silently firing), not
a replacement for it.

Beyond that, catch drift that has no alert rule yet by comparing today against a trailing week — one
concrete query per metric, each with its own threshold. Don't freelance "does this look normal"; only
report what one of these actually crosses, via `/api/v1/query`:

```promql
# LAN devices online vs. the 7-day average — flag a swing past 30% either way
abs(homeobs_lan_devices_online - avg_over_time(homeobs_lan_devices_online[7d]))
  / avg_over_time(homeobs_lan_devices_online[7d]) > 0.3

# today's best download speed vs. the 7-day average — flag a drop past 30%
(avg_over_time(homeobs_speedtest_download_bits_per_second[7d])
    - max_over_time(homeobs_speedtest_download_bits_per_second[1d]))
  / avg_over_time(homeobs_speedtest_download_bits_per_second[7d]) > 0.3

# root filesystem usage growth over the last 7 days, in percentage points — flag past 5pp/week
(1 - node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})
  - (1 - (node_filesystem_avail_bytes{mountpoint="/"} offset 7d)
    / (node_filesystem_size_bytes{mountpoint="/"} offset 7d)) > 0.05
```

If a series has under 7 days of history (e.g. right after a new metric is deployed), skip it rather
than reporting noise from a short window — note that once in Known gaps, not every day. Add a new
query here, same shape, whenever a metric worth trend-watching is added — this list is meant to grow
with the infrastructure like the rest of the skill, not stay fixed at three forever.

### F. Token liveness

`ls /run/secrets/` from the container to see what actually exists today — don't hardcode the set of
token names here, it drifts the moment a secret is added or renamed. For each one found, work out the
cheapest read-only liveness probe the same way a human would: `/workspace/DOCS.md` documents the
OpenAI/Telegram proxies, and the project's own memory files cover the rest (search for the secret's
name — e.g. the Yandex Disk and GitHub entries each name their exact check). Never print a secret's
value, only whether it still gets accepted.

Two cases don't need a dedicated probe because something else already proves them each run: a token
already wired into a configured CLI (ask that CLI's own auth-status command rather than hitting the
API by hand), and the Telegram bot token specifically — this report's own successful delivery via
goteleout **is** the liveness check; if delivery fails, that's the finding (see Output).

## Anomalies

An anomaly is anything that deviates from normal operation, even if it already recovered by the time
you look: worker-server rebooted or was down, the home network was unreachable, worker-pc did not wake
on WoL or did not power off, a service restarted, a check timed out, a step of this skill failed.
"It came back later" never closes one. For every anomaly:

1. **Record** it in the report (see Output), with timestamps.
2. **Diagnose** it before writing the report, as far as current permissions allow. Starting points:
   - unexpected reboot or downtime on worker-server: the tail of the previous boot
     (`journalctl -b -1 --no-pager | tail -100`, plus `-k` for the kernel log), whether it ended with a
     clean `shutdown.target` or just stops (freeze/power loss), `last -x reboot shutdown`,
     `/var/log/apt/history.log` and `/var/log/dpkg.log` around that time, battery/AC state via `upower -d`,
     thermals via `sensors` if installed;
   - home network unreachable: whether both forwarded ports time out (WAN/router) or only one (that
     host), and the gap in homeobs speedtest/LAN-scan logs and metrics for the same window;
   - worker-pc not waking: whether the magic packet left worker-server, whether worker-pc shows up in the
     LAN scan at all, and how long it was given.
3. **Propose a fix**: a concrete next step, with the exact command when one exists. If the fix needs
   privileges this account lacks, say so and give the command the user must run themselves.

If the diagnosis is incomplete (for example, no access to the system journal), say exactly what was not
visible instead of guessing a cause.

## Dashboard screenshot

Every run, after all the checks above, screenshot the `Home observability` dashboard **split into
several readable chunks** (not one tall strip — confirmed unreadable in Telegram's preview, fixed
2026-10-16) and attach the links to the report (user's explicit ask, 2026-10-16):

1. `scp /run/secrets/grafana_viewer_token worker-server:/tmp/grafana_viewer_token`.
2. On worker-server: `uv run --with playwright python
   /srv/aibot/daily-audit/screenshot.py /tmp/grafana_dashboard < /tmp/grafana_viewer_token`, then
   `rm -f /tmp/grafana_viewer_token` immediately — never leave the token sitting in `/tmp`.
   `screenshot.py` (kept in sync with `dotfiles-configs/.claude/skills/daily-audit/screenshot.py` in
   this repo — if they've drifted, the deployed copy wins for today's run, but fix the repo copy too)
   prints one path per line, e.g. `/tmp/grafana_dashboard-1.png` .. `-N.png`. It scrolls the full page
   first — Grafana lazy-renders panels outside the current viewport, a screenshot taken without this
   step comes back with most panels empty — then reads each row's actual Y position from the DOM
   (`[data-testid*="dashboard-row"]`) and groups rows into chunks of roughly 2000px, cutting only at
   row boundaries so no panel is ever split across two images. Don't hardcode specific row names or
   pixel offsets; the grouping is computed fresh each run so it keeps working as panels are added.
3. Upload each chunk with its own ssh+upload call — **do not loop a single `ssh ... < token` over
   multiple files**, the remote `sys.stdin.read()` only gets data on the first loop iteration and
   every later one fails with "token is empty" (hit this 2026-10-16). One `ssh` invocation per file:
   `ssh worker-server 'uv run --no-project --with "yadisk[sync-defaults]" python
   /srv/aibot/mam-26-1/lab2/upload_file.py /tmp/grafana_dashboard-N.png
   "app:/daily-audit-dashboard-N.png"' < /run/secrets/yandex_disk_token` — the redirect must be on
   the outer `ssh` invocation (the container's own stdin), not inside the quoted remote command, or
   the remote shell looks for the secret on worker-server's own filesystem and fails. `overwrite=True`
   in the helper keeps each chunk's link stable across days. If a chunk's printed URL is empty, the
   publish call raced the upload — re-run just the `get_meta`/`publish` pair for that one file once
   before giving up on it (hit this 2026-10-16, a second attempt returned the link immediately). If a
   chunk still has no link after that retry, report it as a WARN line in the output (see Output) —
   don't silently drop it.
4. Clean up the local `/tmp/grafana_dashboard-*.png` files on worker-server once uploaded.

## Output

Send via `goteleout` — now safe to use again (2026-10-16: PATH fixed so it resolves bare, and
`goteleout` itself patched upstream with a 90s internal deadline + longer per-attempt timeouts,
TeaDove/goteleout#4, so a stuck send fails cleanly instead of hanging forever). Write the report to
`/tmp/daily-audit-report.txt` on worker-server first — piping a multi-line heredoc straight into a
remote `"$(cat)"` inside one SSH command has been unreliable here before, the file round-trip is the
version that's actually proven to work — then:

```bash
ssh worker-server 'goteleout --parse-mode HTML "$(cat /tmp/daily-audit-report.txt)" < /dev/null'
ssh worker-server 'rm -f /tmp/daily-audit-report.txt'
```

Keep the `< /dev/null`: goteleout reads stdin to EOF whenever it is not a terminal, before its own
deadline starts, so over a non-interactive `ssh` without it the call waits forever. That was the cause of
the 2026-10-16 "hangs" (confirmed 2026-10-17: same send, 100s timeout without it, 0.5s with it).

Wrap the call in `timeout 100` as a safety margin anyway (slightly past goteleout's own 90s internal
deadline) and retry once if it still fails. A successful send prints nothing, so don't pipe the call
through `grep -v`: an empty output makes `grep` exit 1 and looks like a failure. If both attempts fail,
use the fallback below.

### Fallback: send from the VPS container

If worker-server is unreachable, or goteleout failed twice, send the same text straight from the
container with curl to `api.telegram.org` (it does not go through worker-server, the home network, or
the proxy). Keep the token out of argv by feeding the URL through `curl -K -`. The chat id is the user's
Telegram id from the `telegram-user-and-goteleout` memory; don't write it into this file.

```bash
{ printf 'url = "https://api.telegram.org/bot%s/sendMessage"\n' "$(cat /run/secrets/telegram_bot_token)"; } \
  | curl -s -m 30 -K - -d chat_id="$CHAT_ID" -d parse_mode=HTML --data-urlencode text@/tmp/daily-audit-report.txt \
  | jq -r '.ok'
rm -f /tmp/daily-audit-report.txt
```

`true` means delivered. If it also fails, the run still ends, and the next scheduled fire tries again.

Send one Telegram message this way, in this fixed shape: one emoji at the very top, line-level emoji
dropped everywhere else, the body split into OK/WARN blocks, and every check written as a bold title line
followed by its detail line. Send it with Telegram's HTML parse mode (`goteleout --parse-mode HTML`, or
`-d parse_mode=HTML` in the curl fallback) and write titles as `<b>Title</b>`. Escape `&`, `<` and `>` in
detail text as `&amp;`, `&lt;` and `&gt;`, otherwise Telegram rejects the whole message. Don't switch to
Markdown: an unescaped `_` in names like `yandex_disk` breaks it.

```
{{ ✅ | ⚠️ }} Daily audit — {{ YYYY-MM-DD }}

✅ OK
<b>{{ check name }}</b>
{{ detail }}

<b>{{ check name }}</b>
{{ detail }}

⚠️ WARN
<b>{{ check name }}</b>
{{ detail }}

🔎 Anomalies
<b>{{ what happened }}</b>
{{ when, and for how long }}
Diagnosis: {{ diagnosis }}
Fix: {{ proposed fix }}
```

The `🔎 Anomalies` block is omitted when there were none; when present, the top emoji is `⚠️`. Each anomaly
is a bold title plus up to three short lines (when, diagnosis, fix), not a log dump.

Each of the seven checks (Reachability, Services, worker-pc, Security baseline, Prometheus, Tokens,
Dashboard) contributes exactly one entry: its name as the bold title, then one short detail line that
starts with a capital letter. The entry goes into whichever block matches its own result that run, never
both. Separate entries with one blank line. Both blocks are optional: omit the `✅ OK` block entirely if
nothing came back clean, omit `⚠️ WARN` entirely if nothing did. The top-line emoji follows directly from
that: `⚠️` if the WARN block exists at all, `✅` only when it's empty and omitted. If worker-server itself
is unreachable, every other check is "unknown" rather than skipped: put all seven entries in WARN that
run rather than guessing any of them clean. Keep each detail line short: this is a status scan, not a
dump of raw command output. Examples of entries in each block:

```
✅ OK
<b>Reachability</b>
Worker-server alive, same boot as baseline (up since 16 Oct 12:32 MSK), no reboot

<b>Services</b>
37 running, 0 failed, set unchanged; homeobs metrics and Grafana health OK

<b>Dashboard</b>
3 parts: 1/3 {{ link }}, 2/3 {{ link }}, 3/3 {{ link }}

⚠️ WARN
<b>worker-pc</b>
Not reachable, ssh timeout, WoL ignored

<b>Tokens</b>
Openai, github, telegram live; yandex_disk not checked, the classifier denied the probe

<b>Prometheus</b>
17/18 targets up (blackbox_exporter down); LAN devices 30% below the 7-day average

🔎 Anomalies
<b>worker-server rebooted</b>
Previous boot ended 23:56 with no clean shutdown, down 12.5h
Diagnosis: journal stops mid-stream right after an apt upgrade (systemd, kernel), so a freeze
Fix: read the kernel log of boot -1; hold kernel upgrades until then
```

### Audio

Last step, every run, unattended:

```bash
ssh worker-server '~/bin/say "Daily audit report sent to Telegram"'
```

Skip it if worker-server is unreachable (the fallback delivery already happened from the container).
This is a standing, explicitly user-approved exception to the "speaker only on request" rule (see that
memory) — scoped to exactly this one call in this one skill. Do not add `say` anywhere else without a
fresh explicit ask in that request.

## Known gaps

Internal reference only — these are not surfaced in the Telegram report (see Output), they just save
the next session from re-discovering the same structural limits:

- No firewall visibility on the container or worker-pc with current permissions (worker-server's
  `ufw status` is checked — see Permissions / security baseline).
- No read access to the system journal (only this skill's own service-user-scoped journal entries are
  visible) — deeper incident diagnosis after a finding here still needs a human to grant
  `systemd-journal` group membership or do it themselves.
