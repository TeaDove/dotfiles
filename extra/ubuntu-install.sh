#!/bin/bash

add-apt-repository ppa:longsleep/golang-backports
apt update
apt upgrade -y

apt install -y python3 python3-pip python3-dev python3-setuptools python3-venv \
  build-essential make git net-tools curl wget vim neovim \
  fish zsh tmux neofetch btop golang-go

snap install lsd

curl -sS https://starship.rs/install.sh | sh

git config --global credential.helper store

chsh -s $(which fish)
