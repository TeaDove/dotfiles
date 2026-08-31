package sdd

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"
)

const (
	maxReviews        = 5
	maxVerifyAttempts = 3
)

const permissionAcceptEdits = "acceptEdits"

var readOnlyBashTools = []string{
	toolRead, "Grep", "Glob", "LS", "TodoWrite",
	"Bash(git status:*)", "Bash(git diff:*)", "Bash(git log:*)", "Bash(git show:*)",
	"Bash(ls:*)", "Bash(cat:*)", "Bash(find:*)", "Bash(rg:*)", "Bash(grep:*)", "Bash(head:*)", "Bash(tail:*)",
}

var buildTestBashTools = []string{
	"Bash(go:*)", "Bash(gofmt:*)", "Bash(goimports:*)", "Bash(golangci-lint:*)",
	"Bash(make:*)", "Bash(task:*)", "Bash(just:*)",
	"Bash(npm:*)", "Bash(pnpm:*)", "Bash(yarn:*)", "Bash(node:*)", "Bash(npx:*)",
	"Bash(python:*)", "Bash(python3:*)", "Bash(pytest:*)", "Bash(uv:*)", "Bash(poetry:*)", "Bash(ruff:*)",
	"Bash(cargo:*)", "Bash(rustc:*)", "Bash(mvn:*)", "Bash(gradle:*)", "Bash(./gradlew:*)",
}

var editOnlyTools = []string{toolEdit, toolWrite, toolMultiEdit, toolNotebookEdit}

func verifyStageTools() []string {
	tools := make([]string, 0, len(readOnlyBashTools)+len(buildTestBashTools))
	tools = append(tools, readOnlyBashTools...)
	tools = append(tools, buildTestBashTools...)

	return tools
}

func editStageTools() []string {
	tools := make([]string, 0, len(readOnlyBashTools)+len(buildTestBashTools)+len(editOnlyTools))
	tools = append(tools, readOnlyBashTools...)
	tools = append(tools, buildTestBashTools...)
	tools = append(tools, editOnlyTools...)

	return tools
}

type orchestratorConfig struct {
	specPath    string
	specContent string
	preexisting string
	vcsKnown    bool
	session     string
}

type Orchestrator struct {
	runner   Runner
	reporter Reporter
	config   orchestratorConfig
}

func NewOrchestrator(runner Runner, reporter Reporter, config orchestratorConfig) *Orchestrator {
	return &Orchestrator{
		runner:   runner,
		reporter: reporter,
		config:   config,
	}
}

func (o *Orchestrator) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return ErrCancelled
	}

	blocked, err := o.gate(ctx)
	if err != nil {
		return err
	}

	if blocked {
		return ErrBlocked
	}

	return o.reviewLoop(ctx)
}

func (o *Orchestrator) gate(ctx context.Context) (bool, error) {
	plan, err := o.validateAndPlan(ctx)
	if err != nil {
		return false, err
	}

	if plan == nil {
		return true, nil
	}

	o.reporter.Stage("implement")

	res, err := o.runner.Run(ctx, Request{
		Label:          "implement",
		Prompt:         implementPrompt(o.config.specPath, o.config.specContent, plan.Plan),
		SessionID:      o.config.session,
		PermissionMode: permissionAcceptEdits,
		AllowedTools:   editStageTools(),
		Schema:         implementSchema,
	})
	if err != nil {
		return false, errors.Wrap(err, "implement")
	}

	impl, err := decodeStructured[ImplementResult](res)
	if err != nil {
		return false, err
	}

	if !impl.Status.IsValid() {
		return false, errors.Wrap(ErrInvalidResult, "unknown implement status")
	}

	if impl.Status == StatusBLOCKED {
		o.reportBlocked(impl.Questions, impl.Issues)

		return true, nil
	}

	o.reporter.Success(nonEmpty(impl.Summary, "implementation complete"))

	return false, nil
}

func (o *Orchestrator) validateAndPlan(ctx context.Context) (*PlanResult, error) {
	o.reporter.Stage("validate")

	res, err := o.runner.Run(ctx, Request{
		Label:           "validate",
		Prompt:          validatePrompt(o.config.specPath, o.config.specContent),
		AllowedTools:    readOnlyBashTools,
		DisallowedTools: editOnlyTools,
		Schema:          validateSchema,
	})
	if err != nil {
		return nil, errors.Wrap(err, "validate")
	}

	validation, err := decodeStructured[ValidationResult](res)
	if err != nil {
		return nil, err
	}

	if !validation.Status.IsValid() {
		return nil, errors.Wrap(ErrInvalidResult, "unknown validate status")
	}

	if validation.Status == StatusBLOCKED {
		o.reporter.Error("specification is incomplete")
		o.reportBlocked(nil, validation.Issues)

		return nil, nil
	}

	o.reporter.Success("specification is sufficient")

	return o.plan(ctx)
}

