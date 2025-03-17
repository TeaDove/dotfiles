#!/usr/bin/python3

import argparse
import os
import shutil
import subprocess  # noqa: S404
import sys
from pathlib import Path

SCRIPT = Path(os.path.realpath(__file__))
BASE = SCRIPT.parent
HOME = Path.home()
os.chdir(SCRIPT.parent)


CYAN = "\033[96m"
YELLOW = "\033[92m"
ENDC = "\033[0m"


def create_local() -> None:
    local_ = HOME / ".local/bin"
    local_.mkdir(parents=True, exist_ok=True)


def main() -> None:  # noqa: C901, CCR001
    parser = argparse.ArgumentParser(description="Link dotfiles and other config files")
    parser.add_argument(
        "-f",
        "--full",
        action="store_true",
        help="Install all configs and use third party programms such as lsd, dust etc.",
    )
    parser.add_argument(
        "-i",
        "--install",
        action="store_true",
        help="Install software(works on Debian) required by this dotfiles",
    )
    args = parser.parse_args()
    if not (args.full or args.install or args.raw):
        sys.exit("You need to choose some option..., use --full, --install, --raw")
    else:
        if not args.install:
            create_local()
            bin_folder = BASE / "extra/bin/"
            for file in bin_folder.iterdir():
                try:
                    shutil.copy(file, HOME / ".local/bin")
                except shutil.SameFileError:
                    os.remove(HOME / ".local/bin" / file.name)
                    shutil.copy(file, HOME / ".local/bin")
                print(f"{YELLOW}{file.name}", end=" ")
            print(f"\n{ENDC}Were copied")

            to_home = BASE / "configs/home"
            dotconfig = BASE / "configs/.config"

            for file in to_home.iterdir():
                try:
                    shutil.copy(file, HOME)
                except shutil.SameFileError:
                    os.remove(HOME / file.name)
                    shutil.copy(file, HOME)
                print(f"{YELLOW}{file.name}", end=" ")
            print(f"\n{ENDC}Were copied")

            (HOME / ".config").mkdir(exist_ok=True)
            for file in dotconfig.iterdir():
                filename = HOME / ".config" / file.name
                if filename.exists():
                    if filename.is_dir():
                        shutil.rmtree(filename)
                    else:
                        os.remove(filename)

            for file in dotconfig.iterdir():
                if file.is_dir():
                    shutil.copytree(file, HOME / ".config" / file.name, dirs_exist_ok=True)
                else:
                    shutil.copy(file, HOME / ".config")
                print(f"{YELLOW}{file.name}", end=" ")
            print(f"\n{ENDC}Were copied")

        if args.install:
            for file in Path("install").iterdir():
                print(f"{YELLOW}Processing {file.name}{ENDC}")
                subprocess.call(str(file.absolute()))  # noqa: S603


if __name__ == "__main__":
    main()
