#!/usr/bin/env bash
set -euo pipefail
umask 077

state_dir=${CODEX_SCHEDULE_DIR:-/workspace/codex-schedules}
crontab=${CODEX_CRONTAB:-/workspace/cron/crontab}
script=$(realpath "${BASH_SOURCE[0]}")

help() {
    cat <<'HELP'
Usage:
  codex-schedule.sh once NAME 'YYYY-MM-DD HH:MM' THREAD_ID PROMPT_FILE
  codex-schedule.sh cron NAME 'MIN HOUR DOM MONTH DOW' THREAD_ID PROMPT_FILE
  codex-schedule.sh send THREAD_ID PROMPT_FILE
  codex-schedule.sh list
  codex-schedule.sh remove NAME
  codex-schedule.sh --help

Send prompts to existing Codex chats through the running local app-server.
The chat keeps its model and permission settings. No new chat is created.

once: send once at or after the chosen time, checked every minute. Times use
Europe/Moscow unless CODEX_SCHEDULE_TZ is set. Missed one-time jobs run when the
container returns. Queue acceptance marks the job sent, preventing resubmission.
cron: repeat using a five-field cron expression in the scheduler's timezone
(UTC in this container). Missed recurring runs are not replayed.
send: queue immediately and request that the local daemon resume the chat.
list: show job names, type, schedule, and whether a one-time job was sent.
remove: unschedule the job; preserve its prompt and logs. Does not cancel an
already queued message or a running turn. Names cannot be reused.

NAME must contain only letters, digits, underscores and hyphens.
PROMPT_FILE must contain non-empty UTF-8 text and is copied at registration.
Do not put secrets in prompts. Job state and logs are private files under
/workspace/codex-schedules. Cron changes are validated before atomic replacement.
Codex must already be authenticated and its local app-server must be running.
A successful enqueue does not guarantee execution: codex queue --resume warns
but still exits successfully if resuming fails. Inspect run.log and the chat;
do not resend blindly, because the message may already be queued.
Failed queue commands retry next minute for once, or on the next cron firing.

Examples:
  codex-schedule.sh once audit-reminder '2026-11-03 16:00' "$THREAD_ID" /workspace/prompt.txt
  codex-schedule.sh cron daily-audit '0 13 * * *' "$THREAD_ID" /workspace/prompt.txt
  codex-schedule.sh remove daily-audit

Testing overrides: CODEX_SCHEDULE_DIR and CODEX_CRONTAB select alternate state
and crontab paths. Scheduled commands use the defaults unless their environment
also provides these overrides. The internal run command is used by cron only.
HELP
}

fail() {
    printf 'codex-schedule: %s\n' "$*" >&2
    exit 1
}

valid_name() {
    [[ $1 =~ ^[a-zA-Z0-9_-]+$ ]] || fail 'Invalid job name'
}

check_prompt() {
    [[ $1 =~ ^[a-zA-Z0-9_-]+$ ]] || fail 'Invalid thread ID'
    [ -s "$2" ] || fail 'Prompt file is missing or empty'
    [ -n "$(cat "$2")" ] || fail 'Prompt text is empty'
}

update_cron() {
    local action=$1 entry=$2 candidate
    mkdir -p "$(dirname "$crontab")" || return 1
    exec 9>"$crontab.lock"
    flock 9 || return 1
    touch "$crontab" || return 1
    candidate=$(mktemp "${crontab}.XXXXXX") || return 1
    if ! awk -v target="$entry" '$0 != target' "$crontab" > "$candidate"; then
        rm -f "$candidate"
        printf "%s\n" "Cannot read crontab" >&2
        return 1
    fi
    if [ "$action" = add ]; then
        printf '%s\n' "$entry" >> "$candidate"
    fi
    if ! supercronic -test "$candidate"; then
        rm -f "$candidate"
        printf "%s\n" "Crontab validation failed; schedule unchanged" >&2
        return 1
    fi
    mv "$candidate" "$crontab" || return 1
    flock -u 9
    exec 9>&-
}

send_prompt() {
    check_prompt "$1" "$2"
    timeout 90 codex queue -C /workspace --thread "$1" --message "$(cat "$2")" --resume
}

