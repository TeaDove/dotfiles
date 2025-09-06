package net_scan

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPv4ToUInt32(t *testing.T) {
	t.Parallel()

	ip, _, err := net.ParseCIDR("192.168.150.110/24")
	require.NoError(t, err)
	assert.Equal(t, "192.168.150.110", ip.String())

	v := ipv4touint32(ip)
	assert.Equal(t, uint32(0xc0a8966e), v)

	ip = uint32toipv4(v)
	assert.Equal(t, "192.168.150.110", ip.String())
}

func TestIterateOverNet(t *testing.T) {
	t.Parallel()

	_, ipnet, err := net.ParseCIDR("192.168.150.110/23")
	require.NoError(t, err)

	var idx int
	for ip := range iterateOverNet(ipnet) {
		fmt.Printf("idx: %d, ip: %s\n", idx, ip)
		idx++
	}
}
