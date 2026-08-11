package logs

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type editTarget int

const (
	editNone editTarget = iota
	editInclude
	editExclude
)

// maxBufferedLines caps in-memory history kept for filter-change redraws, so a
// long-running script doesn't grow this process's memory without bound.
const maxBufferedLines = 50_000

type logLineMsg string

type streamClosedMsg struct {
	path      string
	lineCount int
	err       error
}

type model struct {
	filter  Filter
	editing editTarget
	input   textinput.Model
	lastErr string
	lines   []string
}

func newModel() *model {
	input := textinput.New()
	input.Placeholder = ""
	input.CharLimit = 256
	input.Width = 40

	return &model{input: input}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) View() string {
	if m.editing != editNone {
		return m.editingView()
	}

	return m.footer()
}

func (m *model) Update(msgI tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msgI.(type) {
	case tea.KeyMsg:
		return m.updateKey(msg)
	case logLineMsg:
		line := string(msg)
		m.appendLine(line)

		if m.filter.Matches(line) {
			return m, tea.Println(line)
		}

		return m, nil
	case streamClosedMsg:
		return m, tea.Sequence(tea.Println(streamSummary(msg)), tea.Quit)
	default:
		return m, nil
	}
}

func (m *model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editing != editNone {
		return m.updateEditingKey(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "/":
		return m.startEditing(editInclude)
	case "!":
		return m.startEditing(editExclude)
	default:
		return m, nil
	}
}

func (m *model) updateEditingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = editNone

		return m, nil
	case "enter":
		return m.applyEditing()
	default:
		var cmd tea.Cmd

		m.input, cmd = m.input.Update(msg)

		return m, cmd
	}
}

func (m *model) startEditing(target editTarget) (tea.Model, tea.Cmd) {
	m.editing = target

	current := m.filter.IncludePattern()
	if target == editExclude {
		current = m.filter.ExcludePattern()
	}

	m.input.SetValue(current)
	m.input.CursorEnd()

	return m, tea.Batch(m.input.Focus(), textinput.Blink)
}

func (m *model) applyEditing() (tea.Model, tea.Cmd) {
	pattern := m.input.Value()
	target := m.editing
	m.editing = editNone

	var (
		err     error
		changed bool
	)

	if target == editInclude {
		before := m.filter.IncludePattern()
		err = m.filter.SetInclude(pattern)
		changed = err == nil && before != m.filter.IncludePattern()
	} else {
		before := m.filter.ExcludePattern()
		err = m.filter.SetExclude(pattern)
		changed = err == nil && before != m.filter.ExcludePattern()
	}

	if err != nil {
		m.lastErr = err.Error()

		return m, nil
	}

	m.lastErr = ""

	if !changed {
		return m, nil
	}

	return m, m.redrawCmd()
}

func (m *model) appendLine(line string) {
	m.lines = append(m.lines, line)

	if len(m.lines) > maxBufferedLines {
		m.lines = m.lines[len(m.lines)-maxBufferedLines:]
	}
}

// redrawCmd clears the terminal and reprints the full buffered history under
// the current filter, so switching filters retroactively reveals/hides
// already-seen lines instead of only affecting lines that arrive afterwards.
func (m *model) redrawCmd() tea.Cmd {
	matched := filterLines(m.lines, &m.filter)

	cmds := make([]tea.Cmd, 0, len(matched)+1)
	cmds = append(cmds, tea.ClearScreen)

	for _, line := range matched {
		cmds = append(cmds, tea.Println(line))
	}

	return tea.Sequence(cmds...)
}

func filterLines(lines []string, filter *Filter) []string {
	matched := make([]string, 0, len(lines))

	for _, line := range lines {
		if filter.Matches(line) {
			matched = append(matched, line)
		}
	}

	return matched
}

func streamSummary(msg streamClosedMsg) string {
	if msg.err != nil {
		return fmt.Sprintf("saved %d lines to %s (read error: %s)", msg.lineCount, msg.path, msg.err)
	}

	return fmt.Sprintf("saved %d lines to %s", msg.lineCount, msg.path)
}
