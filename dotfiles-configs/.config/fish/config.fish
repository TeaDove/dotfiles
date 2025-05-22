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
alias i="ipython"
alias b="bpython"

alias speed='curl -s https://raw.githubusercontent.com/sivel/speedtest-cli/master/speedtest.py | python3 -B'
alias d='dust -X=.git'

alias jup='cd ~/projects/jup && python3.13 -m jupyterlab ; cd -'
alias jup-darwin='cd ~/projects/jup && python3.13 -m jupyterlab --app-dir=/opt/homebrew/share/jupyter/lab ; cd -'
alias ljup='python3.13 -m jupyterlab'
alias ljup-darwin='python3.13 -m jupyterlab --app-dir=/opt/homebrew/share/jupyter/lab'


alias kubectl="kubecolor"

function sysgrep
    systemctl list-units --type=service | head -n 1 && systemctl list-units --type=service | grep $argv
end
function p
     ps aux | head -n 1 && ps aux | grep -v grep --color=auto | grep $argv
end
function s
    if count $argv > /dev/null
    	source $argv/bin/activate.fish
    else
	    source .venv/bin/activate.fish
    end
end

function gitauto
    echo "git add ."
    git add . || exit 1
    if [ "$argv[1]" ]
        echo git commit -m "$argv[1]" $argv[2]
        git commit -m "$argv[1]" $argv[2] || git add . && git commit -m "$argv[1]" $argv[2]
    else
        echo 'git commit -m "auto: autocommit"'
        git commit -m "auto: autocommit" || git add . && git commit -m "auto: autocommit"
    end
    echo "git push"
    git push
end

function kwatch
    if [ "$argv[1]" ]
        watch -n 0.5 "kubecolor config view --minify -o jsonpath='{..namespace}' && echo && kubecolor get deployments -o='custom-columns=DEPLOYMENT:.metadata.name,CONTAINER_IMAGE:.spec.template.spec.containers[*].image,READY_REPLICAS:.status.readyReplicas' | grep $argv[1] && kubectl get statefulset -o='custom-columns=DEPLOYMENT:.metadata.name,CONTAINER_IMAGE:.spec.template.spec.containers[*].image,READY_REPLICAS:.status.readyReplicas' | grep $argv[1] && echo && kubecolor get pods | grep $argv[1] && echo && kubecolor get events | grep $argv[1] | tail -n 10"
    else
        watch -n 0.5 "kubecolor config view --minify -o jsonpath='{..namespace}' && echo && kubecolor get deployments -o='custom-columns=DEPLOYMENT:.metadata.name,CONTAINER_IMAGE:.spec.template.spec.containers[*].image,READY_REPLICAS:.status.readyReplicas' && kubectl get statefulset -o='custom-columns=DEPLOYMENT:.metadata.name,CONTAINER_IMAGE:.spec.template.spec.containers[*].image,READY_REPLICAS:.status.readyReplicas' && echo && kubecolor get pods && echo && kubecolor get events | tail -n 10"
    end
end

function kexec 
	kubectl exec -it $(kubectl get pod -o custom-columns=CONTAINER:.metadata.name | grep $argv[1]) -- /bin/bash
end

function envsource
    . (sed 's/^/export /' .env | psub)
end

function rsa
	openssl genrsa -out $argv[1]-private.pem 2048
	openssl rsa -in $argv[1]-private.pem -outform PEM -pubout -out $argv[1]-public.pem
end

#if test -z "$(pgrep ssh-agent)"
#    eval (ssh-agent -c)
#end
#eval (ssh-agent -c)
#set -Ux SSH_AUTH_SOCK $SSH_AUTH_SOCK
#set -Ux SSH_AGENT_PID $SSH_AGENT_PID
#set -Ux SSH_AUTH_SOCK $SSH_AUTH_SOCK


# Haskell PATH
set -q GHCUP_INSTALL_BASE_PREFIX[1]; or set GHCUP_INSTALL_BASE_PREFIX $HOME
test -f /home/teadove/.ghcup/env ; and set -gx PATH $HOME/.cabal/bin /home/teadove/.ghcup/bin $PATH

set PATH $HOME/.local/bin /usr/local/go/bin $HOME/go/bin  $HOME/.cargo/bin $PATH $HOME/Library/Python/3.8/bin /opt/homebrew/bin $HOME/yandex-cloud/bin $HOME/ydb/bin $HOME/projects/flutter/bin $HOME/go/bin $HOME/go/bin/darwin_amd64 $HOME/.spoof-dpi/bin /opt/homebrew/opt/libpq/bin

# Starship init
starship init fish | source
