package watch

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keymap struct {
	quit key.Binding
}

type model struct {
	commands []commandExecution

	grepInput textinput.Model
	spinner   spinner.Model
	help      help.Model
	keymap    keymap
}

func (r *model) helpView() string {
	return fmt.Sprintf(
		"\n %s %s ",
		r.spinner.View(),
		r.help.ShortHelpView([]key.Binding{r.keymap.quit}),
	)
}

func (r *model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, r.greppedCommands()...) + r.helpView() + r.grepInput.View()
}

func (r *model) greppedCommands() []string {
	grepValue := strings.ToLower(r.grepInput.Value())

	var newCommands []string

	for _, command := range r.commands {
		var out strings.Builder

		out.WriteString(command.cmd)

		for _, line := range strings.Split(command.out, "\n") {
			if strings.Contains(strings.ToLower(line), grepValue) {
				out.WriteString("\n")
				out.WriteString(line)
			}
		}

		out.WriteString("\n")
		newCommands = append(newCommands, out.String())
	}

	return newCommands
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

	r.grepInput, cmd = r.grepInput.Update(msgI)
	if cmd == nil {
		r.spinner, cmd = r.spinner.Update(msgI)
	}

	return r, cmd
}

func (r *model) Init() tea.Cmd {
	return tea.Batch(r.spinner.Tick, textinput.Blink)
}
