# Temporary SSH for Codex (Steam Deck / SteamOS)

Copy these files to `extra/steamdeck-ssh/` on the `master` branch of
`https://github.com/teadove/dotfiles`:

- `steamdeck-temp-ssh.sh`
- `start-steamdeck-ssh.desktop`

Then download **only** `start-steamdeck-ssh.desktop` on the Steam Deck.
In Dolphin, open it and accept "Trust and Launch" if prompted. On some
KDE versions you must right-click -> Properties -> Permissions ->
"Is executable" before opening. The launcher automatically opens
Konsole and fetches the `.sh` from your own repository's `master` branch.

First run requires setting/entering the local `deck` sudo password.
After setup, the script opens a LOCAL HTML page with an OFFLINE QR code
containing the generated temporary password and the Deck LAN IP. Scan
and forward its text only to your private Codex Telegram bot.

Access lasts 4 hours. A persistent systemd calendar timer restores the
previous deck password hash (unless the account password was independently
changed) and returns sshd to its original service state. On the original
SteamOS setup, the service was stopped, and it will stop again. Password
login is temporarily enabled only for deck if needed; cleanup removes that override.

To revoke early while access is active:

    bash ~/Downloads/steamdeck-temp-ssh.sh --revoke

The remote Codex agent needs a working `worker-server` SSH jump host
with TCP forwarding allowed, a route from that home server to the
Steam Deck LAN IP, and a mechanism for noninteractive password entry
(such as sshpass or SSH_ASKPASS).

**Security:** The temporary password grants `deck` login AND `sudo`
privileges. While connected, an agent can deliberately extend access;
this is not a security sandbox. Existing SSH sessions may remain running
after new password logins are revoked. A QR code forwarded into Telegram
also persists in the bot's chat history; delete it after use.

**No external QR service:** QR generation uses a bundled small BSD-3-Clause
python-qrcode implementation; its license is included in the script.
