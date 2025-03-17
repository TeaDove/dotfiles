PROMPT='%F{132}%n%f in %F{126}%~%f -> '

alias ll='ls -la --color=auto 2&>/dev/null || ls -laG'
alias l='ls -l --color=auto 2&>/dev/null || ls -lG'
alias g="get-loc-by-ip.py"
alias tb="nc termbin.com 9999"
alias i="ipython"
alias b="bpython"
alias weather='curl "http://wttr.in/Moscow?0" '
alias key='openssl rand -hex 32'
alias key2='pwgen -1 30'
alias speed='curl -s https://raw.githubusercontent.com/sivel/speedtest-cli/master/speedtest.py | python3 -B'
alias d='dust -X=.git'

function sysgrep () {
    systemctl list-units --type=service | head -n 1 && systemctl list-units --type=service | grep $argv
}
function p () { ps aux | head -n 1 && ps aux | grep -v grep --color=auto | grep $argv }
function s () {
    if [ "$#" -eq 1 ]; then
    	source $argv/bin/activate
    else
	    source .venv/bin/activate
    fi
}

export PATH=$HOME/.local/bin:/usr/local/go/bin:$HOME/.cargo/bin:/opt/homebrew/bin:$HOME/go/bin:$PATH
