package sdd

import (
	"strconv"
	"strings"
)

func itoa(value int) string {
	return strconv.Itoa(value)
}

func (o *Orchestrator) reportBlocked(questions []string, issues []Issue) {
	o.reporter.Error("BLOCKED")

	for _, question := range questions {
		o.reporter.Info("- " + question)
	}

	for _, issue := range issues {
		o.reporter.Info("- " + issue.Description)
	}

	o.reporter.Info("Update the specification and run the command again.")
}

func (o *Orchestrator) reportReviewGaps(gaps []ReviewIssue) {
	o.reporter.Error("BLOCKED: review found a specification gap")

	for _, gap := range gaps {
		o.reporter.Info("- " + gap.Description)
	}

	o.reporter.Info("Update the specification and run the command again.")
}

func (o *Orchestrator) reportReviewLimit(issues []ReviewIssue) {
	o.reporter.Error("implementation could not be approved after " + itoa(maxReviews) + " independent reviews")

	for _, issue := range issues {
		o.reporter.Info("- " + reviewIssueLine(issue))
	}
}

func (o *Orchestrator) reportChecks(checks []VerifyCheck) {
	for _, check := range checks {
		if check.Passed {
			o.reporter.Success(check.Command)

			continue
		}

		o.reporter.Error(check.Command + ": " + check.Detail)
	}
}

func (o *Orchestrator) reportReviewIssues(issues []ReviewIssue) {
	if len(issues) == 0 {
		return
	}

	for _, issue := range issues {
		if issue.Severity == SeverityMINOR {
			o.reporter.Warning(reviewIssueLine(issue))

			continue
		}

		o.reporter.Error(reviewIssueLine(issue))
	}
}

func reviewIssueLine(issue ReviewIssue) string {
	parts := []string{string(issue.Severity)}

	if issue.File != "" {
		parts = append(parts, issue.File)
	}

	parts = append(parts, issue.Description)

	return strings.Join(parts, ": ")
}

func reviewFindings(issues []ReviewIssue) string {
	lines := make([]string, 0, len(issues))

	for _, issue := range issues {
		line := string(issue.Severity) + " " + string(issue.Type)
		if issue.Requirement != "" {
			line += " (" + issue.Requirement + ")"
		}

		if issue.File != "" {
			line += " " + issue.File
		}

		line += ": " + issue.Description
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func verifyFindings(issues []Issue) string {
	lines := make([]string, 0, len(issues))

	for _, issue := range issues {
		lines = append(lines, "- "+issue.Description)
	}

	return "Verification failed:\n" + strings.Join(lines, "\n")
}
