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

	for src in "$dotfiles"/*.md "$dotfiles"/skills; do
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
else
	fail "$dotfiles is not mounted, shared instructions and skills will be missing"
fi
[ -f "$dotfiles/CLAUDE.md" ] || fail "shared CLAUDE.md is missing"
ln -sfn "$dotfiles/CLAUDE.md" "$HOME/.codex/AGENTS.md"
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

exec codex remote-control --no-daemon
