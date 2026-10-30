#!/usr/bin/env bash
set -euo pipefail

fail() {
	echo "entrypoint: $1" >&2
	exit 1
}

warn() {
	echo "entrypoint: warning: $1" >&2
}

[ -d "$HOME/.codex" ] || fail "$HOME/.codex is missing, mount ./data/codex into it"
[ -w "$HOME/.codex" ] || fail "$HOME/.codex is not writable by uid $(id -u), fix ownership of ./data/codex"
[ -d /workspace ] || fail "/workspace is missing, mount ./workspace into it"
[ -w /workspace ] || fail "/workspace is not writable by uid $(id -u), fix ownership of ./workspace"

dotfiles="/home/claude/.claude-dotfiles"
if [ -d "$dotfiles" ]; then
	for link in "$HOME/.codex"/*; do
		[ -L "$link" ] || continue
		case "$(readlink "$link")" in
		"$dotfiles"/*)
			[ -e "$link" ] || rm "$link"
			;;
		esac
	done

	for src in "$dotfiles"/*.md; do
		[ -e "$src" ] || continue
		dst="$HOME/.codex/$(basename "$src")"
		if [ -d "$dst" ] && [ ! -L "$dst" ]; then
			rmdir "$dst" 2>/dev/null || {
				warn "$dst exists and is not empty, skipping link"
				continue
			}
		fi
		ln -sfn "$src" "$dst"
	done

	skills="$HOME/.codex/skills"
	if [ -L "$skills" ]; then
		[ "$(readlink "$skills")" = "$dotfiles/skills" ] || fail "unexpected skills symlink: $skills"
		unlink "$skills"
	fi
	mkdir -p "$skills"
	for link in "$skills"/*; do
		[ -L "$link" ] || continue
		case "$(readlink "$link")" in
		"$dotfiles/skills"/*)
			[ -e "$link" ] || unlink "$link"
			;;
		esac
	done
	for src in "$dotfiles"/skills/*; do
		[ -d "$src" ] || continue
		dst="$skills/$(basename "$src")"
		if [ -e "$dst" ] && [ ! -L "$dst" ]; then
			warn "$dst already exists, skipping link"
			continue
		fi
		ln -sfn "$src" "$dst"
	done
else
	fail "$dotfiles is not mounted, shared instructions and skills will be missing"
fi
[ -f "$dotfiles/AGENTS.md" ] || fail "shared AGENTS.md is missing"
ln -sfn /home/claude/.codex-config.toml "$HOME/.codex/config.toml"
mkdir -p "$HOME/.codex/rules"
ln -sfn /home/claude/.codex-ssh.rules "$HOME/.codex/rules/aibot-ssh.rules"

mkdir -p "$HOME/.ssh"
chmod 700 "$HOME/.ssh"

[ -f "$HOME/.ssh/config" ] || fail "$HOME/.ssh/config is missing, named ssh hosts will not resolve"
[ -s "$HOME/.ssh/known_hosts" ] || fail "$HOME/.ssh/known_hosts is missing or empty, populate it with ssh-keyscan"

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

stop_daemon() {
	trap - EXIT TERM INT
	if [ -n "${log_pid:-}" ]; then
		kill "$log_pid" 2>/dev/null || true
	fi
	if [ -n "${sleep_pid:-}" ]; then
		kill "$sleep_pid" 2>/dev/null || true
	fi
	timeout 15 codex app-server daemon stop || warn "daemon did not stop cleanly"
}

trap stop_daemon EXIT
trap 'exit 0' TERM INT

codex app-server daemon enable-remote-control
codex app-server daemon start

tail -n 100 -F "$HOME/.codex/app-server-daemon/daemon.stderr.log" &
log_pid=$!

while timeout 5 codex app-server daemon version | jq -e '.status == "running"' >/dev/null; do
	sleep 5 &
	sleep_pid=$!
	wait "$sleep_pid"
done

fail "Codex daemon is no longer running"
