#!/usr/bin/env bash
set -euo pipefail

fail() {
	echo "entrypoint: $1" >&2
	exit 1
}

warn() {
	echo "entrypoint: warning: $1" >&2
}

[ -d "$HOME/.claude" ] || fail "$HOME/.claude is missing, mount ./data/claude into it"
[ -w "$HOME/.claude" ] || fail "$HOME/.claude is not writable by uid $(id -u), fix ownership of ./data/claude"
[ -d /workspace ] || fail "/workspace is missing, mount ./workspace into it"
[ -w /workspace ] || fail "/workspace is not writable by uid $(id -u), fix ownership of ./workspace"

mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"

[ -f "$HOME/.ssh/config" ] || warn "$HOME/.ssh/config is missing, named ssh hosts will not resolve"
[ -s "$HOME/.ssh/known_hosts" ] || warn "$HOME/.ssh/known_hosts is missing or empty, populate it with ssh-keyscan"

if [ -d /run/secrets ]; then
	mkdir -p "$HOME/.ssh/keys"
	chmod 700 "$HOME/.ssh/keys"

	for secret in /run/secrets/*; do
		[ -e "$secret" ] || continue
		[ -r "$secret" ] || fail "$secret is not readable by uid $(id -u), fix ownership of ./secrets"
		install -m 600 "$secret" "$HOME/.ssh/keys/$(basename "$secret")"
	done
fi

if [ "$#" -gt 0 ]; then
	exec "$@"
fi

exec claude rc
