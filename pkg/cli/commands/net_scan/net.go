package net_scan

import (
	"context"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

type NetSystem struct {
	Collection   Collection
	CollectionMu sync.RWMutex

	Model *Model
}

func Run(ctx context.Context, _ *cli.Command) error {
	r := &NetSystem{}

	s := spinner.New()
	s.Spinner = spinner.Dot
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
	p := tea.NewProgram(r.Model, tea.WithContext(ctx))

	var wg sync.WaitGroup
	wg.Go(func() { r.collect(ctx) })

	go func() {
		wg.Wait()
		p.Quit()
	}()

	_, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "run tea")
	}

	if r.Collection.Err != nil {
		return errors.New("collection error")
	}

	return nil
}
