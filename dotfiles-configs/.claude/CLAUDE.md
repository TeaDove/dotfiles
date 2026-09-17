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

## Version control

- Use any VCS (git, SVN, etc.) **read-only**. Reading history and state is always
  allowed: status, log, diff, blame, show, listing branches, inspecting commits, etc.
- Mutating the repository is FORBIDDEN: do not commit, push, pull, fetch, rebase, merge,
  create/switch/delete branches, stash, cherry-pick, reset, tag, or open PRs — and do not
  offer or suggest doing any of these. I handle all of that myself.

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
  repository uses. Big monorepos ship their own VCS and their own LLM/contributor instructions for it;
  follow those. Do not hardcode a specific tool's commands — think in terms of the
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
{{ 🟥 high | 🟨 medium | 🟩 low }}
{{ index }}. {{ file:line }}: {{ short description }}:
  {{ full description }}
```

Example:

```
Functional
🟥 high
1. `pkg/cli/install.go:153`: target overwritten instead of merged:
  When a key is missing in dst the value is taken from src, but an existing hooks
  array is replaced wholesale — confirm this is intended on re-runs.

Conventions
🟨 medium
2. `pkg/cli/install.go:92`: comment outside the allowed NOFIX/LEGACY set:
  CLAUDE.md forbids comments except NOFIX:/LEGACY: — this godoc should be removed.

Spelling
🟩 low
3. `dotfiles-configs/.claude/CLAUDE.md:44`: typo "imposible":
  Should be "impossible".
```

If a review turns up nothing, output no blocks and just write `No issues found`.

---

## General (all languages)

### Code style

- Functions ~80 lines max; return early on errors
- Never mute parse errors from database rows or external input. Always propagate them or log.

### Comments

Do not write comments. The only allowed forms are `NOFIX:` and `LEGACY:` (below). Everything else is
forbidden, however idiomatic it looks.

**Before you type a comment (`//`, `/* */`, `#`, or a Python docstring `"""..."""`), stop.** If it does
not start with `NOFIX:` or `LEGACY:` — and is not a compiler directive (`//go:...`) or an existing
`//nolint:...` — do not write it.

**Forbidden, even though language tooling/linters expect them:**
- doc comments on packages/modules, types/classes, funcs/methods, fields, constants — *including exported/public ones* (godoc in Go, docstrings in Python);
- "what it does" / "why we do this" explanations, rationale, design notes, summaries;
- section headers, banners, dividers;
- commented-out code;
- `TODO`/`FIXME`/`XXX` — a `TODO:` is allowed only appended to a `NOFIX:`/`LEGACY:` line.

If code needs a comment to be understood, rename or split it until it doesn't. Explanation belongs in the
commit message or the ticket, not the source. A godoc comment (even on an exported symbol) is a violation
here, not an exception. Do not smuggle an explanation in by labelling it `NOFIX:`.

#### `NOFIX:` — a workaround a reader would otherwise "fix" and break
It MUST name the concrete wrong fix and why it breaks. If you cannot name the specific misfix, it is not a
NOFIX — delete it. It is not a license to explain code.
```go
type User struct{
    // NOFIX: "nam" is a typo, but clients already use it; renaming to "name" is a breaking API change.
    Name string `json:"nam"`
}
```

#### `LEGACY:` — flag existing confusing legacy code you did not write and cannot change now
New code never gets a LEGACY comment — rewrite it to be clear instead. Add `TODO:` on the next line if a
follow-up is warranted, e.g.:
```go
// LEGACY: the frontend sends the `left` point in the JSON object under the key `right` and vice versa, so we swap them.
// TODO: fix the confusing frontend/backend names.
leftPoint, rightPoint = rightPoint, leftPoint
```

## Python (only for Python projects)

### Definition of done

After every code change, before reporting the task as complete, you **must** run whatever the project ships:

1. Run the code — confirm it executes.
2. Run the tests (e.g. `pytest`) — confirm they pass.
3. Run the formatters/linters (`ruff`, `black`, `pre-commit run -a`, depending on the project) — confirm they pass.

Do not skip these steps, and don't rely on CI to catch what you missed.

### Code style

- Always annotate types.
- When working with JSON, always use pydantic.
- Define classes with `@dataclass`.
- In `.ipynb` notebooks: keep all imports in a single cell at the very top of the file; put all settings (e.g. chart colors) in the second cell.

## Golang (only for Go projects)

### Definition of done

After every code change, before reporting the task as complete, you **must**:

1. `go build ./...` — confirm it compiles.
2. `go test ./...` (or the relevant packages) — confirm tests pass.
3. `golangci-lint run ./...` / `pre-commit run -a` - confirm lints work.

Do not skip these steps, and don't rely on CI to catch what you missed.

### Code style

- Packages, files: lowercase + layer suffix (`userrepo`, `eventservice`), not (`user-service`, `event-service`)
- Prefer long variable names, e.g. `queue := NewQueue()`, not `q := NewQueue()`, with exceptions like `ctx`, `i` (in loops), etc.
- Errors should always be wrapped
- Name error variables `err`, unless that would cause shadowing or hide a wrapped/outer error you still need — only then use a qualified name (e.g. `jsonErr`)
- Always use the explicit two-line form. Never combine assignment and `nil` check in one `if` statement

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

### Tests

- Use `t.Context()`, not `context.Background()`.
- Always call `t.Parallel()` (and `tt.Parallel()` inside subtests).
- Never pass a message to an assertion.

Good:
```go
require.Equal(t, id, user.ID)
```

- Group several related cases into one table-driven test:
```go
func TestAvg(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        name string
        arr  []int
        exp  float64
    }{
        {
            name: "same numbers",
            arr:  []int{2, 2, 2},
            exp:  2,
        },
        {
            name: "simple 3",
            arr:  []int{1, 2, 3},
            exp:  2,
        },
        {
            name: "negative values",
            arr:  []int{-2, 0, 2},
            exp:  0,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(tt *testing.T) {
            tt.Parallel()

            require.InDelta(tt, tc.exp, avg(tc.arr), 0.00001)
        })
    }
}
```
