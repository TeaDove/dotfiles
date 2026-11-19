#!/usr/bin/env bash
# Steam Deck temporary password SSH with offline QR. Requires SteamOS Desktop Mode.
# A copy of python-qrcode 8.2 (BSD-3-Clause) is bundled at the end for fully offline
# QR generation. Source: https://github.com/lincolnloop/python-qrcode
# Its LICENSE.qrcode is included in the embedded archive.
set -Eeuo pipefail
umask 077

STATE=/var/lib/steamdeck-temp-ssh
UNIT=steamdeck-temp-ssh-cleanup
DURATION_HOURS=4

notify_error() {
  printf '\n[ERROR] %s\n' "$*" >&2
  if command -v kdialog >/dev/null 2>&1; then
    kdialog --error "$*" --title 'Steam Deck: temporary SSH' >/dev/null 2>&1 || true
  fi
}
trap 'notify_error "Ошибка на строке $LINENO. Подробности смотри в открывшейся Konsole."' ERR

if [[ "$(id -un)" != 'deck' ]]; then
  notify_error 'Запускай файл из пользователя deck, не от root.'
  exit 1
fi

if [[ "${1:-}" == '--revoke' ]]; then
  sudo -v
  sudo /usr/bin/bash "$STATE/cleanup.sh" --force
  echo 'Временный доступ отозван. Нажми Enter.'
  read -r _ || true
  exit 0
fi

if ! command -v python3 >/dev/null 2>&1; then
  notify_error 'В системе не найден python3.'
  exit 1
fi
if ! command -v konsole >/dev/null 2>&1 && [[ ! -t 0 ]]; then
  notify_error 'Для запуска нужен Desktop Mode (KDE Plasma).'
  exit 1
fi

# Launch via Dolphin (executable .sh): automatically open interactive Konsole.
if [[ ! -t 0 ]]; then
  exec konsole --hold -e /usr/bin/bash "$0" "$@"
fi

if [[ -d "$STATE" ]]; then
  notify_error 'Предыдущая SSH-сессия ещё зарегистрирована. Дождись окончания либо запусти этот скрипт с --revoke.'
  exit 1
fi

# On stock SteamOS user deck has no password, so sudo cannot authenticate yet.
passwd_status="$(LC_ALL=C passwd -S deck 2>/dev/null | awk '{print $2}')"
if [[ "$passwd_status" == 'NP' || "$passwd_status" == 'L' ]]; then
  echo 'Первый запуск: на SteamOS ещё не настроен пароль deck.'
  echo 'Сейчас потребуется один раз задать локальный пароль для sudo.'
  echo 'Введи новый пароль два раза (символы при вводе не показываются).'
  passwd
fi

echo 'Подтверди текущий пароль deck для sudo, если появится запрос.'
sudo -v

# Steam Deck local IP is the source address of its default route.
LAN_IP="$(ip -4 route get 1.1.1.1 2>/dev/null | sed -n 's/.* src \([0-9.]\+\).*/\1/p' | head -1)"
if [[ -z "$LAN_IP" ]]; then
  LAN_IP="$(hostname -I 2>/dev/null | tr ' ' '\n' | grep -E '^[0-9]+(\.[0-9]+){3}$' | head -1 || true)"
fi
if [[ -z "$LAN_IP" || "$LAN_IP" == '127.'* ]]; then
  notify_error 'Не удалось определить IP Steam Deck в локальной сети. Подключись к Wi-Fi.'
  exit 1
fi

# Password contains only hexadecimal characters (160 bits of randomness).
TEMP_PASSWORD="$(od -An -N20 -tx1 /dev/urandom | tr -d ' \n')"
END_UTC="$(date -u -d '+4 hours' '+%Y-%m-%d %H:%M:%S UTC')"
END_EPOCH="$(date -u -d '+4 hours' +%s)"
RUNTIME_DIR="$(mktemp -d "${XDG_RUNTIME_DIR:-/tmp}/deck-ssh.XXXXXXXX")"
chmod 700 "$RUNTIME_DIR"

# Root helper is a separate (non-secret) file. Password goes to its stdin,
# not argv, process environment, the log, or /tmp files.
cat > "$RUNTIME_DIR/setup-root.sh" <<'ROOT_SCRIPT'
#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
state=/var/lib/steamdeck-temp-ssh
unit=steamdeck-temp-ssh-cleanup
rundir="$1"
end_cal="$2"
end_epoch="$3"
IFS= read -r temporary_password

if [[ -e "$state" ]]; then
  echo 'A previous temporary SSH session already exists.' >&2
  exit 1
fi
if [[ ! "$end_epoch" =~ ^[0-9]+$ ]]; then
  echo 'Invalid expiration' >&2
  exit 1
