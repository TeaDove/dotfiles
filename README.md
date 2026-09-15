## Cheat sheet
### Kitty term fix
```shell
infocmp -x xterm-kitty | pssh 2a02:6b8:c02:901:0:fce0:0:2af  'tic -x -o ~/.terminfo /dev/stdin'
```

## Install

```shell
git clone https://github.com/teadove/dotfiles
make install
```
