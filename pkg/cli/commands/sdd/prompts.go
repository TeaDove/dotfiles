package sdd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const issuesSchema = `"issues":{"type":"array","items":{"type":"object","additionalProperties":false,` +
	`"required":["description"],"properties":{"description":{"type":"string"}}}}`

func jsonEnum[T ~string](values ...T) string {
	quoted := make([]string, len(values))

	for i, value := range values {
		quoted[i] = strconv.Quote(string(value))
	}

	return "[" + strings.Join(quoted, ",") + "]"
}

func statusProperty(values ...Status) string {
	return `"status":{"type":"string","enum":` + jsonEnum(values...) + `},`
}

var validateSchema = json.RawMessage(`{"type":"object","additionalProperties":false,` +
	`"required":["status","issues"],"properties":{` +
	statusProperty(StatusREADY, StatusBLOCKED) + issuesSchema + `}}`)

var planSchema = json.RawMessage(`{"type":"object","additionalProperties":false,` +
	`"required":["status","issues","requirementCount","plan"],"properties":{` +
	statusProperty(StatusREADY, StatusBLOCKED) +
	`"requirementCount":{"type":"integer"},"plan":{"type":"string"},` + issuesSchema + `}}`)

var implementSchema = json.RawMessage(`{"type":"object","additionalProperties":false,` +
	`"required":["status","summary","questions","issues"],"properties":{` +
	statusProperty(StatusDONE, StatusBLOCKED) +
	`"summary":{"type":"string"},"questions":{"type":"array","items":{"type":"string"}},` + issuesSchema + `}}`)

var verifySchema = json.RawMessage(`{"type":"object","additionalProperties":false,` +
	`"required":["status","checks","issues","blockedReason"],"properties":{` +
	statusProperty(StatusPASSED, StatusFAILED, StatusBLOCKED) +
	`"blockedReason":{"type":"string"},` +
	`"checks":{"type":"array","items":{"type":"object","additionalProperties":false,` +
	`"required":["command","passed","detail"],"properties":{` +
	`"command":{"type":"string"},"passed":{"type":"boolean"},"detail":{"type":"string"}}}},` + issuesSchema + `}}`)

var reviewSchema = json.RawMessage(`{"type":"object","additionalProperties":false,` +
	`"required":["status","issues"],"properties":{` +
	statusProperty(StatusAPPROVED, StatusREJECTED, StatusBLOCKED) +
	`"issues":{"type":"array","items":{"type":"object","additionalProperties":false,` +
	`"required":["severity","type","requirement","file","description"],"properties":{` +
	`"severity":{"type":"string","enum":` + jsonEnum(SeverityValues()...) + `},` +
	`"type":{"type":"string","enum":` + jsonEnum(FindingTypeValues()...) + `},` +
	`"requirement":{"type":"string"},"file":{"type":"string"},"description":{"type":"string"}}}}}}`)

func specHeader(specPath, specContent string) string {
	return fmt.Sprintf("Specification file: %s\nRepository root: current working directory.\n\n"+
		"--- BEGIN SPECIFICATION ---\n%s\n--- END SPECIFICATION ---\n", specPath, specContent)
}

func validatePrompt(specPath, specContent string) string {
	return specHeader(specPath, specContent) + `
You are the specification gate for a Spec-Driven Development (SDD) workflow. You are the last line of
defense before code is written, so be rigorous and skeptical. You may inspect the repository to judge
sufficiency in context, but do NOT modify any file.

Decide whether the specification pins down the requested EXTERNALLY OBSERVABLE BEHAVIOR precisely enough
that any competent engineer would produce the same observable results, WITHOUT inventing product
requirements.

Check every requirement individually:
1. Enumerate each requirement / command / endpoint / output the spec describes.
2. For each, verify the observable contract is fully determined. In particular:
   - Every described response/output MUST have a concrete shape: exact fields, their types, and their
     meaning. A response described only in prose ("returns the system load", "returns status") with no
     field list is UNDERSPECIFIED -> BLOCKED.
   - When a described field maps to data that can legitimately have several values (e.g. an entity with
     multiple addresses mapped to a single field), the spec MUST say which value or how to pick. If not
     -> BLOCKED.
   - Failure/edge-case semantics that change what the caller observes (error status, error body,
     empty/absent data, conflicts) must be defined where they matter.
3. Ask: could two competent engineers each satisfy the literal text yet produce OBSERVABLY DIFFERENT
   results? If yes -> BLOCKED.
4. If requirement B depends on the output of requirement A and A is underspecified, B is too.

NOT reasons to block (safely the implementer's choice): internal implementation details that do not
change observable behavior (concurrency primitive, file/module layout, algorithm, libraries, logging);
and stylistic latitude the spec explicitly grants ("minimalistic", "responsive"). Do not fill any gap
with a reasonable assumption; if you must invent a field, schema, status code, or selection rule the
spec does not state, that itself is a BLOCKED item.

Return READY only if EVERY requirement passes the checks above. Otherwise return BLOCKED and list, in
issues, each concrete thing the specification must clarify (name the endpoint/requirement and the
missing decision). Prefer BLOCKED when genuinely unsure.

Return only the structured result via the StructuredOutput tool.`
}

