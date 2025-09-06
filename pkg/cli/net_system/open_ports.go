package net_system

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

func (r *NetStats) openPortsView(ctx context.Context) {
	connections, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		r.model.openPorts = prettyErr(err)
		return
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
		r.model.openPorts = prettyWarn(errors.New("no open ports"))
		return
	}

	r.model.openPorts = color.GreenString("Open ports:")
	services := make([]string, 0, len(pidToPorts))

	for pid, ports := range pidToPorts {
		if pid == 0 {
			services = append(services, fmt.Sprintf("\nunknown: %s", strings.Join(ports, ",")))
			continue
		}

		connProcess, err := process.NewProcess(pid)
		if err != nil {
			r.model.openPorts = prettyErr(errors.Wrap(err, "failed to get process name"))

			return
		}

		name, err := connProcess.NameWithContext(ctx)
		if err != nil {
			r.model.openPorts = prettyErr(errors.Wrap(err, "failed to get process name"))
			return
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
		r.model.openPorts += service
	}
}
