#!/usr/bin/env bash
set -euo pipefail

proxy_target="teadove.space:7999"
goteleout_bin="/home/teadove-aibot/go/bin/goteleout"

if [ "$#" -eq 0 ]; then
	echo "usage: $(basename "$0") <file on worker-server>..." >&2
	exit 2
fi

for path in "$@"; do
	case "$path" in
	*[[:space:]]*)
		echo "$(basename "$0"): path with whitespace is not supported: $path" >&2
		exit 2
		;;
	esac
done

port=$((20000 + RANDOM % 10000))

timeout 600 ssh \
	-o ExitOnForwardFailure=yes \
	-o ServerAliveInterval=0 \
	-o PermitRemoteOpen="$proxy_target" \
	-R "127.0.0.1:$port" \
	worker-server \
	"HTTPS_PROXY=socks5://127.0.0.1:$port $goteleout_bin --file $(printf '%q ' "$@") < /dev/null"
