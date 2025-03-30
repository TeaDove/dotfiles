package net_stats

import (
	"context"
	"dotfiles/cmd/cli/gloss_utils"
	"dotfiles/cmd/http_supplier"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	//"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"sync"
)

const elipsis = "..."

type NetStats struct {
	httpSupplier *http_supplier.Supplier
	model        *model
}

func NewNetStats(httpSupplier *http_supplier.Supplier) *NetStats {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	tableData := gloss_utils.NewMappingData(pingCols, addressesToPing)

	t := table.New().
		Wrap(true).
		Headers(pingCols...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#df8e1d"))).
		Data(tableData)

	m := model{
		myIP:       elipsis,
		openPorts:  elipsis,
		interfaces: elipsis,
		pings:      *t,
		pingsData:  tableData,
		spinner:    s,
		help:       help.New(),
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
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
	wg.Add(1)
	go r.pingsView(ctx, &wg)

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

func prettyWarn(err error) string {
	return color.YellowString("warning: ") + color.WhiteString(err.Error())
}
