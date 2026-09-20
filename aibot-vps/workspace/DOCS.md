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
