## Cheat sheet
### Kitty term fix
```shell
infocmp -x xterm-kitty | pssh 2a02:6b8:c02:901:0:fce0:0:2af  'tic -x -o ~/.terminfo /dev/stdin'
```

### Linux decrypt
```shell
ecryptfs-mount-private
exec zsh
```

### Host tint in kitty

`.config/kitty/host_tint.py` is a kitty watcher: when a connect command starts in a
window, its tab gets one of 32 colours derived from the remote host key in
`known_hosts`, and the tab goes back to the theme colours when the command ends.
Enable it with `watcher host_tint.py` in `kitty.conf`.

Needs kitty shell integration (the tint is driven by OSC 133 command marks), so it does
not fire for commands started inside tmux or from scripts. Other connect utilities are
added to `TARGET_COMMANDS` in the watcher: command name -> index of the argument that
names the host.

## Install

```shell
cd ~
git clone https://github.com/teadove/dotfiles
cd dotfiles
make install
```
