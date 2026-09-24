import binascii
import colorsys
import os
import shlex
import subprocess
from typing import TYPE_CHECKING, Any

from kitty.utils import log_error

if TYPE_CHECKING:
    from kitty.boss import Boss
    from kitty.window import Window

SSH_COMMANDS: frozenset[str] = frozenset({'autossh', 'mosh', 'pssh', 's', 'ssh'})
TARGET_COMMANDS: dict[str, int] = {}
COMMAND_PREFIXES: frozenset[str] = frozenset(
    {'command', 'doas', 'env', 'exec', 'nohup', 'sudo', 'time'}
)
KITTEN_COMMANDS: frozenset[str] = frozenset({'kitten', 'kitty'})
SATURATION: float = 0.45
LIGHTNESS: tuple[float, float] = (0.28, 0.42)
ACTIVE_FOREGROUND: str = '#f2f2f2'
INACTIVE_FOREGROUND: str = '#c8c8c8'
RESOLVE_TIMEOUT: float = 2.0

host_key_cache: dict[str, str] = {}
tinted_windows: set[int] = set()


def tint(key: str) -> str:
    n = binascii.crc32(key.encode()) % 32
    hue = (n % 16) * 22.5 / 360.0
    channels = colorsys.hls_to_rgb(hue, LIGHTNESS[n // 16], SATURATION)
    return '#' + ''.join(f'{round(channel * 255):02x}' for channel in channels)


def output_of(argv: list[str]) -> str:
    try:
        result = subprocess.run(
            argv, capture_output=True, text=True, timeout=RESOLVE_TIMEOUT
        )
    except (OSError, subprocess.SubprocessError) as err:
        log_error(f'host_tint: {argv[0]} failed: {err}')
        return ''
    return result.stdout


def ssh_settings(args: list[str]) -> dict[str, list[str]]:
    settings: dict[str, list[str]] = {}
    for line in output_of(['ssh', '-G', *args]).splitlines():
        fields = line.split()
        if fields:
            settings.setdefault(fields[0], fields[1:])
    return settings


def ssh_target(settings: dict[str, list[str]]) -> str | None:
    alias = settings.get('hostkeyalias') or settings.get('hostname') or []
    if not alias:
        return None
    port = (settings.get('port') or ['22'])[0]
    return alias[0] if port == '22' else f'[{alias[0]}]:{port}'


def known_host_keys(target: str, settings: dict[str, list[str]]) -> str | None:
    keys: list[str] = []
    hostfiles = settings.get('userknownhostsfile', [])
    hostfiles += settings.get('globalknownhostsfile', [])
    for path in hostfiles:
        expanded = os.path.expanduser(path.replace('%d', '~'))
        if not os.path.exists(expanded):
            continue
        found = output_of(['ssh-keygen', '-F', target, '-f', expanded])
        for line in found.splitlines():
            fields = line.split()
            if not line.startswith('#') and len(fields) > 2:
                keys.append(fields[2])
    return ','.join(sorted(keys)) if keys else None


def positional(args: list[str], index: int) -> str | None:
    positionals = [arg for arg in args if not arg.startswith('-')]
    return positionals[index] if len(positionals) > index else None


def command_and_args(argv: list[str]) -> tuple[str, list[str]]:
    while argv and ('=' in argv[0] or os.path.basename(argv[0]) in COMMAND_PREFIXES):
        argv = argv[1:]
    if not argv:
        return '', []
    name, args = os.path.basename(argv[0]), argv[1:]
    if name in KITTEN_COMMANDS:
        while args and args[0] in ('+', '+kitten'):
            args = args[1:]
        if not args:
            return '', []
        return args[0], args[1:]
    return name, args


def split_cmdline(cmdline: str) -> list[str]:
    try:
        argv = shlex.split(cmdline)
    except ValueError:
        return []
    for separator in ('|', '&&', '||', ';'):
        if separator in argv:
            argv = argv[: argv.index(separator)]
    return argv


def connection_key(cmdline: str) -> str | None:
    if cmdline in host_key_cache:
        return host_key_cache[cmdline]
    name, args = command_and_args(split_cmdline(cmdline))
    if name in SSH_COMMANDS:
        settings = ssh_settings(args)
        target = ssh_target(settings)
        if target is not None:
            keys = known_host_keys(target, settings)
            if keys is not None:
                host_key_cache[cmdline] = keys
                return keys
            return target
    if name in TARGET_COMMANDS:
        return positional(args, TARGET_COMMANDS[name])
    if name in SSH_COMMANDS:
        return positional(args, 0)
    return None


def connect_process_running(window: 'Window') -> bool:
    for process in window.child.foreground_processes:
        argv = process.get('cmdline') or []
        name, _ = command_and_args(list(argv))
        if name in SSH_COMMANDS or name in TARGET_COMMANDS:
            return True
    return False


def paint_tab(boss: 'Boss', window: 'Window', color: str | None) -> None:
    background = color or 'NONE'
    active = ACTIVE_FOREGROUND if color else 'NONE'
    inactive = INACTIVE_FOREGROUND if color else 'NONE'
    boss.call_remote_control(
        window,
        (
            'set-tab-color',
            '--self',
            f'active_bg={background}',
            f'inactive_bg={background}',
            f'active_fg={active}',
            f'inactive_fg={inactive}',
        ),
    )


def on_cmd_startstop(boss: 'Boss', window: 'Window', data: dict[str, Any]) -> None:
    try:
        key = connection_key(data.get('cmdline') or '')
        if key is None:
            return
        if data.get('is_start'):
            paint_tab(boss, window, tint(key))
            tinted_windows.add(window.id)
            return
        if window.id in tinted_windows and not connect_process_running(window):
            paint_tab(boss, window, None)
            tinted_windows.discard(window.id)
    except Exception as err:
        log_error(f'host_tint: {err}')


def on_close(boss: 'Boss', window: 'Window', data: dict[str, Any]) -> None:
    if window.id not in tinted_windows:
        return
    tinted_windows.discard(window.id)
    try:
        paint_tab(boss, window, None)
    except Exception as err:
        log_error(f'host_tint: {err}')
