set gnubin_path "(find -L "/opt/homebrew/opt" -name gnubin | sort | tr '\n' :)"

set PATH $gnubin_path/opt/homebrew/bin /opt/homebrew/sbin $HOME/.local/bin /usr/local/go/bin $HOME/.cargo/bin $PATH $HOME/Library/Python/3.8/bin /opt/homebrew/bin $HOME/yandex-cloud/bin $HOME/ycp/bin $HOME/ydb/bin $HOME/go/bin $HOME/go/bin/darwin_amd64 /opt/homebrew/opt/libpq/bin
set -gx PATH $PATH $HOME/.krew/bin
