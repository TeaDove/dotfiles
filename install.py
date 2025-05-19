#!/usr/bin/python3

import platform


allowed_systems = {"Darwin", "Linus"}
allowed_machines = {"arm64", "X86-64"}

def main() -> None:
    uname = platform.uname()
    if uname.system not in allowed_systems:
        raise Exception("No allowed system")

    if uname.machine not in allowed_machines:
        raise Exception("No allowed machine")

    print(f"Installing for {uname.system}: {uname.machine}")

if __name__ == main():
    main()