zmodload zsh/terminfo
if (( ! $+terminfo[cuu1] )); then
    export TERM=xterm-256color
fi

autoload -U colors && colors

USER_COLOR=$fg[cyan]
DIR_COLOR=$fg[green]
GIT_COLOR=$fg[yellow]
RESET=$reset_color

PROMPT="${USER_COLOR}%n${RESET} in ${DIR_COLOR}%~${RESET}
$ "

alias ll='ls -la --color=auto 2&>/dev/null || ls -laG'
alias l='ls -l --color=auto 2&>/dev/null || ls -lG'
alias b="bpython"
alias speed='curl -s https://raw.githubusercontent.com/sivel/speedtest-cli/master/speedtest.py | python3 -B'
alias d='dust'
alias s='kitty +kitten ssh'
alias tm='tmux new-session -A -s main'
alias mac-unquarantine='xattr -d com.apple.quarantine'

alias jup='cd ~/projects/jup && python3.14 -m jupyterlab ; cd -'
alias jup-darwin='cd ~/projects/jup && python3.14 -m jupyterlab --app-dir=/opt/homebrew/share/jupyter/lab ; cd -'
alias ljup='python3.14 -m jupyterlab'
alias ljup-darwin='python3.14 -m jupyterlab --app-dir=/opt/homebrew/share/jupyter/lab'
alias cloc-git='cloc $(git ls-tree -r master --name-only)'

function p () { ps aux | head -n 1 && ps aux | grep -v grep --color=auto | grep $argv }

function cbox () {
    local proj=$PWD
    docker run -it --rm \
        --cap-add=NET_ADMIN --cap-add=NET_RAW \
        -v $proj:$proj \
        -v claude-box-go-mod:/go/pkg/mod \
        -v claude-box-go-build:/home/dev/.cache/go-build \
        -v claude-box-pip:/home/dev/.cache/pip \
        -v claude-box-precommit:/home/dev/.cache/pre-commit \
        -v claude-box-home:/home/dev/.claude \
        -v $HOME/.claude/CLAUDE.md:/home/dev/.claude/CLAUDE.md:ro \
        -v $HOME/.claude/skills:/home/dev/.claude/skills:ro \
        -v $HOME/.claude/projects:/home/dev/.claude/projects \
        -w $proj \
        claude-box \
        bash -c "sudo /usr/local/bin/init-firewall.sh && exec claude --dangerously-skip-permissions $*"
}

