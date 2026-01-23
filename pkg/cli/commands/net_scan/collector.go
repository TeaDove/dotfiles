package net_scan

import (
	"context"
	"fmt"
	"iter"
	netstd "net"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/endobit/oui"
	"github.com/shirou/gopsutil/v4/net"
	"golang.org/x/sync/semaphore"
)

type Collection struct {
	Network   string
	Interface string

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

func (r *Service) collect(ctx context.Context, args []string) error {
	mainInterface, addr, err := getMainNetwork(ctx)
	if err != nil {
		return errors.Wrap(err, "get main network")
	}

	ipnet, err := parseIPNet(addr.Addr, args)
	if err != nil {
		return errors.Wrap(err, "parse main address")
	}

	size, _ := ipnet.Mask.Size()

	r.collectionMu.Lock()
	r.collection.Interface = mainInterface.Name
	r.collection.Network = ipnet.String()
	r.collection.IPsTotal = 1 << (32 - size)
	r.collectionMu.Unlock()

	r.scanNet(ctx, iterateOverNet(ipnet))

	return nil
}

func parseIPNet(mainNet string, args []string) (*netstd.IPNet, error) {
	if len(args) == 0 {
		_, ipnet, err := netstd.ParseCIDR(mainNet)
		if err != nil {
			return nil, errors.Wrap(err, "parse main address")
		}

		return ipnet, nil
	}

	ip := netstd.ParseIP(args[0])
	if ip != nil {
		_, ipnet, err := netstd.ParseCIDR(ip.String() + "/32")
		if err != nil {
			return nil, errors.Wrap(err, "parse main address")
		}

		return ipnet, nil
	}

	_, ipnet, err := netstd.ParseCIDR(args[0])
	if err != nil {
		return nil, errors.Wrap(err, "parse main address")
	}

	return ipnet, nil
}

func (r *Service) scanNet(ctx context.Context, ips iter.Seq[netstd.IP]) {
	var (
		wg       sync.WaitGroup
		weighted = semaphore.NewWeighted(50)
	)

	for ip := range ips {
		_ = weighted.Acquire(ctx, 1)

		wg.Go(func() {
			defer weighted.Release(1)

			r.scanIP(ctx, ip)

			r.collectionMu.Lock()
			r.collection.IPsChecked++
			r.collectionMu.Unlock()
		})
	}

	wg.Wait()
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

func (r *Service) scanIP(ctx context.Context, ip netstd.IP) {
	if !ping(ctx, ip.String()) {
		return
	}

	ipStats := &IPStats{IP: ip, Mac: r.getMacDescription(ip), PortsTotal: uint16(len(r.portsToScan))}

	r.collectionMu.Lock()
	r.collection.IPs = append(r.collection.IPs, ipStats)
	r.collectionMu.Unlock()

	var (
		wg       sync.WaitGroup
		weighted = semaphore.NewWeighted(10)
	)

	for _, port := range r.portsToScan {
		_ = weighted.Acquire(ctx, 1)

		wg.Go(func() {
			defer weighted.Release(1)

			r.scanPort(ctx, ipStats, ip, port)
		})
	}

	wg.Wait()
}

func (r *Service) getMacDescription(ip netstd.IP) string {
	mac, ok := r.arpTable[ip.String()]
	if !ok {
		return ""
	}

	vendor := oui.Vendor(mac)
	if vendor == "" {
		return mac
	}

	return fmt.Sprintf("%s %s", mac, vendor)
}

func (r *Service) scanPort(
	ctx context.Context,
	ipStats *IPStats,
	ip netstd.IP,
	port uint16,
) {
	defer func() {
		r.collectionMu.Lock()
		defer r.collectionMu.Unlock()

		ipStats.PortsChecked++
	}()

	conn, err := r.dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return
	}

	defer conn.Close()

	server := r.protoDetection(ctx, ip.String(), port)

	r.collectionMu.Lock()
	defer r.collectionMu.Unlock()

	ipStats.Ports = append(ipStats.Ports, &PortStats{Number: port, Message: server})
}
