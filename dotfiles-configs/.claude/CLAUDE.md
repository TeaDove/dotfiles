# Global user preferences

These rules apply to every project I work on (work and personal). Language-specific
sections at the bottom apply only to that language.

## Language

- All artifacts — code, comments, READMEs, docs, commit messages — in English.
- Chat replies: answer in the language I started the conversation in.

## Working style

- Minimal diff: solve the task by changing as little code as possible. Smaller
  changes are easier to review and less likely to introduce bugs. Don't refactor or
  reformat unrelated code unless asked.

## Rules, linters, CI/CD

- Never break CI/CD checks, linter rules, or the conventions in this file. You may
  bypass a rule (e.g. add a `//nolint` directive, disable a check, skip a gate) ONLY
  when I explicitly allow it for that case, or when that exact rule is already listed
  among documented exceptions (in this file or other README/DOCS/CLAUDE.md).
- Default to the strictest interpretation of every rule. For example, global
  variables are forbidden unless I explicitly say otherwise. When a rule blocks you,
  ask me instead of silencing or working around it on your own.

## SVN

- In SVN repositories, NEVER perform mutating actions: do not create branches, do not
  commit, do not open PRs, and do not offer or suggest doing any of these. I handle
  all of that myself.

## Code review (skills: short-review, long-review)

These rules apply whenever you run a code review. The review skills define *what* scope to
review (`/short-review` = uncommitted changes; `/long-review` = the whole branch vs its base);
the Hard constraints and Output format below are shared and authoritative for both.

### Hard constraints

- **Read the rules first.** Before reviewing a single line, read every `CLAUDE.md` that applies (repo
  root and any nested ones under the changed paths), plus any conventions the repo ships —
  `CONVENTIONS*`, `CONTRIBUTING*`, `AGENTS.md`, and linter configs (`.golangci.yml`, `.editorconfig`,
  ESLint/Prettier configs, etc.). Review against *those* rules, not generic taste. A finding that
  contradicts the project's own stated convention is itself a bug in the review.
- **VCS-agnostic — never assume git.** Gather and describe changes through whatever version control the
  repository uses. Big monorepos ship their own VCS and their own LLM/contributor instructions for it
  (for example `arc`); follow those. Do not hardcode a specific tool's commands — think in terms of the
  *actions* ("list the changed files", "get the diff for the reviewed scope"), and let the repo's own
  tooling/instructions supply the exact commands. If you cannot tell how to read changes, check the
  repo's contributor/LLM guidance rather than guessing.
- **Read-only.** Never edit, stage, commit, or push. A review only reports. Do not "helpfully" fix
  anything — the user fixes it themselves after reading the findings.
- **Ground every finding in the actual change.** Read the surrounding code with `Read` (not just the
  changed hunk) before claiming something is wrong — enough context to be sure the issue is real and
  the line reference is correct. Do not invent problems to fill the list; an empty list is a valid,
  good result.
- **No false positives over volume.** Each finding must be defensible. If you are unsure whether
  something is actually wrong, either verify it or leave it out. Prefer a short list of real issues to
  a long list padded with maybes.
- **Respect the project's comment/style rules.** In this user's Go projects, for example, comments are
  forbidden except `NOFIX:`/`LEGACY:` — so "missing doc comment" is NOT a valid finding here, and an
  added non-`NOFIX`/`LEGACY` comment IS one. Always defer to the loaded CLAUDE.md over defaults.
- **Output language:** English — the user reads the report (rather than editing it) and prefers English
  for reviews, so write the whole report, including every finding, in English.
- **Keep it short.** Terse descriptions, no preamble, no restating the diff. The user wants the list,
  not an essay.

### Output format

Group findings by **error type + importance**. Each block is one error type at one importance level: a
line with the error type, then a colour-coded importance label (`🟥 high:`, `🟨 medium:`, `🟩 low:`),
then the list. Every list item is: the global index, a clickable `file:line` reference, a short
description, then a newline, then the full description. Order the whole report by impact:
functional/correctness findings before convention/style ones, and within one block `high` before
`medium` before `low`. The finding index is global and continuous across all blocks (it does not reset
per block).

```
{{ Error type (functional, spelling, conventions, etc.) }}
{{ 🟥 high: | 🟨 medium: | 🟩 low: }}
{{ index }}. {{ file:line }}: {{ short description }}:
  {{ full description }}
```

Example:

```
Functional
🟥 high:
1 `pkg/cli/install.go:153`: target overwritten instead of merged:
  When a key is missing in dst the value is taken from src, but an existing hooks
  array is replaced wholesale — confirm this is intended on re-runs.

Conventions
🟨 medium:
2 `pkg/cli/install.go:92`: comment outside the allowed NOFIX/LEGACY set:
  CLAUDE.md forbids comments except NOFIX:/LEGACY: — this godoc should be removed.

Spelling
🟩 low:
3 `dotfiles-configs/.claude/CLAUDE.md:44`: typo "imposible":
  Should be "impossible".
```

If a review turns up nothing, output no blocks and just write `No issues found`.

---

## Golang (only for Go projects)

### Definition of done

After every code change, before reporting the task as complete, you **must**:

1. `go build ./...` — confirm it compiles.
2. `go test ./...` (or the relevant packages) — confirm tests pass.

Do not skip these steps, and don't rely on CI to catch what you missed.

### Code style

- Packages, files: lowercase + layer suffix (`userrepo`, `eventservice`), not (`user-service`, `event-service`)
- Prefer long variable names, e.g. `queue := NewQueue()`, not `q := NewQueue()`, with exceptions like `ctx`, `i` (in loops), etc.
- Errors should always be wrapped
- Name error variables `err`, unless that would cause shadowing or hide a wrapped/outer error you still need — only then use a qualified name (e.g. `jsonErr`)
- Functions ~80 lines max; return early on errors
- Always use the explicit two-line form. Never combine assignment and `nil` check in one `if` statement
- Never mute parse errors from database rows or external input. Always propagate them.


### Comments
Never add comments in code, with only the following exceptions:

#### Comment workarounds/hacks that others couldn't guess, so they don't "fix" them
Start these comments with `NOFIX:`, e.g.:
```go
type User struct{
    // NOFIX: "nam" is a typo, but clients are already using it and we don't want to introduce breaking changes
    Name string `json:"nam"`
}
```

#### Comment confusing legacy code. If new code needs comments to be understood — rewrite it to be clearer
Start these comments with `LEGACY:`, and add `TODO:` if applicable, e.g.:
```go
// LEGACY: the frontend, for some reason, sends the `left` point in the JSON object under the key `right` and vice versa, so we need to swap them.
// TODO: fix the confusing frontend/backend names.
leftPoint, rightPoint = rightPoint, leftPoint
```

#### Never add a comment in any other case

Bad:
```go
// QueueFanout is a Queue that fans every write out to several Queue (e.g. a Kafka and a local JSONL file)
type QueueFanout struct {
  queues []Queue
}
```

Bad:
```go
// SaveMetricPoint writes the record to every queue; it does not stop on the
// first error, so one failing destination never starves the others.
func (f *QueueFanout) Send(ctx context.Context, record Record) error {
	var errs []error
	for _, q := range f.queues {
		err := q.Send(ctx, record)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
```

### `new` with arbitrary expressions (Go 1.26)

`new` now accepts any expression, not just a type name. Use this to take the address of a computed value inline:

```go
// Before Go 1.26 — needed a temporary variable
name := "John"
field = &name

// Go 1.26 — inline is fine
field = new("John")
```

Do **not** "fix" these into temporary-variable form — that is a regression, not an improvement.
