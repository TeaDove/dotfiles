#!/bin/bash


apt update
apt upgrade -y

apt install -y python3 python3-pip python3-dev python3-setuptools python3-venv \
  neofetch tmux zsh git curl wget vim build-essential net-tools make neovim btop

git config --global credential.helper store

chsh -s $(which fish)