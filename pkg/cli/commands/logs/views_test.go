package logs

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func typeString(t *testing.T, m *model, s string) {
	t.Helper()

	for _, r := range s {
		m.Update(keyRune(r))
	}
}

func TestModel_SlashThenTypeThenEnter_AppliesIncludeFilter(t *testing.T) {
	t.Parallel()

	m := newModel()

	m.Update(keyRune('/'))
	assert.Equal(t, editInclude, m.editing)

	typeString(t, m, "ERROR")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, editNone, m.editing)
	assert.Equal(t, "ERROR", m.filter.IncludePattern())
	assert.Equal(t, "", m.lastErr)
}

func TestModel_BangThenTypeThenEnter_AppliesExcludeFilter(t *testing.T) {
	t.Parallel()

	m := newModel()

	m.Update(keyRune('!'))
	assert.Equal(t, editExclude, m.editing)

	typeString(t, m, "DEBUG")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, editNone, m.editing)
	assert.Equal(t, "DEBUG", m.filter.ExcludePattern())
}

func TestModel_Esc_DiscardsEditWithoutApplying(t *testing.T) {
	t.Parallel()

	m := newModel()

	m.Update(keyRune('/'))
	typeString(t, m, "ERROR")
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.Equal(t, editNone, m.editing)
	assert.Equal(t, "", m.filter.IncludePattern())
}

func TestModel_InvalidRegex_KeepsPreviousPatternAndSetsError(t *testing.T) {
	t.Parallel()

	m := newModel()

	m.Update(keyRune('/'))
	typeString(t, m, "ERROR")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	m.Update(keyRune('/'))
	typeString(t, m, "(unclosed")
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, "ERROR", m.filter.IncludePattern())
	assert.NotEqual(t, "", m.lastErr)
}

func TestModel_LogLineMsg_MatchingLineProducesPrintCmd(t *testing.T) {
	t.Parallel()

	m := newModel()

	require.NoError(t, m.filter.SetInclude("ERROR"))

	_, cmd := m.Update(logLineMsg("this is an ERROR"))
	assert.NotNil(t, cmd)
}

func TestModel_LogLineMsg_NonMatchingLineProducesNoCmd(t *testing.T) {
	t.Parallel()

	m := newModel()

	require.NoError(t, m.filter.SetInclude("ERROR"))

	_, cmd := m.Update(logLineMsg("all good here"))
	assert.Nil(t, cmd)
}

func TestModel_QuitKeys(t *testing.T) {
	t.Parallel()

	for _, key := range []tea.KeyMsg{keyRune('q'), {Type: tea.KeyCtrlC}} {
		m := newModel()

		_, cmd := m.Update(key)
		require.NotNil(t, cmd)

		msg := cmd()
		assert.IsType(t, tea.QuitMsg{}, msg)
	}
}

func TestModel_StreamClosedMsg_ProducesCmd(t *testing.T) {
	t.Parallel()

	m := newModel()

	_, cmd := m.Update(streamClosedMsg{path: "/tmp/ulogs/x.logs", lineCount: 3})
	assert.NotNil(t, cmd)
}

func TestModel_ChangingFilter_RedrawsBufferedHistory(t *testing.T) {
	t.Parallel()

	m := newModel()

	m.Update(logLineMsg("0 ERROR"))
	m.Update(logLineMsg("0 OK"))
	m.Update(logLineMsg("1 ERROR"))
	m.Update(logLineMsg("1 OK"))

	require.NoError(t, m.filter.SetInclude("ERROR"))

	m.Update(keyRune('/'))
	m.input.SetValue("")
	typeString(t, m, "OK")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	require.NotNil(t, cmd, "changing the filter must trigger a redraw of buffered history")
	assert.Equal(t, "OK", m.filter.IncludePattern())

	matched := filterLines(m.lines, &m.filter)
	assert.Equal(t, []string{"0 OK", "1 OK"}, matched)
}

func TestModel_ReapplyingSameFilter_DoesNotRedraw(t *testing.T) {
	t.Parallel()

	m := newModel()

	m.Update(logLineMsg("0 ERROR"))
	require.NoError(t, m.filter.SetInclude("ERROR"))

	m.Update(keyRune('/'))
	m.input.SetValue("")
	typeString(t, m, "ERROR")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Nil(t, cmd)
}

func TestFilterLines(t *testing.T) {
	t.Parallel()

	var f Filter
	require.NoError(t, f.SetInclude("error"))

	lines := []string{"0 ERROR", "0 OK", "1 error", "1 ok"}
	assert.Equal(t, []string{"0 ERROR", "1 error"}, filterLines(lines, &f))
}
