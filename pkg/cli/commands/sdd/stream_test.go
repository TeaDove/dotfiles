package sdd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEvents(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		line string
		want []Event
	}{
		{
			name: "init",
			line: `{"type":"system","subtype":"init","session_id":"s1","model":"claude-opus","tools":["Read","Bash"]}`,
			want: []Event{
				{Kind: EventKindInit, SessionID: "s1", Model: "claude-opus", Tools: []string{"Read", "Bash"}},
			},
		},
		{
			name: "thinking is skipped",
			line: `{"type":"assistant","session_id":"s1","message":{"content":[{"type":"thinking","thinking":"secret"}]}}`,
			want: nil,
		},
		{
			name: "assistant text",
			line: `{"type":"assistant","session_id":"s1","message":{"content":[{"type":"text","text":"hello"}]}}`,
			want: []Event{{Kind: EventKindAssistantText, SessionID: "s1", Text: "hello"}},
		},
		{
			name: "tool use read",
			line: `{"type":"assistant","session_id":"s1","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"a.go"}}]}}`,
			want: []Event{{Kind: EventKindToolUse, SessionID: "s1", ToolName: "Read", ToolDetail: "a.go"}},
		},
		{
			name: "tool result",
			line: `{"type":"user","session_id":"s1","message":{"content":[{"type":"tool_result","content":"ok"}]}}`,
			want: []Event{{Kind: EventKindToolResult, SessionID: "s1", ToolResult: "ok"}},
		},
		{
			name: "unknown system subtype ignored",
			line: `{"type":"system","subtype":"thinking_tokens","session_id":"s1"}`,
			want: nil,
		},
		{
			name: "unknown type ignored",
			line: `{"type":"rate_limit_event","session_id":"s1"}`,
			want: nil,
		},
		{
			name: "malformed line becomes unknown event",
			line: `{not json`,
			want: []Event{{Kind: EventKindUnknown, Raw: "{not json"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.want, parseEvents(testCase.line))
		})
	}
}

func TestParseEventsResult(t *testing.T) {
	t.Parallel()

	line := `{"type":"result","subtype":"success","is_error":false,"session_id":"s9",` +
		`"result":"{\"status\":\"READY\",\"issues\":[]}","structured_output":{"status":"READY","issues":[]},` +
		`"permission_denials":[{"tool_name":"Bash","tool_input":"rm -rf /"}]}`

	events := parseEvents(line)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, EventKindResult, event.Kind)
	assert.Equal(t, "s9", event.SessionID)
	assert.False(t, event.IsError)
	assert.JSONEq(t, `{"status":"READY","issues":[]}`, string(event.StructuredOutput))
	require.Len(t, event.PermissionDenials, 1)
	assert.Equal(t, "Bash", event.PermissionDenials[0].Tool)
}
