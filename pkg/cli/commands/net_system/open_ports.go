package net_system

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

func (r *Service) openPortsView(ctx context.Context) string {
	connections, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		return prettyErr(err)
	}

	pidToPorts := make(map[int32][]string)

	for _, conn := range connections {
		if conn.Status != "LISTEN" || conn.Type != 1 {
			continue
		}

		_, ok := pidToPorts[conn.Pid]
		if ok {
			pidToPorts[conn.Pid] = append(pidToPorts[conn.Pid], strconv.Itoa(int(conn.Laddr.Port)))
		} else {
			pidToPorts[conn.Pid] = []string{strconv.Itoa(int(conn.Laddr.Port))}
		}
	}

	if len(pidToPorts) == 0 {
		return prettyWarn(errors.New("no open ports"))
	}

	v := color.GreenString("Open ports:")
	services := make([]string, 0, len(pidToPorts))

	for pid, ports := range pidToPorts {
		if pid == 0 {
			services = append(services, fmt.Sprintf("\nunknown: %s", strings.Join(ports, ",")))
			continue
		}

		connProcess, err := process.NewProcess(pid)
		if err != nil {
			return prettyErr(errors.Wrap(err, "get process name"))
		}

		name, err := connProcess.NameWithContext(ctx)
		if err != nil {
			return prettyErr(errors.Wrap(err, "get process name"))
		}

		services = append(services,
			fmt.Sprintf("\n%s (%d): %s",
				color.New(color.FgCyan, color.Faint).Sprint(name),
				pid,
				strings.Join(ports, ","),
			))
	}

	slices.Sort(services)

	for _, service := range services {
		v += service
	}

	return v
}
