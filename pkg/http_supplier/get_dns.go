// Package http_supplier
// Copy pasted from https://github.com/qdm12/dns/blob/v2.0.0-beta/pkg/nameserver/getlocal_unix.go
package http_supplier

import (
	"net/netip"
	"os"
	"strings"
)

func GetDNSServers() []netip.AddrPort {
	const filename = "/etc/resolv.conf"
	return getLocalNameservers(filename)
}

func getLocalNameservers(filename string) []netip.AddrPort {
	const defaultNameserverPort = 53
	defaultLocalNameservers := []netip.AddrPort{
		netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), defaultNameserverPort),
		netip.AddrPortFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 1}), defaultNameserverPort),
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return defaultLocalNameservers
	}

	var nameservers []netip.AddrPort

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "nameserver" {
			continue
		}

		for _, field := range fields[1:] {
			ip, err := netip.ParseAddr(field)
			if err != nil {
				continue
			}

			nameservers = append(nameservers,
				netip.AddrPortFrom(ip, defaultNameserverPort))
		}
	}

	if len(nameservers) == 0 {
		return defaultLocalNameservers
	}

	return nameservers
}
