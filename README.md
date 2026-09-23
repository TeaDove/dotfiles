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

### SSH host tint in kitty

Inside kitty, `ssh` tints the window background with one of 32 colours picked from the
remote host key in `known_hosts`, and restores the previous colours on exit. Unknown
hosts are tinted by hostname. Disable with `SSH_TINT=0`.

## Install

```shell
cd ~
git clone https://github.com/teadove/dotfiles
cd dotfiles
make install
```