func planPrompt(specPath, specContent string) string {
	return specHeader(specPath, specContent) + `
Build an implementation plan from the approved specification for an SDD workflow.

Inspect the repository. Do NOT modify any file.

Enumerate every specification requirement and map each to concrete implementation and
verification work (files to change, tests to add). Every requirement must map to at least one
plan item, and every planned change must trace back to the specification or necessary support.
Set requirementCount to the number of distinct requirements you enumerated.
Put the full traceable plan (requirement -> work -> verification) into the plan field.

If repository inspection reveals the specification is actually insufficient, return BLOCKED
with the gaps in issues instead of guessing.

Return only the structured result via the StructuredOutput tool.`
}

func implementPrompt(specPath, specContent, plan string) string {
	return specHeader(specPath, specContent) + fmt.Sprintf(`
--- BEGIN PLAN ---
%s
--- END PLAN ---

Implement the approved specification according to the plan above, for an SDD workflow.

The specification is authoritative. Follow existing repository conventions. Make the smallest
reasonable change. Do NOT perform unrelated refactoring or cleanup. Add or update meaningful
tests. Preserve backwards compatibility unless the specification says otherwise. Implement the
COMPLETE plan.

If implementation requires a product or behavioral decision absent from the specification, stop
and return BLOCKED with the open questions in the questions field rather than inventing behavior.
Otherwise return DONE with a short summary of what you changed.

Return only the structured result via the StructuredOutput tool once the code changes are done.`, plan)
}

func verifyPrompt(specPath, specContent string) string {
	return specHeader(specPath, specContent) + `
Verify the current repository state for an SDD workflow. Do NOT modify any file.

Inspect the project (Makefile, Taskfile, justfile, package.json, pyproject.toml, go.mod,
.github/workflows, README, CLAUDE.md) and determine the appropriate verification commands.
Prefer commands already used by the project's CI or documented development workflow. Run them
(tests, linters, static analysis, build, race detector where appropriate). Record each command
and whether it passed in checks.

Return PASSED only if every required verification command actually ran and succeeded.
Return FAILED if a command failed because of the implementation; put the failures in issues.
Return BLOCKED (with blockedReason) if a required verification step could not be performed due to
a missing external dependency or environment/tooling problem that you must not work around.
Do not report PASSED if required verification could not actually be performed.

Return only the structured result via the StructuredOutput tool.`
}

func fixPrompt(specPath, specContent, findings string) string {
	return specHeader(specPath, specContent) + fmt.Sprintf(`
--- BEGIN FINDINGS ---
%s
--- END FINDINGS ---

Fix the findings above for an SDD workflow. The specification remains authoritative.

Address every finding. Do NOT reinterpret the specification. Do NOT make unrelated changes. Do
not blindly obey a reviewer suggestion that conflicts with the specification. If a finding reveals
missing specification information (a real spec gap), return BLOCKED with the open questions in the
questions field instead of inventing behavior. Otherwise return DONE.

Return only the structured result via the StructuredOutput tool once the fixes are done.`, findings)
}

func reviewPrompt(specPath, specContent, preexisting string, vcsKnown bool) string {
	return specHeader(specPath, specContent) + fmt.Sprintf(`
You are an independent SDD reviewer. The specification is the source of truth.

Review the current implementation in the repository against the specification. Inspect the actual
repository and the current changes (for example via 'git diff' and 'git status'); do not trust any
summary. Do NOT modify files. Do NOT propose unrelated refactors. Do NOT reject code merely
because you prefer another valid design.

%s

Look for: missing requirements; incorrect behavior; unhandled failure/edge cases; insufficient or
non-meaningful tests; regressions introduced by the change; out-of-scope changes not justified by
the specification; and behavior invented beyond the specification.

Classify every finding with severity (BLOCKER, MAJOR, MINOR) and type (IMPLEMENTATION_BUG,
TEST_GAP, SPEC_GAP, OUT_OF_SCOPE). Set requirement and file where applicable.

Return APPROVED only if there are no BLOCKER or MAJOR issues and no specification violations.
Return REJECTED with the issues otherwise. If correct implementation requires information not
present in the specification, report the finding as type SPEC_GAP and status BLOCKED instead of
inventing an answer.

Return only the structured result via the StructuredOutput tool.`, preexistingNote(preexisting, vcsKnown))
}

func preexistingNote(preexisting string, vcsKnown bool) string {
	if !vcsKnown {
		return "Version control state could not be determined (git is unavailable here), so pre-existing " +
			"changes cannot be separated automatically. Review the current repository state and use judgment " +
			"about which changes belong to this workflow."
	}

	if strings.TrimSpace(preexisting) == "" {
		return "The whole working tree was clean before implementation, so every change is attributable to this workflow."
	}

	return "The following paths already had uncommitted user changes BEFORE this workflow started; " +
		"do not attribute those pre-existing changes to the implementation:\n" + preexisting
}
