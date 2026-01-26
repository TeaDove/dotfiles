package net_traceroute

import (
	"context"
	"dotfiles/pkg/cli/gloss_utils"
	"dotfiles/pkg/http_supplier"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
	"golang.org/x/net/icmp"
)

type Service struct {
	model *model

	hops   []traceResult
	hopsMu sync.Mutex

	httpSupplier *http_supplier.Supplier

	maxHops  int
	timeout  time.Duration
	basePort uint16
	dstIP    net.IP

	icmpConn *icmp.PacketConn
}

func Run(ctx context.Context, command *cli.Command) error {
	dstIP, err := parseAddr(ctx, command.Args().First())
	if err != nil {
		return errors.Wrap(err, "invalid target address")
	}

	service := New(dstIP)
	p := tea.NewProgram(service.model, tea.WithContext(ctx))

	var (
		wg        sync.WaitGroup
		runnerErr error
	)
	wg.Go(func() { runnerErr = service.traceRoute(ctx, dstIP) })

	go func() {
		wg.Wait()
		p.Quit()
	}()

	_, err = p.Run()
	if err != nil {
		return errors.Wrap(err, "run tea")
	}

	if runnerErr != nil {
		return errors.Wrap(runnerErr, "run error")
	}

	return nil
}

func New(dstIP net.IP) *Service {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Spinner.FPS = time.Second / 20

	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	tableData := gloss_utils.NewMappingData(tableCols...)

	t := table.New().
		Wrap(true).
		Headers(tableCols...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("69"))).
		Data(tableData)

	m := model{
		spinner: s,
		help:    help.New(),
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
		traceTableData: tableData,
		traceTable:     *t,
	}

	service := Service{
		hops:         make([]traceResult, 0, 100),
		httpSupplier: http_supplier.New(),
		model:        &m,
		maxHops:      40,
		timeout:      1 * time.Second,
		basePort:     33434,
		dstIP:        dstIP,
	}
	service.model.service = &service
	service.model.target = fmt.Sprintf(
		"traceroute to %s, %d hops max",
		color.CyanString(dstIP.String()),
		service.maxHops,
	)

	return &service
}
