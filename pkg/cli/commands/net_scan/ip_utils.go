package net_scan

import (
	"encoding/binary"
	"iter"
	"net"
	"strings"
)

func iterateOverNet(ipnet *net.IPNet) iter.Seq[net.IP] {
	if ipnet.IP.To4() == nil {
		panic("not ipv4")
	}

	return func(yield func(net.IP) bool) {
		maskOnes, _ := ipnet.Mask.Size()
		curIp := ipv4touint32(ipnet.IP)

		for range 1 << (32 - maskOnes) {
			if !yield(uint32toipv4(curIp)) {
				return
			}

			curIp++
		}
	}
}

func ipv4touint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}

	v := uint32(ip[0]) << 24
	v += uint32(ip[1]) << 16
	v += uint32(ip[2]) << 8
	v += uint32(ip[3])

	return v
}

func uint32toipv4(v uint32) net.IP {
	var ip = make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, v)

	return ip
}

func isIPV6(ip string) bool {
	return strings.Contains(ip, ":")
}
