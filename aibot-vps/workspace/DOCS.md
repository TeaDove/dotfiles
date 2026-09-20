# Infrastructure documentation

## Proxies

Direct OpenAI and Telegram API access does not work from the remote SSH hosts. You must
use these proxies.

### OpenAI

Proxy root:

```text
https://teadove.space:7999/proxy/openai/
```

SDK/API base URL:

```text
https://teadove.space:7999/proxy/openai/v1
```

Example:

```bash
curl https://teadove.space:7999/proxy/openai/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

For SDKs that support a configurable base URL, use:

```text
OPENAI_BASE_URL=https://teadove.space:7999/proxy/openai/v1
```

### Telegram Bot API

Proxy root:

```text
https://teadove.space:7999/proxy/telegram-bot/
```

Example:

```bash
curl "https://teadove.space:7999/proxy/telegram-bot/$TELEGRAM_BOT_TOKEN/sendMessage"
```

For Telegram libraries that support a configurable API base URL, point them at the proxy
instead of `api.telegram.org`.

## Worker PC lifecycle

Wake `worker-pc` through the Raspberry (`wakeonlan` is expected there; install it via apt
if missing):

```bash
ssh raspberry 'wakeonlan <WORKER_MAC>'
```

Then poll until SSH is up (boot takes a minute or two):

```bash
until ssh -o ConnectTimeout=5 worker-pc true; do sleep 10; done
```

The `teadove-aibot` account on `worker-pc` has no general root access; its sudo is limited
to exactly two commands:

```bash
ssh worker-pc 'sudo systemctl poweroff'
ssh worker-pc 'sudo systemctl reboot'
```

Use `poweroff` after your work only if you powered the machine on yourself (see the rules
in CLAUDE.md).
