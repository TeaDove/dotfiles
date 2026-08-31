package sdd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopReporter struct{}

func (nopReporter) Stage(string)        {}
func (nopReporter) Info(string)         {}
func (nopReporter) Tool(string, string) {}
func (nopReporter) Success(string)      {}
func (nopReporter) Warning(string)      {}
func (nopReporter) Error(string)        {}
func (nopReporter) Verbose(string)      {}

type fakeResponse struct {
	result Result
	err    error
}

type fakeRunner struct {
	responses []fakeResponse
	labels    []string
	index     int
}

func (f *fakeRunner) Run(_ context.Context, req Request) (Result, error) {
	f.labels = append(f.labels, req.Label)

	if f.index >= len(f.responses) {
		return Result{}, errors.Newf("unexpected call %d: %s", f.index, req.Label)
	}

	response := f.responses[f.index]
	f.index++

	return response.result, response.err
}

func structured(value any) fakeResponse {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	return fakeResponse{result: Result{StructuredOutput: raw}}
}

func newOrchestrator(runner Runner) *Orchestrator {
	return NewOrchestrator(runner, nopReporter{}, orchestratorConfig{
		specPath:    "/spec.md",
		specContent: "spec body",
		session:     "session-1",
	})
}

func TestRunHappyPath(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 3}),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusPASSED}),
		structured(ReviewResult{Status: StatusAPPROVED}),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"validate", "plan", "implement", "verify", "review"}, runner.labels)
}

func TestRunSpecBlocked(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusBLOCKED, Issues: []Issue{{Description: "missing timeout"}}}),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrBlocked)
	assert.Equal(t, []string{"validate"}, runner.labels)
}

func TestRunPlanBlocked(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusBLOCKED, Issues: []Issue{{Description: "gap"}}}),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrBlocked)
	assert.Equal(t, []string{"validate", "plan"}, runner.labels)
}

func TestRunReviewRejectedThenApproved(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 1}),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusPASSED}),
		structured(ReviewResult{Status: StatusREJECTED, Issues: []ReviewIssue{
			{Severity: SeverityMAJOR, Type: FindingTypeIMPLEMENTATIONBUG, Description: "quota not released"},
		}}),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusPASSED}),
		structured(ReviewResult{Status: StatusAPPROVED}),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{"validate", "plan", "implement", "verify", "review", "fix", "verify", "review"},
		runner.labels,
	)
}

func TestRunReviewSpecGap(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 1}),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusPASSED}),
		structured(ReviewResult{Status: StatusREJECTED, Issues: []ReviewIssue{
			{Severity: SeverityMAJOR, Type: FindingTypeSPECGAP, Description: "cross-tenant behavior undefined"},
		}}),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrBlocked)
	assert.Equal(t, []string{"validate", "plan", "implement", "verify", "review"}, runner.labels)
}

func TestRunReviewLimit(t *testing.T) {
	t.Parallel()

	responses := []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 1}),
		structured(ImplementResult{Status: StatusDONE}),
	}

	for iteration := 1; iteration <= maxReviews; iteration++ {
		responses = append(responses, structured(VerifyResult{Status: StatusPASSED}))
		responses = append(responses, structured(ReviewResult{Status: StatusREJECTED, Issues: []ReviewIssue{
			{Severity: SeverityMAJOR, Type: FindingTypeIMPLEMENTATIONBUG, Description: "issue " + itoa(iteration)},
		}}))

		if iteration < maxReviews {
			responses = append(responses, structured(ImplementResult{Status: StatusDONE}))
		}
	}

	runner := &fakeRunner{responses: responses}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrReviewLimit)

	reviews := 0

	for _, label := range runner.labels {
		if label == "review" {
			reviews++
		}
	}

	assert.Equal(t, maxReviews, reviews)
}

func TestRunNoProgress(t *testing.T) {
	t.Parallel()

	sameIssue := ReviewResult{Status: StatusREJECTED, Issues: []ReviewIssue{
		{Severity: SeverityMAJOR, Type: FindingTypeIMPLEMENTATIONBUG, File: "a.go", Description: "same"},
	}}

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 1}),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusPASSED}),
		structured(sameIssue),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusPASSED}),
		structured(sameIssue),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrNoProgress)
}

func TestRunVerificationFailureThenSuccess(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 1}),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusFAILED, Issues: []Issue{{Description: "test failed"}}}),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusPASSED}),
		structured(ReviewResult{Status: StatusAPPROVED}),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{"validate", "plan", "implement", "verify", "fix", "verify", "review"}, runner.labels)
}

func TestRunVerificationBlocked(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 1}),
		structured(ImplementResult{Status: StatusDONE}),
		structured(VerifyResult{Status: StatusBLOCKED, BlockedReason: "docker unavailable"}),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrBlocked)
	assert.Equal(t, []string{"validate", "plan", "implement", "verify"}, runner.labels)
}

func TestRunImplementBlocked(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 1}),
		structured(ImplementResult{Status: StatusBLOCKED, Questions: []string{"which zone?"}}),
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrBlocked)
	assert.Equal(t, []string{"validate", "plan", "implement"}, runner.labels)
}

func TestRunClaudeExecutionFailure(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		structured(ValidationResult{Status: StatusREADY}),
		structured(PlanResult{Status: StatusREADY, RequirementCount: 1}),
		{err: errors.Wrap(ErrExecution, "boom")},
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrExecution)
	assert.Equal(t, []string{"validate", "plan", "implement"}, runner.labels)
}

func TestRunCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	runner := &fakeRunner{}

	err := newOrchestrator(runner).Run(ctx)
	require.ErrorIs(t, err, ErrCancelled)
	assert.Empty(t, runner.labels)
}

func TestRunMalformedOutput(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{responses: []fakeResponse{
		{result: Result{StructuredOutput: json.RawMessage("{not json")}},
	}}

	err := newOrchestrator(runner).Run(t.Context())
	require.ErrorIs(t, err, ErrInvalidResult)
	assert.Equal(t, []string{"validate"}, runner.labels)
}
