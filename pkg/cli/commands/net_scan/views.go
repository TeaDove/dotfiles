package net_scan

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/teadove/netports"
)

type keymap struct {
	quit key.Binding
}

type Model struct {
	net     *NetSystem
	spinner spinner.Model
	help    help.Model
	keymap  keymap

	tcpToPort map[uint16]netports.Ports
}

func (r *Model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.quit}),
	)
}

func (r *Model) Update(msgI tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msgI.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "й":
			return r, tea.Quit
		}
	default:
		r.spinner, cmd = r.spinner.Update(msg)
	}

	return r, cmd
}

func (r *Model) Init() tea.Cmd {
	return r.spinner.Tick
}
