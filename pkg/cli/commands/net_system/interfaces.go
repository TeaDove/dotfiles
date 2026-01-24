package net_system

import (
	"context"
	"fmt"
	net2 "net"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v4/net"
)

func (r *Service) interfacesView(ctx context.Context) string {
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return prettyErr(errors.Wrap(err, "get my-ip"))
	}

	interfacesWithAddresses := make(net.InterfaceStatList, 0)

	for _, i := range interfaces {
		hasIPV4Add := false

		for _, addr := range i.Addrs {
			ip, _, err := net2.ParseCIDR(addr.Addr)
			if err == nil && ip != nil && ip.To4() != nil {
				hasIPV4Add = true
			}
		}

		if !hasIPV4Add {
			continue
		}

		interfacesWithAddresses = append(interfacesWithAddresses, i)
	}

	if len(interfacesWithAddresses) == 0 {
		return prettyWarn(errors.New("no interfaces found"))
	}

	v := color.GreenString("Interfaces with addresses:")

	for _, i := range interfacesWithAddresses {
		addresses := make([]string, 0, len(i.Addrs))
		for _, a := range i.Addrs {
			addresses = append(addresses, a.Addr)
		}

		v += fmt.Sprintf("\n%s (%s) -> %s",
			color.New(color.FgCyan, color.Faint).Sprint(i.Name),
			i.HardwareAddr,
			strings.Join(addresses, ", "),
		)
	}

	return v
}
