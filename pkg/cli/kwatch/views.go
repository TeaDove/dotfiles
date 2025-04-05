package kwatch

import (
	"dotfiles/pkg/cli/gloss_utils"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"strings"
)

func buildView() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Top)
	//BorderStyle(lipgloss.RoundedBorder()).
	//BorderForeground(lipgloss.Color("#df8e1d"))
}

var (
	containersStyle = buildView()
)

type keymap struct {
	quit key.Binding
}

type model struct {
	containersTable     *table.Table
	containersTableData *gloss_utils.MappingData

	deploymentsTable     *table.Table
	deploymentsTableData *gloss_utils.MappingData

	statefulsetTable     *table.Table
	statefulsetTableData *gloss_utils.MappingData

	cronjobTable     *table.Table
	cronjobTableData *gloss_utils.MappingData

	spinner spinner.Model
	help    help.Model
	keymap  keymap
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.quit}),
	)
}

func (r *model) View() string {
	var v strings.Builder
	if !r.deploymentsTableData.IsEmpty() {
		v.WriteString(r.deploymentsTable.String())
		v.WriteString("\n\n")
	}

	if !r.statefulsetTableData.IsEmpty() {
		v.WriteString(r.statefulsetTable.String())
		v.WriteString("\n\n")
	}

	if !r.cronjobTableData.IsEmpty() {
		v.WriteString(r.cronjobTable.String())
		v.WriteString("\n\n")
	}

	if !r.containersTableData.IsEmpty() {
		v.WriteString(r.containersTable.String())
		v.WriteString("\n\n")
	}

	v.WriteString(r.helpView())

	return v.String()
}

func (r *model) Update(msgI tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msgI.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
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