fi
if [[ ! "$rundir" == /run/user/*/deck-ssh.* && ! "$rundir" == /tmp/deck-ssh.* ]]; then
  echo 'Invalid runtime directory' >&2
  exit 1
fi

# Use the existing SteamOS SSH settings. Fail safely if password login is disabled.
if ! sshd -t; then
  echo 'sshd configuration is invalid.' >&2
  exit 1
fi
if ! sshd -T -C user=deck,host=steamdeck,addr=127.0.0.1 | grep -qx 'passwordauthentication yes'; then
  echo 'PasswordAuthentication is disabled in sshd_config. No changes were made.' >&2
  exit 1
fi

install -d -m 0700 "$state"
getent shadow deck | cut -d: -f2 > "$state/old_hash"
[[ -s "$state/old_hash" ]] || { echo 'Cannot read original deck password hash' >&2; rm -rf "$state"; exit 1; }
if systemctl is-active --quiet sshd.service; then echo yes > "$state/was_active"; else echo no > "$state/was_active"; fi
if systemctl is-enabled --quiet sshd.service; then echo yes > "$state/was_enabled"; else echo no > "$state/was_enabled"; fi
printf '%s\n' "$end_epoch" > "$state/expire_epoch"
printf '%s\n' "$rundir" > "$state/runtime_dir"

cat > "$state/cleanup.sh" <<'CLEANUP'
#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
state=/var/lib/steamdeck-temp-ssh
unit=steamdeck-temp-ssh-cleanup
[[ -f "$state/old_hash" ]] || exit 0
if [[ "${1:-}" != '--force' ]]; then
  end_epoch="$(cat "$state/expire_epoch")"
  if (( $(date -u +%s) < end_epoch )); then exit 0; fi
fi
was_active="$(cat "$state/was_active")"
was_enabled="$(cat "$state/was_enabled")"
old_hash="$(cat "$state/old_hash")"
current_hash="$(getent shadow deck | cut -d: -f2)"
# Never clobber a password manually changed during the 4h window.
if [[ ! -f "$state/temp_hash" || "$current_hash" == "$(cat "$state/temp_hash")" ]]; then
  usermod -p "$old_hash" deck
else
  echo 'deck password has changed since setup; preserving the newer password.' >&2
fi
if [[ "$was_active" == no ]]; then
  systemctl stop sshd.service || true
fi
if [[ "$was_enabled" == no ]]; then
  systemctl disable sshd.service || true
fi
systemctl disable --now "$unit.timer" || true
rm -f "/etc/systemd/system/$unit.timer" "/etc/systemd/system/$unit.service"
systemctl daemon-reload || true
rundir="$(cat "$state/runtime_dir")"
if [[ "$rundir" == /run/user/*/deck-ssh.* || "$rundir" == /tmp/deck-ssh.* ]]; then
  rm -rf -- "$rundir"
fi
rm -rf -- "$state"
CLEANUP
chmod 700 "$state/cleanup.sh"

cat > "/etc/systemd/system/$unit.service" <<SERVICE
[Unit]
Description=Restore deck password and remove temporary Steam Deck SSH access
[Service]
Type=oneshot
ExecStart=/usr/bin/bash $state/cleanup.sh
SERVICE
cat > "/etc/systemd/system/$unit.timer" <<TIMER
[Unit]
Description=Revoke temporary Steam Deck SSH at expiration (including after reboot)
[Timer]
OnCalendar=$end_cal
Persistent=true
AccuracySec=1s
Unit=$unit.service
[Install]
WantedBy=timers.target
TIMER

systemctl daemon-reload
# Arm the durable revocation *before* changing the account password.
systemctl enable --now "$unit.timer"
if ! printf 'deck:%s\n' "$temporary_password" | chpasswd; then
  bash "$state/cleanup.sh" --force || true
  echo 'Unable to set temporary password' >&2
  exit 1
fi
getent shadow deck | cut -d: -f2 > "$state/temp_hash"
if ! systemctl enable --now sshd.service; then
  bash "$state/cleanup.sh" --force || true
  echo 'Unable to start sshd' >&2
  exit 1
fi
if ! systemctl is-active --quiet "$unit.timer"; then
  bash "$state/cleanup.sh" --force || true
  echo 'Expiry timer is not active, temporary access was rolled back' >&2
  exit 1
fi
printf 'Temporary SSH is active. Automatic cleanup: %s\n' "$end_cal"
ROOT_SCRIPT
chmod 700 "$RUNTIME_DIR/setup-root.sh"
printf '%s\n' "$TEMP_PASSWORD" | sudo /usr/bin/bash "$RUNTIME_DIR/setup-root.sh" "$RUNTIME_DIR" "$END_UTC" "$END_EPOCH"
rm -f "$RUNTIME_DIR/setup-root.sh"

