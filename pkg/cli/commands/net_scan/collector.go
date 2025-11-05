package net_scan

import (
	"context"
	"fmt"
	netstd "net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/net"
	"golang.org/x/sync/semaphore"
)

type Collection struct {
	Err        error
	Network    string
	Interface  string
	IPs        []*IPStats
	IPsChecked int
	IPsTotal   int
}

type IPStats struct {
	IP           netstd.IP
	Ports        []*PortStats
	PortsChecked uint16
	PortsTotal   uint16
}

type PortStats struct {
	Number  uint16
	Message string
}

func (r *NetSystem) collect(ctx context.Context) {
	mainInterface, err := getMainNetwork(ctx)
	if err != nil {
		r.CollectionMu.Lock()
		defer r.CollectionMu.Unlock()

		r.Collection.Err = errors.WithStack(err)

		return
	}

	r.CollectionMu.Lock()
	r.Collection.Interface = mainInterface.Name
	r.CollectionMu.Unlock()

	_, ipnet, err := netstd.ParseCIDR(mainInterface.Addrs[0].Addr)
	if err != nil {
		r.CollectionMu.Lock()
		defer r.CollectionMu.Unlock()

		r.Collection.Err = errors.Wrap(err, "parse main address")

		return
	}

	r.CollectionMu.Lock()
	r.Collection.Network = ipnet.String()
	size, _ := ipnet.Mask.Size()
	r.Collection.IPsTotal = 1 << (32 - size)
	r.CollectionMu.Unlock()

	var (
		wg       sync.WaitGroup
		weighted = semaphore.NewWeighted(1000)
	)
	for ip := range iterateOverNet(ipnet) {
		wg.Go(func() {
			r.checkAddress(ctx, weighted, ip)

			r.CollectionMu.Lock()
			r.Collection.IPsChecked++
			r.CollectionMu.Unlock()
		})
	}

	wg.Wait()
}

func isIPV6(ip string) bool {
	return strings.Contains(ip, ":")
}

func getMainNetwork(ctx context.Context) (net.InterfaceStat, error) {
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return net.InterfaceStat{}, errors.Wrap(err, "get interfaces")
	}

	if len(interfaces) <= 1 {
		return net.InterfaceStat{}, errors.New("only loopback found")
	}

	var mainInterface net.InterfaceStat

	for _, i := range interfaces[1:] {
		if len(i.Addrs) != 0 && !isIPV6(i.Addrs[0].Addr) {
			mainInterface = i
			break
		}
	}

	if mainInterface.Name == "" {
		return net.InterfaceStat{}, errors.New("no interfaces with addresses found")
	}

	return mainInterface, nil
}

func (r *NetSystem) checkAddress(ctx context.Context, weighted *semaphore.Weighted, ip netstd.IP) {
	var ok bool

	_ = weighted.Acquire(ctx, 1)
	ok = ping(ctx, ip.String())

	weighted.Release(1)

	if !ok {
		return
	}

	const (
		firstPort  = uint16(1)
		lastPort   = uint16(10_000)
		totalPorts = lastPort - firstPort
	)

	ipStats := &IPStats{IP: ip, PortsTotal: totalPorts}

	r.CollectionMu.Lock()
	r.Collection.IPs = append(r.Collection.IPs, ipStats)
	r.CollectionMu.Unlock()

	var wg sync.WaitGroup

	for i := firstPort; i < lastPort; i++ {
		_ = weighted.Acquire(ctx, 1)

		wg.Go(func() {
			defer weighted.Release(1)
			defer func() {
				r.CollectionMu.Lock()

				ipStats.PortsChecked++

				r.CollectionMu.Unlock()
			}()

			dialer := netstd.Dialer{Timeout: 5 * time.Second}

			conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ip, i))
			if err != nil {
				return
			}

			defer conn.Close()

			r.CollectionMu.Lock()

			ipStats.Ports = append(ipStats.Ports, &PortStats{Number: i})

			r.CollectionMu.Unlock()
		})
	}

	wg.Wait()
}