func (o *Orchestrator) plan(ctx context.Context) (*PlanResult, error) {
	o.reporter.Stage("plan")

	res, err := o.runner.Run(ctx, Request{
		Label:           "plan",
		Prompt:          planPrompt(o.config.specPath, o.config.specContent),
		AllowedTools:    readOnlyBashTools,
		DisallowedTools: editOnlyTools,
		Schema:          planSchema,
	})
	if err != nil {
		return nil, errors.Wrap(err, "plan")
	}

	plan, err := decodeStructured[PlanResult](res)
	if err != nil {
		return nil, err
	}

	if !plan.Status.IsValid() {
		return nil, errors.Wrap(ErrInvalidResult, "unknown plan status")
	}

	if plan.Status == StatusBLOCKED {
		o.reportBlocked(nil, plan.Issues)

		return nil, nil
	}

	o.reporter.Success(itoa(plan.RequirementCount) + " requirements covered by plan")

	return &plan, nil
}

func (o *Orchestrator) reviewLoop(ctx context.Context) error {
	var lastSignature string

	for iteration := 1; iteration <= maxReviews; iteration++ {
		err := o.verifyAndFix(ctx)
		if err != nil {
			return err
		}

		review, err := o.review(ctx, iteration)
		if err != nil {
			return err
		}

		switch review.verdict() {
		case verdictApproved:
			o.reporter.Success("APPROVED")

			return nil
		case verdictBlocked:
			o.reportReviewGaps(review.specGaps())

			return ErrBlocked
		case verdictRejected:
			signature := review.signature()
			if signature == lastSignature {
				o.reporter.Error("the same findings returned after a fix")

				return ErrNoProgress
			}

			lastSignature = signature

			if iteration == maxReviews {
				o.reportReviewLimit(review.actionableIssues())

				return ErrReviewLimit
			}

			err = o.fix(ctx, reviewFindings(review.actionableIssues()))
			if err != nil {
				return err
			}
		default:
			return ErrInvalidResult
		}
	}

	return ErrReviewLimit
}

func (o *Orchestrator) verifyAndFix(ctx context.Context) error {
	for attempt := 1; attempt <= maxVerifyAttempts; attempt++ {
		o.reporter.Stage("verify")

		res, err := o.runner.Run(ctx, Request{
			Label:           "verify",
			Prompt:          verifyPrompt(o.config.specPath, o.config.specContent),
			PermissionMode:  permissionAcceptEdits,
			AllowedTools:    verifyStageTools(),
			DisallowedTools: editOnlyTools,
			Schema:          verifySchema,
		})
		if err != nil {
			return errors.Wrap(err, "verify")
		}

		verify, err := decodeStructured[VerifyResult](res)
		if err != nil {
			return err
		}

		if !verify.Status.IsValid() {
			return errors.Wrap(ErrInvalidResult, "unknown verify status")
		}

		o.reportChecks(verify.Checks)

		switch verify.Status {
		case StatusPASSED:
			o.reporter.Success("verification passed")

			return nil
		case StatusBLOCKED:
			o.reporter.Error("verification blocked: " + verify.BlockedReason)

			return ErrBlocked
		case StatusFAILED:
			o.reporter.Error("verification failed")

			err = o.fix(ctx, verifyFindings(verify.Issues))
			if err != nil {
				return err
			}
		case StatusREADY, StatusDONE, StatusAPPROVED, StatusREJECTED:
			return errors.Wrap(ErrInvalidResult, "unexpected verify status")
		}
	}

	return ErrVerification
}

func (o *Orchestrator) fix(ctx context.Context, findings string) error {
	o.reporter.Stage("fix")

	res, err := o.runner.Run(ctx, Request{
		Label:          "fix",
		Prompt:         fixPrompt(o.config.specPath, o.config.specContent, findings),
		SessionID:      o.config.session,
		Resume:         true,
		PermissionMode: permissionAcceptEdits,
		AllowedTools:   editStageTools(),
		Schema:         implementSchema,
	})
	if err != nil {
		return errors.Wrap(err, "fix")
	}

	fix, err := decodeStructured[ImplementResult](res)
	if err != nil {
		return err
	}

	if !fix.Status.IsValid() {
		return errors.Wrap(ErrInvalidResult, "unknown fix status")
	}

	if fix.Status == StatusBLOCKED {
		o.reportBlocked(fix.Questions, fix.Issues)

		return ErrBlocked
	}

	o.reporter.Success(nonEmpty(fix.Summary, "fixes applied"))

	return nil
}

func (o *Orchestrator) review(ctx context.Context, iteration int) (ReviewResult, error) {
	o.reporter.Stage("review " + itoa(iteration) + "/" + itoa(maxReviews))
	o.reporter.Info("starting independent reviewer...")

	res, err := o.runner.Run(ctx, Request{
		Label:           "review",
		Prompt:          reviewPrompt(o.config.specPath, o.config.specContent, o.config.preexisting, o.config.vcsKnown),
		AllowedTools:    readOnlyBashTools,
		DisallowedTools: editOnlyTools,
		Schema:          reviewSchema,
	})
	if err != nil {
		return ReviewResult{}, errors.Wrap(err, "review")
	}

	review, err := decodeStructured[ReviewResult](res)
	if err != nil {
		return ReviewResult{}, err
	}

	if !review.Status.IsValid() {
		return ReviewResult{}, errors.Wrap(ErrInvalidResult, "unknown review status")
	}

	o.reportReviewIssues(review.Issues)

	return review, nil
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}
