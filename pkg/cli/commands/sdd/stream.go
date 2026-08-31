package sdd

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
)

//go:generate go tool go-enum

// eventKind ENUM(unknown, init, assistantText, toolUse, toolResult, result).
type eventKind int

const (
	toolRead         = "Read"
	toolEdit         = "Edit"
	toolWrite        = "Write"
	toolMultiEdit    = "MultiEdit"
	toolNotebookEdit = "NotebookEdit"
	toolBash         = "Bash"
)

type PermissionDenial struct {
	Tool string
	Info string
}

type Event struct {
	Kind      eventKind
	SessionID string

	Text       string
	ToolName   string
	ToolDetail string
	ToolResult string

	IsError           bool
	RawResult         string
	StructuredOutput  json.RawMessage
	PermissionDenials []PermissionDenial

	Model string
	Tools []string

	Raw string
}

func parseEvents(line string) []Event {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	if !gjson.Valid(line) {
		return []Event{{Kind: EventKindUnknown, Raw: line}}
	}

	root := gjson.Parse(line)
	sessionID := root.Get("session_id").String()

	switch root.Get("type").String() {
	case "system":
		return parseSystem(root, sessionID)
	case "assistant":
		return parseAssistant(root, sessionID)
	case "user":
		return parseUser(root, sessionID)
	case "result":
		return []Event{parseResult(root, sessionID)}
	default:
		return nil
	}
}

func parseSystem(root gjson.Result, sessionID string) []Event {
	if root.Get("subtype").String() != "init" {
		return nil
	}

	event := Event{
		Kind:      EventKindInit,
		SessionID: sessionID,
		Model:     root.Get("model").String(),
	}

	root.Get("tools").ForEach(func(_, value gjson.Result) bool {
		event.Tools = append(event.Tools, value.String())

		return true
	})

	return []Event{event}
}

func parseAssistant(root gjson.Result, sessionID string) []Event {
	var events []Event

	root.Get("message.content").ForEach(func(_, item gjson.Result) bool {
		switch item.Get("type").String() {
		case "text":
			text := strings.TrimSpace(item.Get("text").String())
			if text != "" {
				events = append(events, Event{Kind: EventKindAssistantText, SessionID: sessionID, Text: text})
			}
		case "tool_use":
			events = append(events, Event{
				Kind:       EventKindToolUse,
				SessionID:  sessionID,
				ToolName:   item.Get("name").String(),
				ToolDetail: toolDetail(item.Get("name").String(), item.Get("input")),
			})
		default:
		}

		return true
	})

	return events
}

func parseUser(root gjson.Result, sessionID string) []Event {
	var events []Event

	root.Get("message.content").ForEach(func(_, item gjson.Result) bool {
		if item.Get("type").String() != "tool_result" {
			return true
		}

		content := item.Get("content")

		detail := content.String()
		if content.IsArray() {
			detail = content.Get("0.text").String()
		}

		events = append(
			events,
			Event{Kind: EventKindToolResult, SessionID: sessionID, ToolResult: truncate(detail, 240)},
		)

		return true
	})

	return events
}

func parseResult(root gjson.Result, sessionID string) Event {
	event := Event{
		Kind:      EventKindResult,
		SessionID: sessionID,
		IsError:   root.Get("is_error").Bool(),
		RawResult: root.Get("result").String(),
	}

	structured := root.Get("structured_output")
	if structured.Exists() {
		event.StructuredOutput = json.RawMessage(structured.Raw)
	}

	root.Get("permission_denials").ForEach(func(_, denial gjson.Result) bool {
		event.PermissionDenials = append(event.PermissionDenials, PermissionDenial{
			Tool: denial.Get("tool_name").String(),
			Info: denial.Get("tool_input").String(),
		})

		return true
	})

	return event
}

func toolDetail(name string, input gjson.Result) string {
	switch name {
	case toolRead, toolEdit, toolWrite, toolMultiEdit, toolNotebookEdit:
		return input.Get("file_path").String()
	case toolBash:
		return input.Get("command").String()
	case "Grep", "Glob":
		return input.Get("pattern").String()
	default:
		return truncate(input.Raw, 120)
	}
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}

	return value[:limit] + "…"
}
