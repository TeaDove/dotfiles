#!/bin/bash

cd /tmp
wget https://github.com/sharkdp/bat/releases/download/v0.18.0/bat_0.18.0_amd64.deb
sudo dpkg -i bat_0.18.0_amd64.deb
mkdir -p ~/.local/bin
ln -s /usr/bin/batcat ~/.local/bin/bat
rm bat_0.18.0_amd64.deb
cd -
