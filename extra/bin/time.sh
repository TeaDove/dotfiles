#!/bin/bash


COLOR='\e[1;32m'
printf "${COLOR}"
while true; do tput clear; date +"%H : %M : %S" | figlet ; sleep 1; done
