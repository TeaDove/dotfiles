package logs

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestPrettifyLine(t *testing.T) {
	color.NoColor = true

	parser := NewLogParser()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "info line",
			input:    "2026.08.18 20:56:30.680860 [executor/executor.go:227] I: Preparing execution...\n",
			expected: "20:56:30 [executor/executor.go:227] I: Preparing execution...\n",
		},
		{
			name:     "error line",
			input:    "2026.08.18 20:56:31.000000 [x.go:1] E: boom\n",
			expected: "20:56:31 [x.go:1] E: boom\n",
		},
		{
			name:     "warn line keeps trailing json fields in msg",
			input:    "2026.08.18 20:56:31.000000 [x.go:1] W: done {\"k\": 1}\n",
			expected: "20:56:31 [x.go:1] W: done {\"k\": 1}\n",
		},
		{
			name:     "single leading tag preserved",
			input:    "2026.08.18 00:08:47.000000 [scenario/task.go:313] I: [t=123] Job: total=11977\n",
			expected: "00:08:47 [scenario/task.go:313] I: [t=123] Job: total=11977\n",
		},
		{
			name:     "multiple leading tags preserved with spacing",
			input:    "2026.08.18 00:08:47.000000 [scenario/task.go:443] I: [al=1b453c4d] [t=XYZ] Unexpected error: req=user:123\n",
			expected: "00:08:47 [scenario/task.go:443] I: [al=1b453c4d] [t=XYZ] Unexpected error: req=user:123\n",
		},
		{
			name:     "line without trailing newline",
			input:    "2026.08.18 20:56:31.000000 [x.go:1] I: tail",
			expected: "20:56:31 [x.go:1] I: tail",
		},
		{
			name:     "unparsed plain line returned as is",
			input:    "plain unparsed line\n",
			expected: "plain unparsed line\n",
		},
		{
			name:     "unknown level letter returned as is",
			input:    "2026.08.18 20:56:31.000000 [x.go:1] X: mystery\n",
			expected: "2026.08.18 20:56:31.000000 [x.go:1] X: mystery\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, parser.prettifyLine(testCase.input))
		})
	}
}

func TestPrettifyLineAppliesColor(t *testing.T) {
	color.NoColor = false

	parser := NewLogParser()

	out := parser.prettifyLine("2026.08.18 00:08:47.000000 [main.go:313] I: [t=2026-08-01] Replay service\n")

	assert.Contains(t, out, "\x1b[96m[main.go:313]\x1b[0m", "expected hi-cyan caller")
	assert.Contains(t, out, "\x1b[32mI:\x1b[0m", "expected green level")
	assert.Contains(t, out, "\x1b[90m[t=2026-08-01]\x1b[0m", "expected gray tag")
}
