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

---

## Golang (only for Go projects)

### Definition of done

After every code change, before reporting the task as complete, you **must**:

1. `go build ./...` — confirm it compiles.
2. `go test ./...` (or the relevant packages) — confirm tests pass.

Do not skip these steps, and don't rely on CI to catch what you missed.

### Code style

- Packages, files: lowercase + layer suffix (`userrepo`, `eventservice`), not (`user-service`, `event-service`)
- Errors should always be wrapped
- Name error variables `err`, unless that would cause shadowing or hide a wrapped/outer error you still need — only then use a qualified name (e.g. `jsonErr`)
- Functions ~80 lines max; return early on errors
- Always use the explicit two-line form. Never combine assignment and `nil` check in one `if` statement
- Never mute parse errors from database rows or external input. Always propagate them.


### `new` with arbitrary expressions (Go 1.26)

`new` now accepts any addressable expression, not just a type name. Use this to take the address of a computed value inline:

```go
// Before Go 1.26 — needed a temporary variable
t := loc.Time()
field = &t

// Go 1.26 — inline is fine
field = new(loc.Time())
```

Do **not** "fix" these into temporary-variable form — that is a regression, not an improvement.

### Comments

- Comment shared utilities (godoc style, like stdlib).
- Comment workarounds/hacks so others don't "fix" them.
- Comment confusing legacy code. If new code needs comments to be understood — rewrite it clearer.
- Do not add comments to obvious code (`// increment counter` above `counter++`).
