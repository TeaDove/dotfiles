#!/bin/bash

sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt upgrade -y

sudo apt install -y python3 python3-pip python3-dev python3-setuptools python3-venv \
  build-essential make git net-tools curl wget vim neovim \
  fish zsh zsh-autosuggestions zsh-syntax-highlighting fzf kitty-terminfo tmux neofetch btop golang-go

sudo snap install lsd

git clone --depth 1 https://github.com/zsh-users/zsh-history-substring-search /usr/local/share/zsh-history-substring-search

curl -sS https://starship.rs/install.sh | sh

go install github.com/teadove/goteleout@latest

git config --global credential.helper store

chsh -s $(which zsh)
