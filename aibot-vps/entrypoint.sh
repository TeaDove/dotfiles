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

dotfiles="/home/claude/.claude-dotfiles"
if [ -d "$dotfiles" ]; then
	for link in "$HOME/.claude"/*; do
		[ -L "$link" ] || continue
		case "$(readlink "$link")" in
		"$dotfiles"/*)
			[ -e "$link" ] || rm "$link"
			;;
		esac
	done

	for src in "$dotfiles"/*.md "$dotfiles"/skills "$dotfiles"/agents; do
		[ -e "$src" ] || continue
		dst="$HOME/.claude/$(basename "$src")"
		if [ -d "$dst" ] && [ ! -L "$dst" ]; then
			rmdir "$dst" 2>/dev/null || {
				warn "$dst exists and is not empty, skipping link"
				continue
			}
		fi
		ln -sfn "$src" "$dst"
	done
else
	warn "$dotfiles is not mounted, claude configs will be missing"
fi

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
