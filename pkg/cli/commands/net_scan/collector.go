package net_scan

import (
	"context"
	"fmt"
	netstd "net"
	"strings"
	"sync"
	"time"

	"github.com/endobit/oui"
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
	Mac          string
	Ports        []*PortStats
	PortsChecked uint16
	PortsTotal   uint16
}

type PortStats struct {
	Number  uint16
	Message string
}

func (r *NetSystem) collect(ctx context.Context) {
	mainInterface, addr, err := getMainNetwork(ctx)
	if err != nil {
		r.CollectionMu.Lock()
		defer r.CollectionMu.Unlock()

		r.Collection.Err = errors.WithStack(err)

		return
	}

	r.CollectionMu.Lock()
	r.Collection.Interface = mainInterface.Name
	r.CollectionMu.Unlock()

	_, ipnet, err := netstd.ParseCIDR(addr.Addr)
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

func getMainNetwork(ctx context.Context) (net.InterfaceStat, net.InterfaceAddr, error) {
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return net.InterfaceStat{}, net.InterfaceAddr{}, errors.Wrap(err, "get interfaces")
	}

	if len(interfaces) <= 1 {
		return net.InterfaceStat{}, net.InterfaceAddr{}, errors.New("only loopback found")
	}

	for _, i := range interfaces[1:] {
		for _, addr := range i.Addrs {
			if isIPV6(addr.Addr) {
				continue
			}

			return i, addr, nil
		}
	}

	return net.InterfaceStat{}, net.InterfaceAddr{}, errors.New("no interfaces with addresses found")
}

func (r *NetSystem) checkAddress(ctx context.Context, weighted *semaphore.Weighted, ip netstd.IP) {
	var ok bool

	_ = weighted.Acquire(ctx, 1)
	ok = ping(ctx, ip.String())

	weighted.Release(1)

	if !ok {
		return
	}

	ipStats := &IPStats{IP: ip, Mac: r.getMacDescription(ip), PortsTotal: uint16(len(r.PortsToScan))}

	r.CollectionMu.Lock()
	r.Collection.IPs = append(r.Collection.IPs, ipStats)
	r.CollectionMu.Unlock()

	var wg sync.WaitGroup

	for _, port := range r.PortsToScan {
		_ = weighted.Acquire(ctx, 1)

		wg.Go(func() {
			r.checkPort(ctx, weighted, ipStats, ip, port)
		})
	}

	wg.Wait()
}

func (r *NetSystem) getMacDescription(ip netstd.IP) string {
	mac, ok := r.ARPTable[ip.String()]
	if !ok {
		return ""
	}

	vendor := oui.Vendor(mac)
	if vendor == "" {
		return mac
	}

	return fmt.Sprintf("%s %s", mac, vendor)
}

func (r *NetSystem) checkPort(
	ctx context.Context,
	weighted *semaphore.Weighted,
	ipStats *IPStats,
	ip netstd.IP,
	port uint16,
) {
	defer weighted.Release(1)
	defer func() {
		r.CollectionMu.Lock()

		ipStats.PortsChecked++

		r.CollectionMu.Unlock()
	}()

	dialer := netstd.Dialer{Timeout: 500 * time.Millisecond}

	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return
	}

	defer conn.Close()

	r.CollectionMu.Lock()
	defer r.CollectionMu.Unlock()

	ipStats.Ports = append(ipStats.Ports, &PortStats{Number: port})
}
