# aibot-vps

```bash
cd dotfiles/aibot-vps

mkdir -p data/codex workspace secrets
touch ssh/known_hosts
ssh-keygen -t ed25519 -N "" -C claude-vps -f secrets/worker_server_ed25519
ssh-keygen -t ed25519 -N "" -C "claude-vps -> worker-pc" -f secrets/worker_pc_ed25519
ssh-copy-id -i secrets/worker_server_ed25519.pub -p <port> teadove@<home-endpoint>
 printf '%s' '<worker-server-sudo-password>' > secrets/worker_server_sudo_pass
 printf '%s' '<github-token>' > secrets/github_token
 printf '%s' '<openai-api-key>' > secrets/openai_api_key
 printf '%s' '<telegram-bot-token>' > secrets/telegram_bot_token
 printf '%s' '<yandex-disk-token>' > secrets/yandex_disk_token
 printf '%s' '<grafana-viewer-token>' > secrets/grafana_viewer_token
chmod 600 secrets/*

cp ssh/config.example ssh/config
ssh-keyscan -p <port> <home-endpoint> > ssh/known_hosts
ssh-keyscan -p <worker-port> <home-endpoint> >> ssh/known_hosts
chown -R 1000:1000 data workspace secrets ssh/known_hosts

docker compose run --rm agent codex login --device-auth # Login once
docker compose up -d
```

## Scheduled jobs

Supercronic runs alongside the agent as the same unprivileged user. No systemd or root access is needed.
The persistent crontab is `/workspace/cron/crontab`; changes are automatically reloaded.
Use standard five-field cron entries without a username. Times default to UTC.

For example, inside the container:

```bash
mkdir -p /workspace/cron
cp /workspace/cron/crontab /workspace/cron/crontab.next
printf '%s\n' '*/10 * * * * bash /workspace/availability-monitor/monitor.sh --once' \
    >> /workspace/cron/crontab.next
supercronic -test /workspace/cron/crontab.next && mv /workspace/cron/crontab.next /workspace/cron/crontab
```

Add each job only once. Scripts, configuration, and job state belong under `/workspace` to survive recreation.
Keep secrets in `/run/secrets`; do not put their values in scripts, cron entries, or logs.
Run each new job manually before scheduling it. Replace existing background loops when migrating to cron.
Job output and scheduler errors go to the container logs, available through `docker compose logs agent`.
Supercronic prevents overlapping executions of the same job by default.

If the scheduler exits or the daemon health check fails, the entrypoint stops services and exits.
The Compose restart policy restarts the container, including the scheduler. Docker must start on host boot.
Jobs do not run while the container is stopped, and missed runs are not replayed.
Allow sufficient container stop time for the longest job to finish gracefully.
Explicit command overrides, such as the login commands above, bypass the scheduler.

After changing the image sources, rebuild and recreate the service to activate the scheduler:

```bash
docker compose up -d --build
```
