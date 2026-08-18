package logs

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestPrettifyLine(t *testing.T) {
	color.NoColor = true

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "info line",
			input:    "2026.08.18 20:56:30.680860 [scheduler-simulation/executor.go:227] I: Preparing job specs...\n",
			expected: "20:56:30 [scheduler-simulation/executor.go:227] I: Preparing job specs...\n",
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
			assert.Equal(t, testCase.expected, prettifyLine(testCase.input))
		})
	}
}
