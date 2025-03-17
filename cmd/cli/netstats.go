package cli

import (
	"context"
	"dotfiles/cmd/http_supplier"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/teadove/teasutils/utils/strings_utils"
	"github.com/urfave/cli/v3"
	"strconv"
	"strings"
)

func (r *CLI) commandNet(ctx context.Context, command *cli.Command) error {
	text, err := r.pprintMyIP(ctx)
	if err != nil {
		color.Red("Failed to print my ip: %v\n", err)
	} else if text != "" {
		fmt.Println(text)
	}

	fmt.Println(r.pprintDNS())
	fmt.Println()

	text, err = r.pprintNetInterfaces(ctx)
	if err != nil {
		color.Red("Failed to print net interfaces: %v\n", err)
	} else if text != "" {
		fmt.Println(text)
		fmt.Println()
	}

	text, err = r.pprintOpenPorts(ctx)
	if err != nil {
		color.Red("Failed to print open ports: %v\n", err)
	} else if text != "" {
		fmt.Println(text)

	}

	return nil
}

func (r *CLI) pprintNetInterfaces(ctx context.Context) (string, error) {
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to get interfaces")
	}

	interfacesWithAddressess := make(net.InterfaceStatList, 0)
	for _, i := range interfaces {
		if len(i.Addrs) == 0 {
			continue
		}

		interfacesWithAddressess = append(interfacesWithAddressess, i)
	}

	if len(interfacesWithAddressess) == 0 {
		return "", errors.New("no interfaces found")
	}

	var text strings.Builder
	text.WriteString("Interfaces with addresses:")
	for _, i := range interfacesWithAddressess {
		addresses := make([]string, 0, len(i.Addrs))
		for _, a := range i.Addrs {
			addresses = append(addresses, a.Addr)
		}

		text.WriteString(
			fmt.Sprintf("\n%s (%s) -> %s",
				color.New(color.FgCyan, color.Faint).Sprintf(i.Name),
				i.HardwareAddr,
				strings.Join(addresses, ", "),
			),
		)
	}

	return text.String(), nil
}

func (r *CLI) pprintOpenPorts(ctx context.Context) (string, error) {
	connections, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		return "", errors.Wrap(err, "failed to get connections")
	}

	pidToPorts := make(map[int32][]string)
	for _, conn := range connections {
		if conn.Status != "LISTEN" || conn.Family != 2 || conn.Type != 1 {
			continue
		}
		_, ok := pidToPorts[conn.Pid]
		if ok {
			pidToPorts[conn.Pid] = append(pidToPorts[conn.Pid], strconv.Itoa(int(conn.Laddr.Port)))
		} else {
			pidToPorts[conn.Pid] = []string{strconv.Itoa(int(conn.Laddr.Port))}
		}
	}

	var text strings.Builder
	if len(pidToPorts) == 0 {
		return "", errors.New("no open ports")
	}

	text.WriteString("Open ports\n")

	for pid, ports := range pidToPorts {
		connProcess, err := process.NewProcess(pid)
		if err != nil {
			return "", errors.Wrap(err, "failed to get process")
		}

		name, err := connProcess.NameWithContext(ctx)
		if err != nil {
			return "", errors.Wrap(err, "failed to get process name")
		}

		text.WriteString(
			fmt.Sprintf("%s (%d): %s\n",
				color.New(color.FgBlue, color.Faint).Sprintf(name),
				pid,
				strings.Join(ports, ","),
			),
		)
	}

	return text.String(), nil
}

func (r *CLI) pprintMyIP(ctx context.Context) (string, error) {
	myIp, err := r.httpSupplier.MyIP(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to get MyIP")
	}

	return fmt.Sprintf("My IP: %s (%s)", color.YellowString(myIp), r.shortLocationOrErr(ctx, myIp)), nil
}

func (r *CLI) pprintDNS() string {
	dnss := http_supplier.GetDNSServers()
	dnssStrings := make([]string, 0, len(dnss))
	for _, dns := range dnss {
		dnssStrings = append(dnssStrings, dns.String())
	}

	return fmt.Sprintf("DNS Servers: %s", color.WhiteString(strings_utils.JoinStringers(dnss, ", ")))
}

func (r *CLI) shortLocationOrErr(ctx context.Context, ipOrDomain string) string {
	location, err := r.httpSupplier.LocateByIP(ctx, ipOrDomain)
	if err != nil {
		return fmt.Sprintf("failed to get location %v", err)
	}

	return fmt.Sprintf("%s, %s, %s", location.Country, location.City, color.BlueString(location.Org))
}
