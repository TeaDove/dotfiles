package sdd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"
)

const (
	maxStreamLine = 64 * 1024 * 1024
)

type Request struct {
	Label           string
	Prompt          string
	SystemPrompt    string
	SessionID       string
	Resume          bool
	PermissionMode  string
	AllowedTools    []string
	DisallowedTools []string
	Schema          json.RawMessage
}

type Result struct {
	SessionID         string
	StructuredOutput  json.RawMessage
	RawResult         string
	IsError           bool
	PermissionDenials []PermissionDenial
}

type Runner interface {
	Run(ctx context.Context, req Request) (Result, error)
}

type claudeRunner struct {
	bin      string
	dir      string
	reporter Reporter
	verbose  bool
}

func newClaudeRunner(bin, dir string, reporter Reporter, verbose bool) *claudeRunner {
	return &claudeRunner{bin: bin, dir: dir, reporter: reporter, verbose: verbose}
}

func (r *claudeRunner) Run(ctx context.Context, req Request) (Result, error) {
	//nolint:gosec // launches the trusted claude binary resolved via LookPath; args are orchestrator-controlled
	cmd := exec.CommandContext(ctx, r.bin, r.buildArgs(req)...)
	cmd.Dir = r.dir
	cmd.Stdin = strings.NewReader(req.Prompt)

	var stderrBuf bytes.Buffer

	cmd.Stderr = &stderrBuf

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, errors.Wrap(err, "stdout pipe")
	}

	err = cmd.Start()
	if err != nil {
		return Result{}, errors.Wrap(err, "start claude")
	}

	result, gotResult := r.consume(stdout)

	waitErr := cmd.Wait()
	if waitErr != nil {
		if ctx.Err() != nil {
			return Result{}, errors.Wrap(ErrCancelled, "claude cancelled")
		}

		return Result{}, errors.Wrapf(
			ErrExecution,
			"claude exited: %v: %s",
			waitErr,
			strings.TrimSpace(stderrBuf.String()),
		)
	}

	if !gotResult {
		return Result{}, errors.Wrap(ErrExecution, "claude produced no result event")
	}

	if result.IsError {
		return Result{}, errors.Wrapf(ErrExecution, "claude reported error: %s", result.RawResult)
	}

	return result, nil
}

func (r *claudeRunner) buildArgs(req Request) []string {
	args := []string{"-p", "--output-format", "stream-json", "--verbose"}

	if req.Resume {
		args = append(args, "--resume", req.SessionID)
	} else if req.SessionID != "" {
		args = append(args, "--session-id", req.SessionID)
	}

	if req.PermissionMode != "" {
		args = append(args, "--permission-mode", req.PermissionMode)
	}

	if len(req.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(req.AllowedTools, " "))
	}

	if len(req.DisallowedTools) > 0 {
		args = append(args, "--disallowedTools", strings.Join(req.DisallowedTools, " "))
	}

	if len(req.Schema) > 0 {
		args = append(args, "--json-schema", string(req.Schema))
	}

	if req.SystemPrompt != "" {
		args = append(args, "--append-system-prompt", req.SystemPrompt)
	}

	return args
}

func (r *claudeRunner) consume(stdout io.Reader) (Result, bool) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), maxStreamLine)

	var (
		result    Result
		gotResult bool
	)

	for scanner.Scan() {
		for _, event := range parseEvents(scanner.Text()) {
			if event.Kind == EventKindResult {
				result = Result{
					SessionID:         event.SessionID,
					StructuredOutput:  event.StructuredOutput,
					RawResult:         event.RawResult,
					IsError:           event.IsError,
					PermissionDenials: event.PermissionDenials,
				}
				gotResult = true
			}

			r.dispatch(event)
		}
	}

	return result, gotResult
}

func (r *claudeRunner) dispatch(event Event) {
	switch event.Kind {
	case EventKindInit:
		r.reporter.Verbose("session " + event.SessionID + " (" + event.Model + ")")
	case EventKindAssistantText:
		r.reporter.Verbose("[claude] " + truncate(event.Text, 400))
	case EventKindToolUse:
		r.reporter.Tool(event.ToolName, event.ToolDetail)
	case EventKindToolResult:
		r.reporter.Verbose("[tool-result] " + event.ToolResult)
	case EventKindResult:
		for _, denial := range event.PermissionDenials {
			r.reporter.Warning("permission denied: " + denial.Tool + " " + denial.Info)
		}
	case EventKindUnknown:
		r.reporter.Verbose("[unknown] " + truncate(event.Raw, 200))
	default:
	}
}
