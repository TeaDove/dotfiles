# Global user preferences

These rules apply to every project I work on (work and personal). Language-specific
rules live in separate files linked at the bottom and apply only to that language.

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
  among documented exceptions (in this file or other README/DOCS/CLAUDE.md/AGENTS.md).
- Default to the strictest interpretation of every rule. For example, global
  variables are forbidden unless I explicitly say otherwise. When a rule blocks you,
  ask me instead of silencing or working around it on your own.

## Version control

- Use any VCS (git, SVN, etc.) **read-only** by default. Reading history and state is always
  allowed: status, log, diff, blame, show, listing branches, inspecting commits, etc.
- Every mutating operation is FORBIDDEN by default: commit, push, pull, fetch, rebase, merge,
  create/switch/delete branches, stash, cherry-pick, reset, tag, open/update/merge PRs. Do not
  offer or suggest doing any of these. I handle all of that myself.
- A project's `CLAUDE.md` or `AGENTS.md` may allow some of these operations explicitly.
  Silence means forbidden, and a permission covers only the operations it names. When
  several files apply, the narrower one (closer to the code) wins.
- When you rely on such a permission, say so before the first mutating operation and name
  the file and section it comes from.
- Report every mutating operation you perform, one line per operation, in this exact
  format, with an empty line before and after the block of lines:

  ```
  ▶ **{action}**: [{name}]({url})
  ```

  `action` is a past-tense verb (`committed`, `pushed`, `fetched`, `checked out`, `opened
  PR`, `tagged`). `name` is what identifies the result: the short hash for commits and
  pushes, the number for PRs, the name for tags, branches and stashes. Link it to the
  commit/PR/tag page when one exists; otherwise write `name` in plain text and without a
  link.

## SVN

- In SVN repositories, NEVER perform mutating actions: do not create branches, do not
  commit, do not open PRs, and do not offer or suggest doing any of these. I handle
  all of that myself.

## Privacy (PII)

Never put personal or identifying data into a public artifact.

- **Public artifact**: anything readable outside the workstation — commits (message and content),
  PR/MR titles, descriptions, comments and reviews, issues, gists, published packages and docs,
  CI logs. Treat every repository as public unless you have verified that it is private.
- **Never include**: names, emails, phone numbers, postal addresses, account and user IDs,
  public or private IP addresses, ports, hostnames and domains of my machines, network layout,
  MAC addresses, device serials, geolocation, SSH host keys and fingerprints, secrets and tokens,
  internal names of my employer's systems, and any data about other people.
- Describe such things neutrally ("the home server"). Where the value itself would go, refuse
  explicitly or write a visible redaction marker such as `[REDACTED IP ADDRESS 1]` — never a
  realistic-looking stand-in such as an example domain or a documentation IP range, so the reader
  sees the conflict with this rule at once.
- Before you create or edit a public artifact, check its full text for the items above.
  Debugging evidence (real addresses, keys, logs) stays in the chat.
- If a task seems to need PII in a public artifact, ask me first. If you find PII that is
  already published, tell me and do not repeat it anywhere.
- The VCS author identity already configured for the repository is allowed.

## Code review (skills: short-review, long-review)

These rules apply whenever you run a code review. The review skills define *what* scope to
review (`/short-review` = uncommitted changes; `/long-review` = the whole branch vs its base);
the Hard constraints and Output format below are shared and authoritative for both.

### Hard constraints

- **Read the rules first.** Before reviewing a single line, read every `CLAUDE.md`/`AGENTS.md` that
  applies (repo root and any nested ones under the changed paths), plus any conventions the repo ships —
  `CONVENTIONS*`, `CONTRIBUTING*`, and linter configs (`.golangci.yml`, `.editorconfig`,
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
- **Ground every finding in the actual change.** Read the surrounding code (not just the
  changed hunk) before claiming something is wrong — enough context to be sure the issue is real and
  the line reference is correct. Do not invent problems to fill the list; an empty list is a valid,
  good result.
- **No false positives over volume.** Each finding must be defensible. If you are unsure whether
  something is actually wrong, either verify it or leave it out. Prefer a short list of real issues to
  a long list padded with maybes.
- **Respect the project's comment/style rules.** In this user's Go projects, for example, comments are
  forbidden except `NOFIX:`/`LEGACY:` — so "missing doc comment" is NOT a valid finding here, and an
  added non-`NOFIX`/`LEGACY` comment IS one. Always defer to the loaded CLAUDE.md/AGENTS.md over
  defaults.
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
  CLAUDE.md/AGENTS.md forbids comments except NOFIX:/LEGACY: — this godoc should be removed.

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

## Language-specific rules

Read the matching file before working in that language; it applies on top of everything above.
When starting a new project, also read [NEWPROJECTS.md](NEWPROJECTS.md) (reference project, layout).

- Python: [PYTHON.md](PYTHON.md) — preferred stack, definition of done, code style.
- Go: [GO.md](GO.md) — preferred stack, definition of done, code style, tests.
