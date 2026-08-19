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
