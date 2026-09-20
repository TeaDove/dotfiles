#!/bin/bash

brew install tmux git jql yq lsd dust lazygit fish zsh zsh-autosuggestions zsh-syntax-highlighting zsh-history-substring-search fzf gopass 2fa curlie wget cloc curlie tree neovim bat lolcat kitty terraform graphviz

brew install --cask karabiner-elements

# Starship
curl -sS https://starship.rs/install.sh | sh

# Python libs
pip3 install pre-commit --break-system-packages

# Go libs
go install rsc.io/2fa@latest
go install github.com/teadove/goteleout@latest

chsh -s /bin/zsh
