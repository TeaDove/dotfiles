package net_traceroute

import (
	"context"
	"dotfiles/pkg/cli/gloss_utils"
	"dotfiles/pkg/http_supplier"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v3"
)

type Service struct {
	model *model

	hops   []traceResult
	hopsMu sync.RWMutex

	httpSupplier *http_supplier.Supplier
}

func Run(ctx context.Context, command *cli.Command) error {
	service := New()
	p := tea.NewProgram(service.model, tea.WithContext(ctx))

	var (
		wg        sync.WaitGroup
		runnerErr error
	)
	wg.Go(func() { runnerErr = service.run(ctx, command.Args().First()) })

	go func() {
		wg.Wait()
		p.Quit()
	}()

	_, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "run tea")
	}

	if runnerErr != nil {
		return errors.Wrap(runnerErr, "run error")
	}

	return nil
}

func New() *Service {
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
	}
	service.model.service = &service

	return &service
}
