package net_system

import (
	"context"
	"dotfiles/pkg/cli/gloss_utils"
	"dotfiles/pkg/http_supplier"
	"sync"
	"sync/atomic"

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
	elipsis := "..."
	m.myIP.Store(&elipsis)
	m.openPorts.Store(&elipsis)
	m.interfaces.Store(&elipsis)

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

	wg.Go(func() { runAndSet(ctx, &r.model.myIP, r.myIPView) })
	wg.Go(func() { runAndSet(ctx, &r.model.interfaces, r.interfacesView) })
	wg.Go(func() { runAndSet(ctx, &r.model.openPorts, r.openPortsView) })
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

func runAndSet(ctx context.Context, ptr *atomic.Pointer[string], fn func(ctx context.Context) string) {
	v := fn(ctx)
	ptr.Store(&v)
}
