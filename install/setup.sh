#!/bin/bash


apt update
apt upgrade -y

apt install -y python3 python3-pip python3-dev python3-setuptools python3-venv \
  build-essential make git net-tools curl wget vim neovim \
  fish zsh lsd tmux neofetch btop

git config --global credential.helper store

chsh -s $(which fish)
