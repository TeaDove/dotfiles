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
			input:    "2026.08.18 20:56:30.680860 [runner/executor.go:227] I: Preparing runner\n",
			expected: "20:56:30 [runner/executor.go:227] I: Preparing runner\n",
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
			input:    "2026.08.18 00:08:47.000000 [runner/replay.go:313] I: [t=2026-08-01] Replay runner: total=11977\n",
			expected: "00:08:47 [runner/replay.go:313] I: [t=2026-08-01] Replay runner: total=11977\n",
		},
		{
			name: "multiple leading tags preserved with spacing",
			input: "2026.08.18 00:08:47.000000 [runner/replay.go:443] I: [al=1b453c4d] " +
				"[t=2026-08-01_06-00-54] Unexpected error: req={user:56}\n",
			expected: "00:08:47 [runner/replay.go:443] I: [al=1b453c4d] " +
				"[t=2026-08-01_06-00-54] Unexpected error: req={user:56}\n",
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
			input: "2026.08.18 20:56:30.680860 [runner/replay.go:313] I: [tag=123abc] " +
				"[t=2026-08-01] Unexpected error: req=x\n",
			expected: "2026.08.18 20:56:30.68 [runner/replay.go:313] I: [tag=123abc] " +
				"[t=2026-08-01]\nUnexpected error: req=x\n",
		},
		{
			name:     "no tags keeps message on the new line",
			input:    "2026.08.18 20:56:35.379705 [runner/state.go:123] E: no tags, boom\n",
			expected: "2026.08.18 20:56:35.37 [runner/state.go:123] E:\nno tags, boom\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, testCase.expected, formatter.format(testCase.input))
		})
	}
}

func TestArgRegexp(t *testing.T) {
	t.Parallel()

	found := argRegexp.FindAllString("Job completed, elapsed=20m, total=200, deleted=1", -1)

	assert.Equal(t, []string{"elapsed=", "total=", "deleted="}, found)
}