register() {
    local mode=$1 name=$2 schedule=$3 thread=$4 prompt=$5 job due=0 entry
    local -a fields
    valid_name "$name"
    check_prompt "$thread" "$prompt"
    [[ $state_dir =~ ^/[a-zA-Z0-9_./-]+$ && $script =~ ^/[a-zA-Z0-9_./-]+$ ]] ||
        fail 'Script and state paths must be absolute and contain no shell metacharacters'
    if [ "$mode" = once ]; then
        due=$(TZ="${CODEX_SCHEDULE_TZ:-Europe/Moscow}" date -d "$schedule" +%s)
        [ "$due" -gt "$(date +%s)" ] || fail 'One-time date must be in the future'
    else
        [[ $schedule != *$'\n'* && $schedule != *$'\r'* ]] || fail 'Invalid cron expression'
        read -r -a fields <<< "$schedule"
        [ "${#fields[@]}" -eq 5 ] || fail 'Cron expression must have five fields'
        [[ $schedule =~ ^[a-zA-Z0-9*/?,[:space:]-]+$ ]] || fail 'Invalid cron expression'
    fi
    mkdir -p "$state_dir"
    job="$state_dir/$name"
    mkdir "$job" || fail 'Job already exists; choose a new name'
    printf '%s\n' "$mode" > "$job/mode"
    printf '%s\n' "$schedule" > "$job/schedule"
    printf '%s\n' "$thread" > "$job/thread"
    printf '%s\n' "$due" > "$job/due"
    cp "$prompt" "$job/prompt"
    entry="${schedule} /bin/bash $script run $name >> $job/run.log 2>&1"
    if [ "$mode" = once ]; then
        entry="* * * * * /bin/bash $script run $name >> $job/run.log 2>&1"
    fi
    printf '%s\n' "$entry" > "$job/entry"
    if ! update_cron add "$entry"; then
        touch "$job/disabled"
        fail "Registration failed; saved job is disabled: $name"
    fi
    printf 'Scheduled %s (%s): %s\n' "$name" "$mode" "$schedule"
}

run_job() {
    local job="$state_dir/$1"
    [ -d "$job" ] || fail 'Unknown job'
    exec 8>"$job/lock"
    flock -n 8 || return 0
    [ ! -e "$job/disabled" ] || return 0
    if [ -e "$job/sent" ]; then
        update_cron remove "$(cat "$job/entry")"
        return 0
    fi
    if [ "$(cat "$job/mode")" = once ]; then
        [ "$(date +%s)" -ge "$(cat "$job/due")" ] || return 0
    fi
    printf '%s Sending queued prompt\n' "$(date -u +%FT%TZ)"
    send_prompt "$(cat "$job/thread")" "$job/prompt"
    if [ "$(cat "$job/mode")" = once ]; then
        touch "$job/sent"
        update_cron remove "$(cat "$job/entry")"
    fi
}

case "${1:---help}" in
    --help|-h|help) help ;;
    send)
        [ "$#" -eq 3 ] || fail 'send requires THREAD_ID PROMPT_FILE; use --help'
        send_prompt "$2" "$3"
        ;;
    once|cron)
        [ "$#" -eq 5 ] || fail 'Scheduling requires NAME SCHEDULE THREAD_ID PROMPT_FILE; use --help'
        register "$1" "$2" "$3" "$4" "$5"
        ;;
    run)
        [ "$#" -eq 2 ] || fail 'run requires NAME'
        valid_name "$2"
        run_job "$2"
        ;;
    remove)
        [ "$#" -eq 2 ] || fail 'remove requires NAME'
        valid_name "$2"
        [ -d "$state_dir/$2" ] || fail 'Unknown job'
        exec 8>"$state_dir/$2/lock"
        flock 8
        update_cron remove "$(cat "$state_dir/$2/entry")"
        touch "$state_dir/$2/disabled"
        ;;
    list)
        for job in "$state_dir"/*; do
            [ -d "$job" ] || continue
            status=scheduled
            [ ! -e "$job/sent" ] || status=sent
            [ ! -e "$job/disabled" ] || status=disabled
            printf '%s\t%s\t%s\t%s\n' "${job##*/}" "$(cat "$job/mode")" "$(cat "$job/schedule")" "$status"
        done
        ;;
    *) fail 'Unknown command; use --help' ;;
esac
