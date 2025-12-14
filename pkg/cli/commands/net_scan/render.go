package net_scan

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/fatih/color"
)

func (r *Model) renderInterface() string {
	return fmt.Sprintf(
		"Scanning TCP ports %s on %s (%d/%d ips)",
		color.CyanString(r.net.Collection.Network),
		color.CyanString(r.net.Collection.Interface),
		r.net.Collection.IPsChecked,
		r.net.Collection.IPsTotal,
	)
}

func (r *Model) View() string {
	r.net.CollectionMu.RLock()
	defer r.net.CollectionMu.RUnlock()

	return lipgloss.JoinVertical(lipgloss.Left, r.renderInterface(), r.renderIPList()) + r.helpView()
}

func (r *Model) renderIPList() string {
	if len(r.net.Collection.IPs) == 0 {
		return "no pingable addresses found"
	}

	slices.SortFunc(r.net.Collection.IPs, func(a, b *IPStats) int {
		if a.IP.String() > b.IP.String() {
			return 1
		}

		return -1
	})

	ipList := list.New()
	for _, ip := range r.net.Collection.IPs {
		ipList.Items(r.renderIP(ip)...)
	}

	return ipList.String()
}

func (r *Model) renderIP(ip *IPStats) []any {
	var items []any

	if ip.PortsChecked == ip.PortsTotal {
		items = append(items, color.GreenString(ip.IP.String()))
	} else {
		items = append(items, fmt.Sprintf("%s (%d/%d ports)", ip.IP.String(), ip.PortsChecked, ip.PortsTotal))
	}

	if len(ip.Ports) == 0 {
		return items
	}

	slices.SortFunc(ip.Ports, func(a, b *PortStats) int {
		if a.Number < b.Number {
			return -1
		}

		return 1
	})

	portsList := list.New()

	for _, port := range ip.Ports {
		var descriptions []string
		for _, info := range r.tcpToPort[port.Number] {
			descriptions = append(descriptions, info.Description)
		}

		portsList.Item(fmt.Sprintf(":%d %s", port.Number, color.WhiteString(strings.Join(descriptions, "; "))))
	}

	items = append(items, portsList)

	return items
}
