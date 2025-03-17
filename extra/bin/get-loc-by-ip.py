#!/usr/bin/python3
# -*- coding: utf-8 -*-
import argparse
import ipaddress
import sys

import requests

RED = "\033[91;1m"
RED_B = "\033[91;5m"
YELLOW = "\033[93m"
BLUE = "\033[94m"
CYAN = "\033[36;4m"
MAGENTA = "\033[35;1m"
GREEN = "\033[92m"
ENDC = "\033[0m"


def search_on_ip_api(to_search: ipaddress.IPv4Address | ipaddress.IPv6Address | str) -> None:
    res = requests.get("http://ip-api.com/json/{}".format(to_search), timeout=3)
    if not res.ok:
        sys.exit(f"{RED}Error {res.status_code}{ENDC}")

    res_json = res.json()
    ok = res_json["status"] == "success"
    res_json["status"] = f"{GREEN}{res_json['status']}{ENDC}"
    res_json["query"] = f"{RED}{res_json['query']}{ENDC}"
    if ok:
        link = f"\n{CYAN}https://www.google.com/maps/place/{res_json['lat']}%20" f"{res_json['lon']}{ENDC}"
        res_json["country"] = f"{YELLOW}{res_json['country']}{ENDC}"
        res_json["regionName"] = f"{YELLOW}{res_json['regionName']}{ENDC}"
        res_json["city"] = f"{YELLOW}{res_json['city']}{ENDC}"
        res_json["lat"] = f"{MAGENTA}{res_json['lat']}{ENDC}"
        res_json["lon"] = f"{MAGENTA}{res_json['lon']}{ENDC}"
        res_json["isp"] = f"{YELLOW}{res_json['isp']}{ENDC}"
    else:
        res_json["message"] = f"{RED_B}{res_json['message']}{ENDC}"

    for key, value in res_json.items():
        print(f"\t{key}: {value}")

    if ok:
        print(link)
    else:
        print()


def main() -> None:
    parser = argparse.ArgumentParser(description="Get location by ipv4 or ipv6 from http://ip-api.com/json/")
    parser.add_argument("ip_address", action="store", type=str, help="ipv4 or ipv6 address", nargs=1)
    parser.add_argument("-j", "--json", action="store_true", help="output as json")
    args = parser.parse_args()
    if args.json:
        res = requests.get(f"http://ip-api.com/json/{args.ip_address[0]}", timeout=3)
        print(res.json())
        return

    try:
        ip_to_search = ipaddress.ip_address(args.ip_address[0])
    except ValueError as e:
        print(f"{YELLOW}{e}, assuming it is hostname{ENDC}\n")
        search_on_ip_api(args.ip_address[0])
        return

    if ip_to_search.is_global:
        search_on_ip_api(ip_to_search)
    else:
        sys.exit(f"{RED}This ip address '{ip_to_search}' is not global{ENDC}")


if __name__ == "__main__":
    main()
