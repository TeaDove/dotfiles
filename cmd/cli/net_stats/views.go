package net_stats

import (
	"dotfiles/cmd/cli/gloss_utils"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	//"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func buildView() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Top).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#df8e1d"))
}

var (
	openPortsStyle  = buildView()
	interfacesStyle = buildView()
	myIPStyle       = buildView().Width(60)
	pingStyle       = buildView().Width(60).UnsetBorderStyle()
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
	pingsData  *gloss_utils.MappingData
	pings      table.Table
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.verbose, r.keymap.quit}),
	)
}

func (r *model) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinVertical(
			lipgloss.Top,
			myIPStyle.Render(r.myIP),
			pingStyle.Render(r.pings.String()),
		),
		lipgloss.JoinVertical(
			lipgloss.Top,
			interfacesStyle.Render(r.interfaces),
			openPortsStyle.Render(r.openPorts),
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
