#!/usr/bin/python3

import platform
import os
import urllib.request
import json

install_location = '~/.local/bin/'
python_to_go_system: dict[str, str] = {"darwin": "darwin", "linux": "linux"}
python_to_go_machines: dict[str, str] = {"arm64": "arm64", "aarch64": "arm64", "x86-64": "amd64", "x86_64": "amd64"}

def download_release(system: str, machine: str) -> None:
    with urllib.request.urlopen('https://api.github.com/repos/teadove/dotfiles/releases/latest') as response:
       json_response = json.loads(response.read())

    with urllib.request.urlopen(json_response['assets_url']) as response:
          json_response = json.loads(response.read())

    for asset in json_response:
       parts = asset['name'].split('-')
       if system == parts[2] and machine == parts[3]:
           with urllib.request.urlopen(asset['browser_download_url']) as response:
                with open("u", "wb") as f:
                    f.write(response.read())
                    print(f"Release {asset['browser_download_url']} downloaded!")
                    return



    raise Exception(f"Not found: {system=}, {machine=}")

def install() -> None:
    print()
    os.system("chmod u+x u")
    os.system("./u install")
    print("Dotfiles installed")

    os.system(f"mv u {install_location}")

def main() -> None:
    uname = platform.uname()
    system = uname.system.lower()
    machine = uname.machine.lower()

    if system not in python_to_go_system:
        raise Exception(f"No allowed system: {system}")

    if machine not in python_to_go_machines:
        raise Exception(f"No allowed machine: {machine}")

    print(f"Installing for {system}: {machine}")
    download_release(python_to_go_system[system], python_to_go_machines[machine])
    install()

if __name__ == main():
    main()
