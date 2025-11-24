autoload -U colors && colors

USER_COLOR=$fg[cyan]
DIR_COLOR=$fg[green]
GIT_COLOR=$fg[yellow]
RESET=$reset_color

PROMPT="${USER_COLOR}%n${RESET} in ${DIR_COLOR}%~${RESET}
$ "

alias ll='ls -la --color=auto 2&>/dev/null || ls -laG'
alias l='ls -l --color=auto 2&>/dev/null || ls -lG'
alias i="ipython"
alias b="bpython"
alias speed='curl -s https://raw.githubusercontent.com/sivel/speedtest-cli/master/speedtest.py | python3 -B'
alias d='dust'
alias s='source .venv/bin/activate'

function p () { ps aux | head -n 1 && ps aux | grep -v grep --color=auto | grep $argv }
export PATH=$HOME/.local/bin:/usr/local/go/bin:$HOME/.cargo/bin:/opt/homebrew/bin:$HOME/go/bin:$PATH