typeset -U path
path=(
    /opt/homebrew/opt/*/libexec/gnubin(N)
    /opt/homebrew/bin
    /opt/homebrew/sbin
    $HOME/.local/bin
    /Applications/kitty.app/Contents/MacOS(N)
    $HOME/Applications/kitty.app/Contents/MacOS(N)
    /usr/local/go/bin
    $HOME/.cargo/bin
    $path
    $HOME/Library/Python/3.8/bin
    $HOME/yandex-cloud/bin
    $HOME/ycp/bin
    $HOME/ydb/bin
    $HOME/go/bin
    $HOME/go/bin/darwin_amd64
    /opt/homebrew/opt/libpq/bin
    $HOME/.krew/bin
)
export PATH
export HOMEBREW_NO_AUTO_UPDATE=1

if (( $+commands[lsd] )); then
    alias l='lsd -lh --blocks=permission,user,size,date,name'
    alias ll='lsd -lha --blocks=permission,user,size,date,name'
fi

HISTFILE=$HOME/.zsh_history
HISTSIZE=100000
SAVEHIST=100000
setopt EXTENDED_HISTORY INC_APPEND_HISTORY HIST_IGNORE_ALL_DUPS HIST_IGNORE_SPACE HIST_REDUCE_BLANKS HIST_VERIFY
setopt AUTO_CD INTERACTIVE_COMMENTS COMPLETE_IN_WORD NO_LIST_AMBIGUOUS NO_BEEP

[[ $COLORTERM == (truecolor|24bit) ]] || zmodload zsh/nearcolor

completions_cache=${XDG_CACHE_HOME:-$HOME/.cache}/zsh/completions
mkdir -p $completions_cache
for completion_command in 'stern --completion zsh' 'poetry completions zsh'; do
    completion_file=$completions_cache/_${completion_command%% *}
    if (( $+commands[${completion_command%% *}] )) && [[ ! -s $completion_file ]]; then
        ${=completion_command} > $completion_file
    fi
done
fpath=($completions_cache $fpath)
unset completions_cache completion_command completion_file

autoload -Uz compinit && compinit
zstyle ':completion:*' menu select
zstyle ':completion:*' matcher-list '' 'm:{a-zA-Z}={A-Za-z}' 'l:|=* r:|=*'
zstyle ':completion:*' list-colors ${(s.:.)LS_COLORS} '=(#b)*(-- *)=0=38;2;188;175;139' 'ma=7'
zstyle ':completion:*' use-cache on

bindkey -e
WORDCHARS=${WORDCHARS//[\/.=-]}

if (( $+commands[fzf] )); then
    fzf_integration=$(fzf --zsh 2>/dev/null)
    if [[ -n $fzf_integration ]]; then
        eval $fzf_integration
    else
        for fzf_script in /usr/share/doc/fzf/examples/{key-bindings,completion}.zsh /usr/share/fzf/{key-bindings,completion}.zsh; do
            if [[ -r $fzf_script ]]; then
                source $fzf_script
            fi
        done
    fi
    unset fzf_integration fzf_script
fi

ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE='fg=#BD93F9'
HISTORY_SUBSTRING_SEARCH_HIGHLIGHT_FOUND='fg=11,bg=8'
HISTORY_SUBSTRING_SEARCH_ENSURE_UNIQUE=1

for plugin in zsh-autosuggestions zsh-syntax-highlighting zsh-history-substring-search; do
    for plugin_root in /usr/share /usr/share/zsh/plugins /usr/local/share /opt/homebrew/share $HOME/.local/share/zsh/plugins; do
        if [[ -r $plugin_root/$plugin/$plugin.zsh ]]; then
            source $plugin_root/$plugin/$plugin.zsh
            break
        fi
    done
done
unset plugin plugin_root

typeset -A ZSH_HIGHLIGHT_STYLES
ZSH_HIGHLIGHT_STYLES[default]='fg=#D6DAE4'
ZSH_HIGHLIGHT_STYLES[single-hyphen-option]='fg=#D6DAE4'
ZSH_HIGHLIGHT_STYLES[double-hyphen-option]='fg=#D6DAE4'
ZSH_HIGHLIGHT_STYLES[path]='fg=#D6DAE4,underline'
ZSH_HIGHLIGHT_STYLES[path_prefix]='fg=#D6DAE4,underline'
ZSH_HIGHLIGHT_STYLES[command]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[builtin]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[alias]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[suffix-alias]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[global-alias]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[function]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[precommand]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[hashed-command]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[reserved-word]='fg=#2F9FAE'
ZSH_HIGHLIGHT_STYLES[autodirectory]='fg=#2F9FAE,underline'
ZSH_HIGHLIGHT_STYLES[unknown-token]='fg=#FFB86C'
ZSH_HIGHLIGHT_STYLES[single-quoted-argument]='fg=#F1FA8C'
ZSH_HIGHLIGHT_STYLES[double-quoted-argument]='fg=#F1FA8C'
ZSH_HIGHLIGHT_STYLES[dollar-quoted-argument]='fg=#F1FA8C'
ZSH_HIGHLIGHT_STYLES[back-double-quoted-argument]='fg=#00A6B2'
ZSH_HIGHLIGHT_STYLES[back-dollar-quoted-argument]='fg=#00A6B2'
ZSH_HIGHLIGHT_STYLES[dollar-double-quoted-argument]='fg=#00A6B2'
ZSH_HIGHLIGHT_STYLES[globbing]='fg=#00A6B2'
ZSH_HIGHLIGHT_STYLES[history-expansion]='fg=#00A6B2'
ZSH_HIGHLIGHT_STYLES[commandseparator]='fg=#50FA7B'
ZSH_HIGHLIGHT_STYLES[redirection]='fg=#8BE9FD'
ZSH_HIGHLIGHT_STYLES[comment]='fg=#6272A4'

if (( $+widgets[history-substring-search-up] )); then
    history_search_up=history-substring-search-up
    history_search_down=history-substring-search-down
else
    autoload -Uz up-line-or-beginning-search down-line-or-beginning-search
    zle -N up-line-or-beginning-search
    zle -N down-line-or-beginning-search
    history_search_up=up-line-or-beginning-search
    history_search_down=down-line-or-beginning-search
fi
bindkey '^[[A' $history_search_up
bindkey '^[OA' $history_search_up
bindkey '^[[B' $history_search_down
bindkey '^[OB' $history_search_down
unset history_search_up history_search_down

function toggle-sudo-prefix () {
    if [[ -z $BUFFER ]]; then
        BUFFER=$(fc -ln -1)
    fi
    if [[ $BUFFER == sudo\ * ]]; then
        BUFFER=${BUFFER#sudo }
    else
        BUFFER="sudo $BUFFER"
    fi
    CURSOR=$#BUFFER
}
zle -N toggle-sudo-prefix
autoload -Uz edit-command-line
zle -N edit-command-line

bindkey '^[s' toggle-sudo-prefix
bindkey '^[e' edit-command-line
bindkey '^[v' edit-command-line
bindkey '^[[Z' reverse-menu-complete
bindkey '^[[1;5C' forward-word
bindkey '^[[1;5D' backward-word
bindkey '^[[1;3C' forward-word
bindkey '^[[1;3D' backward-word
bindkey '^[[3~' delete-char
bindkey '^[[H' beginning-of-line
bindkey '^[[F' end-of-line
bindkey '^[[1~' beginning-of-line
bindkey '^[[4~' end-of-line

if (( $+commands[starship] )); then
    eval "$(starship init zsh)"
fi

if [[ -n $KITTY_INSTALLATION_DIR ]]; then
    export KITTY_SHELL_INTEGRATION=enabled
    autoload -Uz -- $KITTY_INSTALLATION_DIR/shell-integration/zsh/kitty-integration
    kitty-integration
    unfunction kitty-integration
fi

source /Users/teadove/yandex-cloud/completion.zsh.inc
