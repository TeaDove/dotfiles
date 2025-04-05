package net_stats

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/net"
	net2 "net"
	"strings"
	"sync"
)

func (r *NetStats) interfacesView(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		r.model.interfaces = prettyErr(errors.Wrap(err, "failed to get my-ip"))
		return
	}

	interfacesWithAddresses := make(net.InterfaceStatList, 0)

	for _, i := range interfaces {
		hasIpV4Add := false
		for _, addr := range i.Addrs {
			ip, _, err := net2.ParseCIDR(addr.Addr)
			if err == nil && ip != nil && ip.To4() != nil {
				hasIpV4Add = true
			}
		}

		if !hasIpV4Add {
			continue
		}

		interfacesWithAddresses = append(interfacesWithAddresses, i)
	}

	if len(interfacesWithAddresses) == 0 {
		r.model.interfaces = prettyWarn(errors.New("no interfaces found"))
	}

	r.model.interfaces = color.GreenString("Interfaces with addresses:")

	for _, i := range interfacesWithAddresses {
		addresses := make([]string, 0, len(i.Addrs))
		for _, a := range i.Addrs {
			addresses = append(addresses, a.Addr)
		}

		r.model.interfaces +=
			fmt.Sprintf("\n%s (%s) -> %s",
				color.New(color.FgCyan, color.Faint).Sprintf(i.Name),
				i.HardwareAddr,
				strings.Join(addresses, ", "),
			)
	}
}
