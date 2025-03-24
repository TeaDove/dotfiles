package net_stats

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func buildView() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Top).
		BorderStyle(lipgloss.RoundedBorder()).Width(80).
		BorderForeground(lipgloss.Color("69"))
}

var (
	myIPStyle       = buildView()
	interfacesStyle = buildView()
	openPortsStyle  = buildView()
	speedStyle      = buildView()
)

type keymap struct {
	verbose key.Binding
	quit    key.Binding
}

type model struct {
	spinner spinner.Model
	help    help.Model
	keymap  keymap

	myIP       string
	openPorts  string
	interfaces string
	pings      table.Model
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.verbose, r.keymap.quit}),
	)
}

func (r *model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			interfacesStyle.Render(r.interfaces),
			openPortsStyle.Render(r.openPorts),
		),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			myIPStyle.Render(r.myIP),
			speedStyle.Render(r.pings.View()),
		),
	) + r.helpView()
}

func (r *model) Update(msgI tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msgI.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return r, tea.Quit
		}
	default:
		r.spinner, cmd = r.spinner.Update(msg)
	}

	return r, cmd
}

func (r *model) Init() tea.Cmd {
	return r.spinner.Tick
}
