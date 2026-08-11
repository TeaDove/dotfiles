package logs

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	hintStyle  = lipgloss.NewStyle().Faint(true)                       //nolint:gochecknoglobals // is ok
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) //nolint:gochecknoglobals // is ok
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) //nolint:gochecknoglobals // is ok
)

func (m *model) editingView() string {
	label := "include (/)"
	if m.editing == editExclude {
		label = "exclude (!)"
	}

	return labelStyle.Render(label+": ") + m.input.View()
}

func (m *model) footer() string {
	include := m.filter.IncludePattern()
	if include == "" {
		include = "-"
	}

	exclude := m.filter.ExcludePattern()
	if exclude == "" {
		exclude = "-"
	}

	line := fmt.Sprintf("[/ include: %s] [! exclude: %s] [q quit]", include, exclude)

	if m.lastErr != "" {
		line += " " + errStyle.Render("invalid regex: "+m.lastErr)
	}

	return hintStyle.Render(line)
}
