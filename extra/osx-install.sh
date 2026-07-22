#!/bin/bash

brew install tmux git jql lsd dust lazygit fish zsh gopass 2fa curlie wget

# Starship
curl -sS https://starship.rs/install.sh | sh

# 2fa
go install rsc.io/2fa@latest