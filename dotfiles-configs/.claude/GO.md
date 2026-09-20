# Go rules

These rules apply only to Go projects, on top of the general rules in [CLAUDE.md](CLAUDE.md).

## Preferred stack

- In an existing project, use what the project already uses. Do not migrate it to the stack
  below unless I ask.
- In a new project, use Go 1.27 and:
  - `fiber` for HTTP — take the setup from https://github.com/TeaDove/teasutils/tree/master/fiberutils
  - `gorm` for database access — take the setup from https://github.com/TeaDove/teasutils/tree/master/serviceutils/dbutils
  - `zerolog` for logging — take the setup from https://github.com/TeaDove/teasutils/tree/master/serviceutils/loggerutils

## Definition of done

After every code change, before reporting the task as complete, you **must**:

1. `go build ./...` — confirm it compiles.
2. `go test ./...` (or the relevant packages) — confirm tests pass.
3. `golangci-lint run ./...` / `pre-commit run -a` - confirm lints work.

Do not skip these steps, and don't rely on CI to catch what you missed.

## Code style

- Packages, files: lowercase + layer suffix (`userrepo`, `eventservice`), not (`user-service`, `event-service`)
- Prefer long variable names, e.g. `queue := NewQueue()`, not `q := NewQueue()`, with exceptions like `ctx`, `i` (in loops), etc.
- Errors should always be wrapped
- Name error variables `err`, unless that would cause shadowing or hide a wrapped/outer error you still need — only then use a qualified name (e.g. `jsonErr`)
- Always use the explicit two-line form. Never combine assignment and `nil` check in one `if` statement

## `new` with arbitrary expressions (Go 1.26)

`new` now accepts any expression, not just a type name. Use this to take the address of a computed value inline:

```go
// Before Go 1.26 — needed a temporary variable
name := "John"
field = &name

// Go 1.26 — inline is fine
field = new("John")
```

Do **not** "fix" these into temporary-variable form — that is a regression, not an improvement.

## Tests

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
