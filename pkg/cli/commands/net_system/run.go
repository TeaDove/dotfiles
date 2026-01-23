package net_system

import (
	"context"
	"dotfiles/pkg/cli/gloss_utils"
	"dotfiles/pkg/http_supplier"
	"sync"

	"github.com/urfave/cli/v3"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"

	// "github.com/charmbracelet/bubbles/table".
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
)

const elipsis = "..."

func Run(ctx context.Context, _ *cli.Command) error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	tableData := gloss_utils.NewMappingData(pingCols...)

	t := table.New().
		Wrap(true).
		Headers(pingCols...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("69"))).
		Data(tableData)

	m := model{
		myIP:           elipsis,
		openPorts:      elipsis,
		interfaces:     elipsis,
		pingsTable:     *t,
		pingsTableData: tableData,
		spinner:        s,
		help:           help.New(),
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
	}

	r := Service{httpSupplier: http_supplier.New(), model: &m}

	return r.Run(ctx)
}

type Service struct {
	httpSupplier *http_supplier.Supplier
	model        *model
}

func (r *Service) Run(ctx context.Context) error {
	p := tea.NewProgram(r.model, tea.WithContext(ctx))

	var wg sync.WaitGroup

	wg.Go(func() { r.myIPView(ctx) })
	wg.Go(func() { r.interfacesView(ctx) })
	wg.Go(func() { r.openPortsView(ctx) })
	wg.Go(func() { r.pingsView(ctx) })

	go func() {
		wg.Wait()
		p.Quit()
	}()

	_, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "run tea")
	}

	return nil
}

func prettyErr(err error) string {
	return color.RedString("unexpected error: ") + color.WhiteString(err.Error())
}

func prettyWarn(err error) string {
	return color.YellowString("warning: ") + color.WhiteString(err.Error())
}
