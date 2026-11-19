#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
script_path="${1:-$script_dir/steamdeck-temp-ssh.sh}"
test_dir="$(mktemp -d /tmp/deck-ssh.test.XXXXXXXX)"
trap 'rm -rf -- "$test_dir"' EXIT
sshd_executable="$(command -v sshd || command -v /usr/sbin/sshd)"
command ssh-keygen -q -t ed25519 -N "" -f "$test_dir/host_key"

awk '
  /^if ! ssh-keygen -A; then$/ { copying = 1 }
  /^install -d -m 0700 / { exit }
  copying { print }
' "$script_path" > "$test_dir/preflight.sh"
test -s "$test_dir/preflight.sh"

ssh-keygen() {
  if [[ "${keygen_failure:-0}" == 1 ]]; then return 1; fi
  : > "$test_dir/keygen_called"
}

sshd() {
  test -f "$test_dir/keygen_called" || return 1
  if [[ "$1" == -T && "${canonical_output:-0}" == 1 ]]; then
    local settings
    settings="$("$sshd_executable" -f "$test_dir/sshd_config" "$@")" || return "$?"
    printf '%s\n' "${settings//passwordauthentication /PasswordAuthentication }"
  else
    "$sshd_executable" -f "$test_dir/sshd_config" "$@" || return "$?"
  fi
  if [[ "$1" == -T && "${extended_failure:-0}" == 1 ]]; then return 42; fi
}
export -f ssh-keygen sshd
export test_dir sshd_executable

run_case() {
  local name="$1" expected="$2" pattern="${3:-}" status=0
  rm -f "$test_dir/keygen_called"
  bash -Eeuo pipefail "$test_dir/preflight.sh" > "$test_dir/output" 2>&1 || status="$?"
  if [[ "$expected" == success && "$status" != 0 ]] || [[ "$expected" == failure && "$status" == 0 ]]; then
    printf "FAIL: %s (exit %s)\n" "$name" "$status" >&2
    cat "$test_dir/output" >&2
    exit 1
  fi
  if [[ -n "$pattern" ]]; then
    grep -Fq "$pattern" "$test_dir/output"
  fi
  printf "PASS: %s\n" "$name"
}

printf 'HostKey %s\nPasswordAuthentication yes\n' "$test_dir/host_key" > "$test_dir/sshd_config"
run_case "password authentication enabled" success

{
  printf "SetEnv"
  for ((index=0; index<2500; index++)); do printf " TEST%s=value" "$index"; done
  printf "\n"
} >> "$test_dir/sshd_config"
run_case "large effective configuration" success

export extended_failure=1
run_case "configuration output followed by failure" failure "Could not read the effective sshd configuration."
unset extended_failure

printf 'HostKey %s\nPasswordAuthentication no\n' "$test_dir/host_key" > "$test_dir/sshd_config"
run_case "password authentication disabled before setup" success

printf "InvalidSetting yes\n" >> "$test_dir/sshd_config"
run_case "invalid configuration" failure "sshd configuration is invalid"

export keygen_failure=1
run_case "host key generation failed" failure "Could not generate SSH host keys."

unset keygen_failure
awk '
  /^cat > .*setup-root.sh.*ROOT_SCRIPT/ { copying = 1; next }
  /^ROOT_SCRIPT$/ { exit }
  copying { print }
' "$script_path" | sed \
  -e "s#/var/lib/steamdeck-temp-ssh#$test_dir/state#g" \
  -e "s#/etc/ssh/sshd_config#$test_dir/sshd_config#g" \
  -e "s#/etc/systemd/system#$test_dir/units#g" > "$test_dir/setup.sh"
test -s "$test_dir/setup.sh"

getent() {
  printf 'deck:%s:0:0:0:0:0:0:\n' "$(cat "$test_dir/shadow")"
}

chpasswd() {
  local password_line
  IFS= read -r password_line
  [[ "$password_line" == deck:* ]] || return 1
  if [[ -f "$test_dir/fail_password" ]]; then return 1; fi
  printf '%s\n' 'temporary-hash' > "$test_dir/shadow"
}

usermod() {
  printf '%s\n' "$2" > "$test_dir/shadow"
}

systemctl() {
  local action="$1" target="${!#}"
  printf '%s\n' "$*" >> "$test_dir/systemctl.log"
  case "$action" in
    is-active) test -f "$test_dir/services/$target.active" ;;
    is-enabled) test -f "$test_dir/services/$target.enabled" ;;
    enable)
      if [[ "$target" == *.timer && -f "$test_dir/fail_timer" ]]; then return 1; fi
      touch "$test_dir/services/$target.active" "$test_dir/services/$target.enabled"
      ;;
    stop) rm -f "$test_dir/services/$target.active" ;;
    disable)
      rm -f "$test_dir/services/$target.enabled"
      if [[ "$*" == *--now* ]]; then rm -f "$test_dir/services/$target.active"; fi
      ;;
    reload)
      if [[ -f "$test_dir/fail_reload" ]]; then rm "$test_dir/fail_reload"; return 1; fi
      ;;
    daemon-reload) ;;
    *) printf 'Unexpected systemctl command: %s\n' "$*" >&2; return 1 ;;
  esac
}
export -f getent chpasswd usermod systemctl

