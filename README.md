## Cheat sheet
### Kitty term fix
```shell
infocmp -x xterm-kitty | ssh 192.168.1.1  'tic -x -o ~/.terminfo /dev/stdin'
```
```shell
# Or simpler
export TERM=xterm-256color
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
`Space` — страница вниз
`b` — страница вверх
`d` и `u` — полстраницы вниз и вверх
`G` — в конец, `g` — в начало

Убрать обрезку - нажмите -, потом S. Повторное нажатие возвращает обрезку.

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
