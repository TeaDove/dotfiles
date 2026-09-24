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

The worker's MAC address is deliberately not written here — it is stored in your memory,
recall it from there. Exact hardware specs of the Raspberry and `worker-pc` are in memory
as well.

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

## Yandex Disk

The OAuth application `AI-Junk` has the single scope `cloud_api:disk.app_folder`, so the token
only sees the application folder (`Приложения/AI-Junk` in the web UI). Address it with the
`app:/` path prefix; `disk:/` paths are forbidden for this token.

The token is stored in this container at `/run/secrets/yandex_disk_token`. It is an
`access_token` obtained with `response_type=token`; there is no refresh token, and the client
secret is not needed. When the token expires, the user issues a new one at
`https://oauth.yandex.ru/authorize?response_type=token&client_id=<CLIENT_ID>`.

Use the Python library [yadisk](https://github.com/ivknv/yadisk)
([documentation](https://yadisk.readthedocs.io)) on the Raspberry through `uv`, no installation
step is required:

```bash
ssh raspberry 'uv run --quiet --with "yadisk[sync-defaults]" python /path/to/script.py' \
  < /run/secrets/yandex_disk_token
```

```python
import sys

import yadisk

with yadisk.Client(token=sys.stdin.read().strip()) as client:
    client.upload("/local/report.pdf", "app:/report.pdf", overwrite=True)
    client.publish("app:/report.pdf")
    print(client.get_meta("app:/report.pdf").public_url)
```

The script must live in a file because stdin is taken by the token. `get_disk_info()` returns
403 because the `cloud_api:disk.info` scope is absent.

Pass the token over stdin or a protected file, never as a command argument. The REST API itself
is described at https://yandex.ru/dev/disk-api/doc/ru/; upload and download go through one-time
links returned by the API, which `yadisk` handles on its own.

Other clients: `rclone` supports the application folder only starting with the release after
v1.75 (`app_folder` backend option), the `ydcmd` CLI is archived, and there is no mature Go SDK,
so Go services should call the REST API directly.
