package net_scan

import (
	"context"
	"crypto/tls"
	netstd "net"
	"net/http"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mostlygeek/arp"
	"github.com/pkg/errors"
	"github.com/teadove/netports"
	"github.com/urfave/cli/v3"
)

type NetSystem struct {
	Collection   Collection
	CollectionMu sync.Mutex

	WellKnownPorts map[uint16]netports.Ports
	PortsToScan    []uint16
	ARPTable       arp.ArpTable

	client *http.Client
	dialer netstd.Dialer

	Model *Model
}

func New() *NetSystem {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint: gosec // As expected
	}

	r := &NetSystem{
		WellKnownPorts: netports.KnownPorts.FilterCollect(
			netports.FilterByProto(netports.TCP),
			netports.FilterByCategory(netports.CategoryWellKnown, netports.CategoryRegistered),
		).GroupByNumber(),
		ARPTable: arp.Table(),
		client:   &http.Client{Timeout: 3 * time.Second, Transport: tr},
		dialer:   netstd.Dialer{Timeout: 200 * time.Millisecond},
	}

	r.PortsToScan = getPortsToScan(r.WellKnownPorts)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Spinner.FPS = time.Second / 20

	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := Model{
		spinner: s,
		help:    help.New(),
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
		net: r,
	}
	r.Model = &m

	return r
}

func Run(ctx context.Context, cmd *cli.Command) error {
	r := New()
	p := tea.NewProgram(r.Model, tea.WithContext(ctx))

	var (
		wg         sync.WaitGroup
		collectErr error
	)
	wg.Go(func() { collectErr = r.collect(ctx, cmd.Args().Slice()) })

	go func() {
		wg.Wait()
		p.Quit()
	}()

	_, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "run tea")
	}

	if collectErr != nil {
		return errors.Wrap(collectErr, "collection error")
	}

	return nil
}

func getPortsToScan(wellKnown map[uint16]netports.Ports) []uint16 {
	const (
		firstPort = uint16(1)
		lastPort  = uint16(10_000)
	)

	ports := make([]uint16, 0, lastPort-firstPort)
	for port := range wellKnown {
		ports = append(ports, port)
	}

	for port := firstPort; port <= lastPort; port++ {
		_, ok := wellKnown[port]
		if ok {
			continue
		}

		ports = append(ports, port)
	}

	return ports
}
