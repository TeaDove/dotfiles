package kwatch

import (
	"dotfiles/pkg/cli/gloss_utils"
	"fmt"
	"regexp"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type keymap struct {
	quit  key.Binding
	space key.Binding
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

	regexp      *regexp.Regexp
	regexpInput textinput.Model
	spinner     spinner.Model
	help        help.Model
	keymap      keymap
	drawLock    sync.Mutex
	altscreen   bool
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.space, r.keymap.quit}),
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

	return r.regexpInput.View() + "\n" + lipgloss.JoinVertical(lipgloss.Left, tables...) + r.helpView()
}

func (r *model) Update(msgI tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msgI.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return r, tea.Quit
		case " ":
			if r.altscreen {
				cmd = tea.ExitAltScreen
			} else {
				cmd = tea.EnterAltScreen
			}

			r.altscreen = !r.altscreen

			return r, cmd
		}
	}

	r.regexpInput, cmd = r.regexpInput.Update(msgI)
	if cmd == nil {
		r.spinner, cmd = r.spinner.Update(msgI)
	} else {
		var err error

		r.regexp, err = regexp.Compile(r.regexpInput.Value())
		if err != nil {
			r.regexpInput.Prompt = fmt.Errorf("bad regexp: %w > ", err).Error()
			r.regexp = nil
		} else {
			r.regexpInput.Prompt = "> "
		}
	}

	return r, cmd
}

func (r *model) Init() tea.Cmd {
	return tea.Batch(r.spinner.Tick, textinput.Blink)
}
