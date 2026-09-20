---
name: fix-todos
description: "Find the TODO markers this branch added (used to flag agent-written code for follow-up), then for each one: validate it is clear, unambiguous, and safe to act on. Ambiguous ones are reported as impossible to resolve; clear ones are fixed in place and the TODO removed. After fixing, run the project's Definition of Done (for Go: build, tests, linters) and report the list of TODOs with a short note on what was done — no code. VCS-agnostic — never assumes git; gathers the branch's changes via whatever version control the repo uses. Trigger phrases: 'поправь TODO', 'разбери TODO в ветке', 'исправь мои TODO', 'fix todos', 'resolve todos', '/fix-todos'."
---

# /fix-todos

## Usage

```
/fix-todos          # resolve TODOs only in the uncommitted changes — the lines you just added
/fix-todos branch   # resolve every TODO the current branch added on top of its parent
```

Two modes:

- **`/fix-todos` (default)** — scope is the TODO markers in the **uncommitted working copy**: the lines
  you just added but haven't committed yet. Nothing already committed is touched.
- **`/fix-todos branch`** — scope is every TODO the **current branch** added on top of its parent
  (committed and uncommitted). Do not ask for or accept an explicit base ref: the agent resolves the
  branch's parent itself (the branch's merge-base / upstream via the repo's own VCS tooling — `master`,
  then `main`, then the repo default when nothing else identifies the parent) and diffs against it.

Either way, pre-existing TODOs that were already in scope's base are out of scope and must be left
untouched.

## Goal

You (the user) review agent-written code by dropping `TODO:` comments where something needs fixing
(e.g. `// TODO: this should return early on nil`). This skill collects those TODOs, decides for each
whether the instruction is clear and safe enough to act on, applies the clear ones, and refuses the
ambiguous ones instead of guessing. A wrong "fix" from a misread TODO is worse than an honest "I can't
tell what you meant here" — so the bar for acting is high.

## Hard constraints

- **Only branch-added TODOs.** Resolve TODOs introduced by this branch (and the working copy). Never
  touch a TODO that already existed on the base ref.
- **VCS-agnostic — never assume git.** Not all code lives in git; gather the branch's changes through
  whatever version control the repository actually uses, and follow any contributor/LLM tooling the repo
  ships. Think in terms of the *actions* ("list what this branch changed", "get the branch diff vs its
  parent"), and let the repo's own tooling supply the exact commands. If you cannot tell how to read
  changes, check the repo's contributor/LLM guidance rather than guessing. If the repository is under no
  VCS at all, there is no base to diff against — resolve every TODO in scope, but **tell the user
  explicitly** that you did so because nothing was version-controlled.
- **Don't guess intent.** A TODO must be acted on only when its meaning is singular and the fix is
  obvious from the code around it. If it is vague ("fix this", "make it better"), has more than one
  reasonable interpretation, depends on context you don't have, or asks for a product/behavior decision
  you can't derive — do **not** implement it. Report it as unresolved (see Output format).
- **Safety first.** Do not apply a TODO whose fix would plausibly break behavior, change a public API,
  or ripple beyond the local change without you being confident it's correct and complete. When acting
  would require broad risky edits, treat it as unresolved and say why.
- **Minimal diff.** Change as little as possible to satisfy each TODO. Don't refactor or reformat
  unrelated code. Follow every rule in the applicable `CLAUDE.md`/`AGENTS.md` (comment policy, naming, error
  wrapping, tests, etc.).
- **Remove the TODO you resolved.** When a TODO is fixed, delete its comment line — its job is done.
  Leave untouched every TODO you did not resolve.
- **No new TODOs, no silencing.** Never satisfy a TODO by adding another comment, a `nolint`, or by
  disabling a check. Fix the actual code.

## Steps

### 1. Load the conventions
- Read the applicable `CLAUDE.md`/`AGENTS.md` file(s) (repo root and any nested under the changed paths) and any
  repo conventions/linter configs. Hold these as the rules for both the fixes and the DoD.

### 2. Resolve the scope and collect the TODOs
- Pick the scope from the mode (see Usage): default `/fix-todos` → the uncommitted working copy only;
  `/fix-todos branch` → the whole branch vs its parent. For `branch`, resolve the parent yourself via
  the repo's VCS tooling (merge-base / upstream, else `master` → `main` → repo default); if the repo is
  under no VCS at all, fall back to every TODO and tell the user (see Hard constraints).
- Get the diff for that scope and find every **added** line that introduces a `TODO` marker
  (`// TODO:`, `# TODO:`, `TODO(...)`, etc. — match the languages in the repo).
- For each TODO, read the surrounding code (not just the diff hunk) to understand what it refers to.
- If the branch added no TODOs, say so and stop.

### 3. Triage each TODO
For every collected TODO decide: **✅ clear & safe** or **❌ unresolved**.
- ✅ Clear & safe → the instruction has one obvious meaning and a bounded, low-risk fix.
- ❌ Unresolved → ambiguous, underspecified, needs a decision you can't derive, or the fix would be
  risky or far-reaching. Record a one-line reason ("cannot resolve — ambiguous: …").

Before applying anything, print this triage to the user: the two lists (clear & safe vs unresolved,
each item as `file:line` + one line) so they can see up front which TODOs will be fixed and which won't.

### 4. Apply the clear fixes
- Implement each clear & safe TODO with the smallest correct change, obeying the loaded conventions.
- Delete the resolved TODO comment.
- Leave every unresolved TODO exactly as it is.

### 5. Run the Definition of Done
Only if you changed code. Run the project's Definition of Done for the affected languages and make it
pass — see **Definition of done** in the global `CLAUDE.md`/`AGENTS.md` for the per-language gates (e.g. Go: build,
tests, linters). If a fix breaks the DoD and you can't cleanly resolve it, revert that one fix and move
the TODO to the unresolved list with the failure as the reason. Never leave the tree broken.

### 6. Emit the report
Output the report described below — no code snippets.

## Output format

Two sections. In each list item, reference the TODO by a clickable `file:line`.

```
## ✅ Fixed
- `path/to/file.go:42`: {{ what the TODO asked, one line }} → {{ what you changed, one line — no code }}

## ❌ Unresolved
- `path/to/file.go:88`: {{ the TODO text }} → cannot resolve: {{ why — ambiguity / missing decision / risk }}

## DoD
{{ build / tests / linters: pass or fail, one line each; or "skipped — no code changed" }}
```

Rules:
- If a section is empty, keep its heading and write `—` under it.
- Descriptions are terse and describe *what* changed, never *how* in code.
- Close with one line: either "All branch TODOs resolved." or "N TODO(s) left unresolved — see above."
