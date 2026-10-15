---
name: daily-audit
description: "Daily hardware/security/availability check across the VPS container, worker-server and worker-pc. Checks: worker-server reachability; homeobs and its dependent services; a worker-pc wake/ping/shutdown cycle; a permissions/security baseline diff on all three hosts; a Prometheus health pass; liveness of every token in /run/secrets; a full-page Grafana dashboard screenshot uploaded to Yandex Disk. Reports to Telegram via goteleout with the screenshot link, then announces the fact of sending with `say` on worker-server. Trigger phrases: 'ежедневный аудит', 'daily audit', '/daily-audit'. Runs unattended on a daily schedule; also runnable on demand."
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

## Dashboard screenshot

Every run, after all the checks above, take a full-page screenshot of the `Home observability`
dashboard and attach its link to the report (user's explicit ask, 2026-10-16):

1. `scp /run/secrets/grafana_viewer_token worker-server:/tmp/grafana_viewer_token`.
2. On worker-server: `~/.local/bin/uv run --with playwright python
   /srv/aibot/daily-audit/screenshot.py /tmp/grafana_dashboard.png < /tmp/grafana_viewer_token`, then
   `rm -f /tmp/grafana_viewer_token` immediately — never leave the token sitting in `/tmp`.
   `screenshot.py` (kept in sync with `dotfiles-configs/.claude/skills/daily-audit/screenshot.py` in
   this repo — if they've drifted, the deployed copy wins for today's run, but fix the repo copy too)
   sets the viewport to the page's own full scroll height and scrolls through it in steps before
   shooting — Grafana lazy-renders panels outside the current viewport, so a naive `full_page`
   screenshot at a fixed small viewport comes back with most panels empty.
3. Upload and publish, reusing the existing helper so the link stays stable across days
   (`overwrite=True` keeps the same public URL — see `yandex-disk-access` memory):
   `ssh worker-server '~/.local/bin/uv run --no-project --with "yadisk[sync-defaults]" python
   /srv/aibot/mam-26-1/lab2/upload_file.py /tmp/grafana_dashboard.png
   "app:/daily-audit-dashboard.png"' < /run/secrets/yandex_disk_token` — the redirect must be on the
   outer `ssh` invocation (the container's own stdin), not inside the quoted remote command, or the
   remote shell looks for the secret on worker-server's own filesystem and fails.
4. Clean up `/tmp/grafana_dashboard.png` on worker-server once uploaded.

## Output

`goteleout` is at `~/go/bin/goteleout` on worker-server, `uv` at `~/.local/bin/uv` — a bare `ssh
worker-server 'goteleout ...'` fails with "command not found", the non-login SSH shell doesn't have
either on `$PATH`; always use the full path. Write the report to a temp file first
(`ssh worker-server 'cat > /tmp/daily-audit-report.txt' <<'EOF' ... EOF`), then
`ssh worker-server '~/go/bin/goteleout "$(cat /tmp/daily-audit-report.txt)"; rm -f
/tmp/daily-audit-report.txt'` — piping a multi-line heredoc straight into a remote `"$(cat)"` inside
one SSH command silently returned exit 1 with no stderr on 2026-10-15; the file round-trip is the
version that's actually been proven to work, don't revert to the shorter form.

This proxy path occasionally stalls for minutes before succeeding (known, see the
`telegram-user-and-goteleout` memory) — confirmed again on 2026-10-16: two calls hung past a 90s
timeout back to back, the third went through in under a second. Wrap the call in `timeout 90`; if it
times out, kill the hung `goteleout` process **by its exact PID** (`pgrep -af goteleout`, never `pkill
-f` on a pattern — it also matches the wrapping `ssh`/`bash -c` command line) and retry once more
before treating it as a real delivery failure worth a Known-gaps note.

Send one Telegram message this way, in this fixed shape (same ✅/⚠️ emoji vocabulary homeobs already
uses elsewhere in this project) — add a `Dashboard:` line with today's screenshot link:

```
{{ ✅ | ⚠️ | 🔴 }} Daily audit — {{ YYYY-MM-DD }}

Reachability: {{ one line }}
Services: {{ ✅ all expected, running | one line per failed/new/removed service }}
worker-pc: {{ woke, pinged, shut down OK | already on, cycle skipped }}
Security baseline: {{ ✅ no changes | one line per flagged delta }}
Prometheus: {{ N/M targets up; firing alerts if any; trend flags if any }}
Tokens: {{ ✅ all live | one line per dead token }}
Dashboard: {{ 📸 <today's Yandex Disk link> | ⚠️ screenshot failed, see Known gaps }}
{{ one line per relevant entry from Known gaps, omit the section entirely if none apply today }}
```

Pick the top-line emoji from the worst thing found: 🔴 only if worker-server itself is unreachable
(Reachability failing makes every other section "unknown", say so instead of guessing), ⚠️ if there is
at least one other finding anywhere, ✅ only when every section reads all-clear — so the user can judge
severity from the first line alone. Keep each line short: this is a status scan, not a dump of raw
command output.

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

- No firewall visibility on the container or worker-pc with current permissions (worker-server's
  `ufw status` is checked — see Permissions / security baseline).
- No read access to the system journal (only this skill's own service-user-scoped journal entries are
  visible) — deeper incident diagnosis after a finding here still needs a human to grant
  `systemd-journal` group membership or do it themselves.
