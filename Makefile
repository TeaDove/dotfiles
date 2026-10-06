GO ?= GO111MODULE=on CGO_ENABLED=0 go


test:
	go tool gotestsum --format-hide-empty-pkg -- ./... --race

install:
	$(GO) install u.go
	$(GO) run . install
	exec zsh

fresh-install-darwin: export PATH := /opt/homebrew/bin:$(PATH)
fresh-install-darwin:
	command -v brew >/dev/null || NONINTERACTIVE=1 /bin/bash -c \
		"$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
	brew install tmux git yq lsd lazygit go btop \
		zsh zsh-autosuggestions zsh-syntax-highlighting zsh-history-substring-search \
		fzf gopass curlie wget cloc tree neovim bat lolcat kitty terraform graphviz
	brew install --cask karabiner-elements
	curl -sS https://starship.rs/install.sh | sh -s -- -y
	curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path
	$$HOME/.cargo/bin/cargo install du-dust jql
	pip3 install pre-commit --break-system-packages
	go install rsc.io/2fa@latest
	go install github.com/teadove/goteleout@latest
	chsh -s /bin/zsh
	$(MAKE) install

fresh-install-ubuntu:
	sudo add-apt-repository -y ppa:longsleep/golang-backports
	curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo gpg --batch --yes --dearmor \
		-o /usr/share/keyrings/hashicorp-archive-keyring.gpg
	echo "deb [arch=$$(dpkg --print-architecture) signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg]" \
		"https://apt.releases.hashicorp.com $$(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
	sudo apt update
	sudo env DEBIAN_FRONTEND=noninteractive apt upgrade -y
	sudo env DEBIAN_FRONTEND=noninteractive apt install -y \
		python3 python3-pip python3-dev python3-setuptools python3-venv \
		build-essential make git net-tools curl wget vim neovim \
		zsh zsh-autosuggestions zsh-syntax-highlighting fzf kitty-terminfo tmux neofetch btop golang-go \
		gopass cloc tree bat lolcat graphviz pre-commit terraform
	[ -d /usr/local/share/zsh-history-substring-search ] || sudo git clone --depth 1 \
		https://github.com/zsh-users/zsh-history-substring-search /usr/local/share/zsh-history-substring-search
	curl -sS https://starship.rs/install.sh | sh -s -- -y
	curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path
	$$HOME/.cargo/bin/cargo install du-dust jql lsd
	go install rsc.io/2fa@latest
	go install github.com/mikefarah/yq/v4@latest
	go install github.com/jesseduffield/lazygit@latest
	go install github.com/rs/curlie@latest
	go install github.com/teadove/goteleout@latest
	git config --global credential.helper store
	chsh -s $$(which zsh)
	$(MAKE) install
