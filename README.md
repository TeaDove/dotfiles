## Cheat sheet
### Kitty term fix
```shell
infocmp -x xterm-kitty | ssh 192.168.1.1  'tic -x -o ~/.terminfo /dev/stdin'
```

### Linux decrypt
```shell
ecryptfs-mount-private
exec zsh
```

### No pager journalctl
```shell
sudo journalctl -x -u --no-pager -o short-iso
```

### Lintin
```shell
gotestsum --format-hide-empty-pkg -- ./... --race && golangci-lint run -D exhaustruct_v5,exhaustruct
```

## Install

```shell
cd ~
git clone https://github.com/teadove/dotfiles
cd dotfiles
make install
```
