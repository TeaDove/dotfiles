package watch

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"regexp"
)

type keymap struct {
	quit key.Binding
}

type model struct {
	commands []string

	grep      *regexp.Regexp
	grepInput textinput.Model
	spinner   spinner.Model
	help      help.Model
	keymap    keymap
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{}),
	)
}

func (r *model) View() string {
	return r.grepInput.View() + "\n" + lipgloss.JoinVertical(lipgloss.Left, r.commands...) + r.helpView()
}

func (r *model) Update(msgI tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msgI.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return r, tea.Quit
		}
	}

	r.spinner, cmd = r.spinner.Update(msgI)

	return r, cmd
}

func (r *model) Init() tea.Cmd {
	return tea.Batch(r.spinner.Tick, textinput.Blink)
}