prepare_setup() {
  rm -rf "$test_dir/state" "$test_dir/runtime" "$test_dir/services" "$test_dir/units"
  rm -f "$test_dir"/fail_*
  mkdir "$test_dir/runtime" "$test_dir/services" "$test_dir/units"
  printf 'initial-hash\n' > "$test_dir/shadow"
  printf 'HostKey %s\nPasswordAuthentication %s\n' "$test_dir/host_key" "$1" > "$test_dir/sshd_config"
  if [[ "$1" == no ]]; then
    printf 'Match User deck\n    PasswordAuthentication no\nMatch all\n' >> "$test_dir/sshd_config"
  fi
  cp "$test_dir/sshd_config" "$test_dir/original_config"
  if [[ "$2" == active ]]; then
    touch "$test_dir/services/sshd.service.active" "$test_dir/services/sshd.service.enabled"
  fi
  : > "$test_dir/systemctl.log"
}

run_setup() {
  local status=0
  printf 'test-password\n' | bash "$test_dir/setup.sh" "$test_dir/runtime" \
    "$(date -u -d '+4 hours' '+%Y-%m-%d %H:%M:%S UTC')" \
    "$(date -u -d '+4 hours' +%s)" > "$test_dir/output" 2>&1 || status="$?"
  if [[ "$status" != "$1" ]]; then
    cat "$test_dir/output" >&2
    printf 'Unexpected setup status: %s\n' "$status" >&2
    exit 1
  fi
}

assert_restored() {
  cmp "$test_dir/sshd_config" "$test_dir/original_config"
  test "$(cat "$test_dir/shadow")" = initial-hash
  test ! -e "$test_dir/state"
}

for original_service in active inactive; do
  prepare_setup no "$original_service"
  run_setup 0
  settings="$(sshd -T -C user=deck)"
  grep -qix 'passwordauthentication yes' <<< "$settings"
  settings="$(sshd -T -C user=unrelated)"
  grep -qix 'passwordauthentication no' <<< "$settings"
  grep -qx 'Persistent=true' "$test_dir/units/steamdeck-temp-ssh-cleanup.timer"
  bash "$test_dir/state/cleanup.sh"
  test -f "$test_dir/state/sshd_override.conf"
  bash "$test_dir/state/cleanup.sh" --force
  assert_restored
  if [[ "$original_service" == active ]]; then
    test -f "$test_dir/services/sshd.service.active"
    grep -q '^reload sshd.service$' "$test_dir/systemctl.log"
  else
    test ! -f "$test_dir/services/sshd.service.active"
    test ! -f "$test_dir/services/sshd.service.enabled"
  fi
  printf 'PASS: per-user access and revocation with originally %s SSH\n' "$original_service"
done

prepare_setup no inactive
run_setup 0
printf '0\n' > "$test_dir/state/expire_epoch"
bash "$test_dir/state/cleanup.sh"
assert_restored
printf 'PASS: expired access cleanup\n'

prepare_setup yes active
run_setup 0
test ! -f "$test_dir/state/sshd_override.conf"
cmp "$test_dir/sshd_config" "$test_dir/original_config"
bash "$test_dir/state/cleanup.sh" --force
assert_restored
printf 'PASS: existing password authentication preserved\n'

for failure in timer password reload; do
  prepare_setup no active
  touch "$test_dir/fail_$failure"
  run_setup 1
  assert_restored
  printf 'PASS: rollback after %s failure\n' "$failure"
done

prepare_setup no active
run_setup 0
printf 'ClientAliveInterval 300\n' >> "$test_dir/sshd_config"
printf 'ClientAliveInterval 300\n' >> "$test_dir/original_config"
printf 'manually-changed-hash\n' > "$test_dir/shadow"
bash "$test_dir/state/cleanup.sh" --force
cmp "$test_dir/sshd_config" "$test_dir/original_config"
test "$(cat "$test_dir/shadow")" = manually-changed-hash
test ! -e "$test_dir/state"
printf 'PASS: independent configuration and password changes preserved\n'

prepare_setup no active
run_setup 0
touch "$test_dir/fail_reload"
if bash "$test_dir/state/cleanup.sh" --force; then
  printf 'Cleanup unexpectedly ignored a reload failure\n' >&2
  exit 1
fi
test -f "$test_dir/state/cleanup.sh"
bash "$test_dir/state/cleanup.sh" --force
assert_restored
printf 'PASS: failed cleanup retains state for retry\n'

export canonical_output=1
for original_authentication in yes no; do
  prepare_setup "$original_authentication" active
  run_setup 0
  if [[ "$original_authentication" == yes ]]; then
    test ! -f "$test_dir/state/sshd_override.conf"
    cmp "$test_dir/sshd_config" "$test_dir/original_config"
  else
    test -f "$test_dir/state/sshd_override.conf"
  fi
  bash "$test_dir/state/cleanup.sh" --force
  assert_restored
  printf 'PASS: canonical option casing with initial password authentication %s\n' "$original_authentication"
done
