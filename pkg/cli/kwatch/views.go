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
	"sync"
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

	spinner  spinner.Model
	help     help.Model
	keymap   keymap
	drawLock sync.Mutex
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.quit}),
	)
}

func (r *model) View() string {
	r.drawLock.Lock()
	defer r.drawLock.Unlock()

	tables := make([]string, 0)

	if !r.deploymentsTableData.IsEmpty() {
		tables = append(tables, r.deploymentsTable.Render())
	}

	if !r.statefulsetTableData.IsEmpty() {
		tables = append(tables, r.statefulsetTable.Render())
	}

	if !r.cronjobTableData.IsEmpty() {
		tables = append(tables, r.cronjobTable.Render())
	}

	if !r.containersTableData.IsEmpty() {
		tables = append(tables, r.containersTable.Render())
	}

	return lipgloss.JoinVertical(lipgloss.Left, tables...) + r.helpView()
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
