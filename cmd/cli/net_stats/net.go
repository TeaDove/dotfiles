package net_stats

import (
	"context"
	"dotfiles/cmd/http_supplier"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/strings_utils"
	"sync"
)

type NetStats struct {
	httpSupplier *http_supplier.Supplier
	model        *model
}

func NewNetStats(httpSupplier *http_supplier.Supplier) *NetStats {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{
		myIP:       "...",
		openPorts:  "...",
		interfaces: "...",
		speed:      "...",
		spinner:    s,
		help:       help.New(),
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
			verbose: key.NewBinding(
				key.WithKeys("v"),
				key.WithHelp("v", "verbose"),
			),
		},
	}

	return &NetStats{httpSupplier: httpSupplier, model: &m}
}

func (r *NetStats) Net(ctx context.Context) error {
	p := tea.NewProgram(r.model, tea.WithContext(ctx))

	var wg sync.WaitGroup
	wg.Add(1)
	go r.myIPView(ctx, &wg)
	wg.Add(1)
	go r.interfacesView(ctx, &wg)
	wg.Add(1)
	go r.openPortsView(ctx, &wg)

	go func() {
		wg.Wait()
		p.Quit()
	}()

	_, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "failed to run tea")
	}

	return nil
}

func prettyErr(err error) string {
	return color.RedString("unexpected error: ") + color.WhiteString(err.Error())
}

func strOrPrettyErr(v string, err error) string {
	if err != nil {
		return prettyErr(err)
	}

	return v
}

//func (r *CLI) pprintPings(ctx context.Context) (string, error) {
//	pinger, err := ping.NewPinger("www.google.com")
//	if err != nil {
//		return "", errors.Wrap(err, "failed to init pinger")
//	}
//
//	pinger.Timeout = time.Second * 3
//
//	var res = ""
//	pinger.OnFinish = func(stats *ping.Statistics) {
//		res = fmt.Sprintf("%d packets transmitted, %d packets received, %v%% packet loss\n"+
//			"round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
//			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss,
//			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt,
//		)
//	}
//
//	err = pinger.Run()
//	if err != nil {
//		return "", errors.Wrap(err, "failed to ping www.google.com")
//	}
//
//	return res, nil
//}

func (r *NetStats) pprintMyIP(ctx context.Context) (string, error) {
	myIp, err := r.httpSupplier.MyIP(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to get MyIP")
	}

	return fmt.Sprintf(
		"%s: %s (%s)",
		color.GreenString("My IP"),
		color.YellowString(myIp.String()),
		r.shortLocationOrErr(ctx, myIp.String()),
	), nil
}

func (r *NetStats) pprintDNS(ctx context.Context) (string, error) {
	dnss := http_supplier.GetDNSServers()
	dnssStrings := make([]string, 0, len(dnss))

	for _, dns := range dnss {
		if dns.Addr().IsGlobalUnicast() {
			dnssStrings = append(
				dnssStrings,
				fmt.Sprintf("%s (%s)", dns.String(), r.shortLocationOrErr(ctx, dns.String())),
			)
		}

		dnssStrings = append(dnssStrings, dns.String())
	}

	return fmt.Sprintf(
		"%s: %s",
		color.GreenString("DNS Servers"),
		color.WhiteString(strings_utils.JoinStringers(dnss, ", ")),
	), nil
}

func (r *NetStats) shortLocationOrErr(ctx context.Context, ipOrDomain string) string {
	location, err := r.httpSupplier.LocateByIP(ctx, ipOrDomain)
	if err != nil {
		return prettyErr(errors.Wrap(err, "failed to get location"))
	}

	return fmt.Sprintf("%s, %s, %s", location.Country, location.City, color.BlueString(location.Org))
}
