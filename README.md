![image](https://i.imgur.com/XvMhlOZ.png)
# Dotfiles
Установка:
`./dotfiles_setup.py --help`
> Внимание, помимо конфигов, тут собрано куча мелких полезных и не очень скриптов, темплейтов и тд
`dotfiles_setup.sh` для создание софтлинка из dotfile'ов в ~, ~/.config и тд

# extra
## bin
- `get-loc-by-ip.py` python3 скприпт для получения локации айпишника из [БД](ip-api.com/)
- `wg.py` включение и выключение Wireguard
- `my-ip.sh` получение ip адреса через [этот сайт](https://tesseract.club:8000/docs)
- `new-ssh.sh` см. ниже
- `weaboo.py` смена обоев в гноме
- `time.sh` зелёные аски часы
- `security-logs.sh` показывает открытые порты, текущих юзеров и принятые попыткы подсоединения по ssh
# Терминальный утилы:
- [Bat](https://github.com/sharkdp/bat) - улучшенный cat<br>
Установка:
``` bash
wget https://github.com/sharkdp/bat/releases/download/v0.18.0/bat_0.18.0_amd64.deb
sudo dpkg -i bat_0.18.0_amd64.deb  # adapt version number and architecture
```
- [LSD](https://github.com/Peltoche/lsd) - улучшенный ls<br>
Установка:
``` bash
sudo snap install lsd
```
- [Starship](starship.rs/) - улучшенный prompt
- [Cheat](https://github.com/cheat/cheat) - шпоры!

# Полезное
### Новый ssh.sh
<code>new-ssh.sh server_name username@address</code><br>
## Лучшие ссылки, конфиги и шелы, что используются.
### General
- [Drop-down терминал Guake](http://guake-project.org/)
- [Простой, но красивый shell из коробки "fish"](https://fishshell.com/)
- [Grub theme](https://www.gnome-look.org/p/1420727/)
### Cinnamon spices
- [Windows list](https://cinnamon-spices.linuxmint.com/applets/view/287)
- [System monitor in tray](https://cinnamon-spices.linuxmint.com/applets/view/88)
- [CPU temperature indicator](https://cinnamon-spices.linuxmint.com/applets/view/106)
# Система
![image](https://i.imgur.com/DS8hfDZ.png)
