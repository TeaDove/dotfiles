package logs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatDefault(t *testing.T) {
	t.Parallel()

	formatter, err := NewLogFormatter(false)
	require.NoError(t, err)

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
			name:     "single leading tag preserved",
			input:    "2026.08.18 00:08:47.000000 [scenario/replay.go:313] I: [t=2026-08-01_18-00-50] Replay job: total=11977\n",
			expected: "00:08:47 [scenario/replay.go:313] I: [t=2026-08-01_18-00-50] Replay job: total=11977\n",
		},
		{
			name: "multiple leading tags preserved with spacing",
			input: "2026.08.18 00:08:47.000000 [scenario/replay.go:443] I: [al=1b453c4d] " +
				"[t=2026-08-01_06-00-54] Unexpected NER: req=allocation_spec:{cores:56}\n",
			expected: "00:08:47 [scenario/replay.go:443] I: [al=1b453c4d] " +
				"[t=2026-08-01_06-00-54] Unexpected NER: req=allocation_spec:{cores:56}\n",
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
			t.Parallel()
			assert.Equal(t, testCase.expected, formatter.format(testCase.input))
		})
	}
}

func TestFormatVerbose(t *testing.T) {
	t.Parallel()

	formatter, err := NewLogFormatter(true)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "leading tags stay on the header line",
			input: "2026.08.18 20:56:30.680860 [scenario/replay.go:313] I: [al=c586d418] " +
				"[t=2026-08-01] Unexpected NER: req=x\n",
			expected: "2026.08.18 20:56:30.68 [scenario/replay.go:313] I: [al=c586d418] " +
				"[t=2026-08-01]\nUnexpected NER: req=x\n",
		},
		{
			name:     "no tags keeps message on the new line",
			input:    "2026.08.18 20:56:35.379705 [simulation/state.go:262] E: no tags, boom\n",
			expected: "2026.08.18 20:56:35.37 [simulation/state.go:262] E:\nno tags, boom\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, testCase.expected, formatter.format(testCase.input))
		})
	}
}
