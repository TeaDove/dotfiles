# Colors

set fish_color_command 2F9FAE
set fish_pager_color_selected_prefix black --underline
set fish_pager_color_selected_completion black
set fish_pager_color_description BCAF8B
set fish_pager_color_selected_description A24000
set fish_color_param D6DAE4

# Aliases
alias l="lsd -lh --blocks=permission,user,size,date,name"
alias ll="lsd -lha --blocks=permission,user,size,date,name"
alias d='dust'
alias b="bpython"
alias s='source .venv/bin/activate'

alias speed='curl -s https://raw.githubusercontent.com/sivel/speedtest-cli/master/speedtest.py | python3 -B'

alias jup='cd ~/projects/jup && python3.13 -m jupyterlab ; cd -'
alias jup-darwin='cd ~/projects/jup && python3.13 -m jupyterlab --app-dir=/opt/homebrew/share/jupyter/lab ; cd -'
alias ljup='python3.13 -m jupyterlab'
alias ljup-darwin='python3.13 -m jupyterlab --app-dir=/opt/homebrew/share/jupyter/lab'
alias cloc-git='cloc (git ls-tree -r master --name-only)'

alias kubectl="kubecolor"
alias kwatch='u watch -i=1s "kubecolor --force-colors config view --minify -o jsonpath={..namespace}" "kubecolor --force-colors get deployments -o=custom-columns=DEPLOYMENT:.metadata.name,CONTAINER_IMAGE:.spec.template.spec.containers[*].image,READY_REPLICAS:.status.readyReplicas" "kubecolor --force-colors get statefulset -o=custom-columns=DEPLOYMENT:.metadata.name,CONTAINER_IMAGE:.spec.template.spec.containers[*].image,READY_REPLICAS:.status.readyReplicas" "kubecolor --force-colors get pods"'

function p
     ps aux | head -n 1 && ps aux | grep -v grep --color=auto | grep $argv
end

function kexec 
	kubectl exec -it $(kubectl get pod -o custom-columns=CONTAINER:.metadata.name | grep $argv[1]) -- /bin/bash
end

function envsource
    . (sed 's/^/export /' .env | psub)
end

function sss
    .
end

# Haskell PATH
set -q GHCUP_INSTALL_BASE_PREFIX[1]; or set GHCUP_INSTALL_BASE_PREFIX $HOME
test -f /home/teadove/.ghcup/env ; and set -gx PATH $HOME/.cabal/bin /home/teadove/.ghcup/bin $PATH

set PATH $HOME/.local/bin /usr/local/go/bin $HOME/go/bin  $HOME/.cargo/bin $PATH $HOME/Library/Python/3.8/bin /opt/homebrew/bin $HOME/yandex-cloud/bin $HOME/ydb/bin $HOME/projects/flutter/bin $HOME/go/bin $HOME/go/bin/darwin_amd64 $HOME/.spoof-dpi/bin /opt/homebrew/opt/libpq/bin
set HOMEBREW_NO_AUTO_UPDATE 1

# Starship init
starship init fish | source
