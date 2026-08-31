package sdd

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
)

//go:generate go tool go-enum --values

// Status ENUM(READY, BLOCKED, DONE, APPROVED, REJECTED, PASSED, FAILED).
type Status string

// Severity ENUM(BLOCKER, MAJOR, MINOR).
type Severity string

// FindingType ENUM(IMPLEMENTATION_BUG, TEST_GAP, SPEC_GAP, OUT_OF_SCOPE).
type FindingType string

var (
	ErrBlocked       = errors.New("blocked: specification requires clarification")
	ErrExecution     = errors.New("execution error: claude exited unexpectedly")
	ErrVerification  = errors.New("verification failed")
	ErrReviewLimit   = errors.New("implementation was not approved after 5 independent reviews")
	ErrCancelled     = errors.New("cancelled")
	ErrInvalidResult = errors.New("claude returned malformed structured output")
	ErrNoProgress    = errors.New("review is not making progress")
)

type Issue struct {
	Description string `json:"description"`
}

type ValidationResult struct {
	Status Status  `json:"status"`
	Issues []Issue `json:"issues"`
}

type PlanResult struct {
	Status           Status  `json:"status"`
	Issues           []Issue `json:"issues"`
	RequirementCount int     `json:"requirementCount"`
	Plan             string  `json:"plan"`
}

type ImplementResult struct {
	Status    Status   `json:"status"`
	Summary   string   `json:"summary"`
	Questions []string `json:"questions"`
	Issues    []Issue  `json:"issues"`
}

type VerifyCheck struct {
	Command string `json:"command"`
	Passed  bool   `json:"passed"`
	Detail  string `json:"detail"`
}

type VerifyResult struct {
	Status        Status        `json:"status"`
	Checks        []VerifyCheck `json:"checks"`
	Issues        []Issue       `json:"issues"`
	BlockedReason string        `json:"blockedReason"`
}

type ReviewIssue struct {
	Severity    Severity    `json:"severity"`
	Type        FindingType `json:"type"`
	Requirement string      `json:"requirement"`
	File        string      `json:"file"`
	Description string      `json:"description"`
}

type ReviewResult struct {
	Status Status        `json:"status"`
	Issues []ReviewIssue `json:"issues"`
}

type reviewVerdict int

const (
	verdictApproved reviewVerdict = iota
	verdictRejected
	verdictBlocked
)

func (i ReviewIssue) actionable() bool {
	return i.Severity == SeverityBLOCKER || i.Severity == SeverityMAJOR
}

func (r ReviewResult) verdict() reviewVerdict {
	var (
		hasSpecGap    bool
		hasActionable bool
	)

	for _, issue := range r.Issues {
		if issue.Type == FindingTypeSPECGAP {
			hasSpecGap = true
		}

		if issue.actionable() {
			hasActionable = true
		}
	}

	if r.Status == StatusBLOCKED || hasSpecGap {
		return verdictBlocked
	}

	if hasActionable {
		return verdictRejected
	}

	return verdictApproved
}

func (r ReviewResult) actionableIssues() []ReviewIssue {
	actionable := make([]ReviewIssue, 0, len(r.Issues))

	for _, issue := range r.Issues {
		if issue.actionable() {
			actionable = append(actionable, issue)
		}
	}

	return actionable
}

func (r ReviewResult) specGaps() []ReviewIssue {
	gaps := make([]ReviewIssue, 0, len(r.Issues))

	for _, issue := range r.Issues {
		if issue.Type == FindingTypeSPECGAP {
			gaps = append(gaps, issue)
		}
	}

	return gaps
}

func (r ReviewResult) signature() string {
	parts := make([]string, 0, len(r.Issues))

	for _, issue := range r.Issues {
		key := string(issue.Severity) + ":" + issue.File + ":" + issue.Description
		parts = append(parts, strings.ToLower(strings.TrimSpace(key)))
	}

	sort.Strings(parts)

	return strings.Join(parts, "|")
}

func decodeStructured[T any](res Result) (T, error) {
	var out T

	raw := res.StructuredOutput
	if len(raw) == 0 {
		raw = json.RawMessage(res.RawResult)
	}

	if len(strings.TrimSpace(string(raw))) == 0 {
		return out, errors.Wrap(ErrInvalidResult, "empty result")
	}

	err := json.Unmarshal(raw, &out)
	if err != nil {
		return out, errors.Wrap(ErrInvalidResult, "unmarshal result")
	}

	return out, nil
}
