#!/usr/bin/python3

import platform
import os
import urllib.request
import json

python_to_go_system: dict[str, str] = {"Darwin": "darwin", "Linux": "linux"}
python_to_go_machines: dict[str, str] = {"arm64": "arm64", "X86-64": "amd64"}

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
                    print("Release downloaded!")
                    return



    raise Exception(f"Not found: {system=}, {machine=}")

def install() -> None:
    print()
    os.system("./u install")
    print("Dotfiles installed")

    os.system("mv u ~/.local/bin/")

def main() -> None:
    uname = platform.uname()
    if uname.system not in python_to_go_system:
        raise Exception("No allowed system")

    if uname.machine not in python_to_go_machines:
        raise Exception("No allowed machine")

    print(f"Installing for {uname.system}: {uname.machine}")
    download_release(python_to_go_system[uname.system], python_to_go_machines[uname.machine])
    install()

if __name__ == main():
    main()