#!/bin/bash

brew install tmux git jql lsd dust lazygit fish zsh gopass 2fa curlie wget cloc curlie tree

# Starship
curl -sS https://starship.rs/install.sh | sh

# Python libs
pip3 install pre-commit --break-system-packages

# Go libs
go install rsc.io/2fa@latest
go install github.com/teadove/goteleout@latest
