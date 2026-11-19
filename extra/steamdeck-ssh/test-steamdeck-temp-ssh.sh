#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
script_path="${1:-$script_dir/steamdeck-temp-ssh.sh}"
test_dir="$(mktemp -d)"
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
  "$sshd_executable" -f "$test_dir/sshd_config" "$@" || return "$?"
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
run_case "password authentication disabled" failure "PasswordAuthentication is disabled"

printf "InvalidSetting yes\n" >> "$test_dir/sshd_config"
run_case "invalid configuration" failure "sshd configuration is invalid"

export keygen_failure=1
run_case "host key generation failed" failure "Could not generate SSH host keys."