# Include everything Codex needs in ONE QR. Password never leaves this device
# until the user explicitly forwards the QR contents to their Telegram bot.
export LAN_IP TEMP_PASSWORD END_UTC
if ! python3 - "$0" "$RUNTIME_DIR" <<'PYTHON'
import base64, html, os, pathlib, sys, tarfile, io
script = pathlib.Path(sys.argv[1])
rundir = pathlib.Path(sys.argv[2])
marker = b'__QRPY_PAYLOAD_BELOW__\n'
blob = script.read_bytes().split(marker, 1)[1]
libdir = rundir / 'qrlib'
libdir.mkdir(mode=0o700)
with tarfile.open(fileobj=io.BytesIO(base64.b64decode(blob)), mode='r:gz') as archive:
    for member in archive.getmembers():
        if not member.isfile():
            continue
        target = (libdir / member.name).resolve()
        if not target.is_relative_to(libdir.resolve()):
            raise ValueError('Unsafe bundled QR library path')
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(archive.extractfile(member).read())
sys.path.insert(0, str(libdir))
from qrcode.main import QRCode
from qrcode.constants import ERROR_CORRECT_M

ip = os.environ['LAN_IP']
password = os.environ['TEMP_PASSWORD']
end = os.environ['END_UTC']
message = (f'Steam Deck SSH until {end}\n'
           f'IP: {ip}\n'
           'User: deck\n'
           f'Password: {password}\n'
           'Port: 22\n'
           'Jump: worker-server\n'
           f'Connect: ssh -J worker-server deck@{ip}\n'
           'Task: Fix Battle.net/Proton; do not delete games.\n')
qr = QRCode(error_correction=ERROR_CORRECT_M, border=4, box_size=9)
qr.add_data(message)
qr.make(fit=True)
mat = qr.get_matrix()
size = len(mat)
path = ' '.join(f'M{x},{y}h1v1h-1z' for y,row in enumerate(mat) for x,on in enumerate(row) if on)
svg = (f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {size} {size}" '
       'shape-rendering="crispEdges" width="440" height="440">'
       f'<rect width="{size}" height="{size}" fill="white"/>'
       f'<path d="{path}" fill="black"/></svg>')
page = f'''<!doctype html><html lang="en"><meta charset="utf-8">
<title>Temporary SSH for Steam Deck</title>
<style>body{{background:#101821;color:#e9f1f5;font-family:system-ui;margin:20px auto;max-width:650px;text-align:center}}
.qr{{background:white;padding:16px;display:inline-block;border-radius:14px}}
pre{{white-space:pre-wrap;text-align:left;overflow-wrap:anywhere;background:#182b35;border-radius:10px;padding:16px}}
small{{color:#a5bcc9}} button{{font-size:17px;padding:10px 20px;border-radius:8px;cursor:pointer}}</style>
<h2>Steam Deck: temporary SSH (4 hours)</h2>
<div class="qr">{svg}</div>
<p>Scan this QR with your phone, copy the text and send it to your private Codex Telegram bot.</p>
<p><strong>Expires: {html.escape(end)}</strong></p>
<pre id="credentials">{html.escape(message)}</pre>
<button onclick="navigator.clipboard.writeText(document.getElementById('credentials').innerText)">Copy details</button>
<p><small>The QR is generated locally. No external QR service is contacted.</small></p>
<p><small>Important: it changes the deck account password for 4 hours, including sudo.</small></p></html>'''
(rundir/'qr.html').write_text(page, encoding='utf-8')
(rundir/'qr.html').chmod(0o600)
PYTHON
then
  echo 'QR generation failed; revoking the temporary SSH access now.' >&2
  sudo /usr/bin/bash "$STATE/cleanup.sh" --force || true
  exit 1
fi

xdg-open "file://$RUNTIME_DIR/qr.html" >/dev/null 2>&1 || {
  echo "QR page: $RUNTIME_DIR/qr.html"
  echo 'Could not open the browser automatically. Open the page manually in Dolphin.'
}

echo
echo 'READY: temporary SSH access via LAN'
echo "IP: $LAN_IP | User: deck | Port: 22"
echo "Expires: $END_UTC"
echo 'Scan QR code in the browser, forward the message to the Codex bot.'
echo 'To revoke early: bash steamdeck-temp-ssh.sh --revoke'
echo 'You may close this terminal; systemd keeps the 4-hour timer running.'
exit 0
__QRPY_PAYLOAD_BELOW__
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
