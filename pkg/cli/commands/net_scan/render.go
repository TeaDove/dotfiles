package net_scan

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/fatih/color"
)

func (r *model) renderInterface() string {
	return fmt.Sprintf(
		"Scanning TCP ports %s on %s (%d/%d ips)",
		color.CyanString(r.net.collection.Network),
		color.CyanString(r.net.collection.Interface),
		r.net.collection.IPsChecked,
		r.net.collection.IPsTotal,
	)
}

func (r *model) View() string {
	r.net.collectionMu.Lock()
	defer r.net.collectionMu.Unlock()

	return lipgloss.JoinVertical(lipgloss.Left, r.renderInterface(), r.renderIPList()) + r.helpView()
}

func (r *model) renderIPList() string {
	if len(r.net.collection.IPs) == 0 {
		return "no pingable addresses found"
	}

	slices.SortFunc(r.net.collection.IPs, func(a, b *IPStats) int {
		if a.IP.String() > b.IP.String() {
			return 1
		}

		return -1
	})

	ipList := list.New()
	for _, ip := range r.net.collection.IPs {
		ipList.Items(r.renderIP(ip)...)
	}

	return ipList.String()
}

func (r *model) renderIP(ip *IPStats) []any {
	items := []any{r.renderIPLine(ip)}

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
		for _, info := range r.net.wellKnownPorts[port.Number] {
			descriptions = append(descriptions, info.Description)
		}

		var item strings.Builder
		fmt.Fprintf(&item, ":%d", port.Number)

		if port.Message != "" {
			fmt.Fprintf(&item, " %s", color.CyanString(port.Message))
		}

		if len(descriptions) != 0 {
			fmt.Fprintf(&item, " [%s]", color.WhiteString(strings.Join(descriptions, "; ")))
		}

		portsList.Item(item.String())
	}

	items = append(items, portsList)

	return items
}

func (r *model) renderIPLine(ip *IPStats) string {
	var ipString string

	if ip.PortsChecked == ip.PortsTotal {
		ipString = color.GreenString(ip.IP.String())
	} else {
		ipString = fmt.Sprintf("%s (%d/%d ports)", ip.IP.String(), ip.PortsChecked, ip.PortsTotal)
	}

	if ip.Mac != "" {
		ipString += fmt.Sprintf(" (%s)", color.WhiteString(ip.Mac))
	}

	return ipString
}
