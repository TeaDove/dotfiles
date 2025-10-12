package net_system

import (
	"dotfiles/pkg/cli/gloss_utils"
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func buildView() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Top).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69"))
}

var (
	openPortsStyle  = buildView()                              //nolint:gochecknoglobals // is ok
	interfacesStyle = buildView()                              //nolint:gochecknoglobals // is ok
	myIPStyle       = buildView().Width(60)                    //nolint:gochecknoglobals // is ok
	pingStyle       = buildView().Width(60).UnsetBorderStyle() //nolint:gochecknoglobals // is ok
)

type keymap struct {
	quit key.Binding
}

type model struct {
	spinner spinner.Model
	help    help.Model
	keymap  keymap

	myIP       string
	openPorts  string
	interfaces string

	pingsTable     table.Table
	pingsTableData *gloss_utils.MappingData
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.quit}),
	)
}

func (r *model) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.JoinVertical(
			lipgloss.Top,
			myIPStyle.Render(r.myIP),
			pingStyle.Render(r.pingsTable.String()),
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
