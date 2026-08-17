#!/usr/bin/env bash
# Claude Code desktop-notification hook (macOS).
#
# Wired from ~/.claude/settings.json for two events:
#   Stop         -> notify.sh stop          (Claude finished the turn)
#   Notification -> notify.sh notification  (Claude waits for permission/input)
#
# Reads the hook payload (JSON) on stdin. The (possibly long) auto-generated
# chat title goes into the notification BODY — it fits there but not in the
# short title bar; the title bar carries a fixed headline instead. The repo
# folder is the subtitle, so several agents over the same repo stay distinct.

event="${1:-}"
input="$(cat)"

json() { printf '%s' "$input" | jq -r "$1" 2>/dev/null || true; }

cwd="$(json '.cwd // empty')"
[ -z "$cwd" ] && cwd="$PWD"
dir="$(basename "$cwd")"

# The auto-generated chat title is stored in the transcript as a line
# {"type":"ai-title","aiTitle":"..."}. Take the most recent one.
title=""
transcript="$(json '.transcript_path // empty')"
if [ -n "$transcript" ] && [ -f "$transcript" ]; then
  title="$(grep '"aiTitle"' "$transcript" 2>/dev/null | tail -1 | jq -r '.aiTitle // empty' 2>/dev/null || true)"
fi
[ -z "$title" ] && title="$dir"

case "$event" in
  stop)         headline="Task done";          sound="Hero" ;;
  notification) headline="Need you attention";  sound="Hero" ;;
  *)            headline="Claude Code";         sound="Hero" ;;
esac

# AppleScript strings are double-quoted; strip characters that would break out.
sanitize() { printf '%s' "$1" | tr -d '"\\'; }
body="$(sanitize "$title")"
title_bar="$(sanitize "Claude · $headline")"
dir="$(sanitize "$dir")"

osascript -e "display notification \"$body\" with title \"$title_bar\" subtitle \"$dir\" sound name \"$sound\"" 2>/dev/null || true
goteleout "$title_bar in ($dir) about \"$body\""
