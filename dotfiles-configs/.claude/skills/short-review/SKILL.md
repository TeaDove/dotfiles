---
name: short-review
description: "Short review of the uncommitted changes in the working copy relative to the current commit: modified files plus new/untracked files. Reviews functionality/bugs/correctness first, then standards, patterns, and conventions. Always reads CLAUDE.md and any project conventions before reviewing. VCS-agnostic — never assumes git; gathers changes via whatever version control the repo uses. Trigger phrases: 'сделай короткое ревью', 'отревьюй незакоммиченные изменения', 'короткое ревью', 'short review', '/short-review'. Read-only: reports a grouped list of findings and never edits, stages, or commits."
---

# /short-review

## Usage

```
/short-review        # review all uncommitted changes vs the current commit
```

Scope is the uncommitted changes in the working copy — modified files plus new/untracked files —
compared against the last commit on the current branch. It deliberately ignores already-committed
work; for a whole-branch review against a base branch use `/code-review` or `/review` instead.

## Goal

Give the user a fast, focused review of what they have changed but not yet committed, so they can fix
it before committing. A review always goes in order of impact: **first** correctness — does it work,
are there bugs, logic errors, missed edge cases, broken invariants — and **only then** the softer
layers: project standards, established patterns, naming conventions, style, spelling. A subtle
functional bug matters more than a style nit, and the report must reflect that ordering.

## Hard constraints

- **Read the rules first.** Before reviewing a single line, read every `CLAUDE.md` that applies (repo
  root and any nested ones under the changed paths), plus any conventions the repo ships —
  `CONVENTIONS*`, `CONTRIBUTING*`, `AGENTS.md`, and linter configs (`.golangci.yml`, `.editorconfig`,
  ESLint/Prettier configs, etc.). Review against *those* rules, not generic taste. A finding that
  contradicts the project's own stated convention is itself a bug in the review.
- **VCS-agnostic — never assume git.** Gather and describe changes through whatever version control the
  repository uses. Big monorepos ship their own VCS and their own LLM/contributor instructions for it
  (for example `arc`); follow those. Do not hardcode a specific tool's commands — think in terms of the
  *actions* ("list the changed files", "get the diff against the current commit"), and let the repo's
  own tooling/instructions supply the exact commands. If you cannot tell how to read changes, check the
  repo's contributor/LLM guidance rather than guessing.
- **Read-only.** Never edit, stage, commit, or push. This skill only reports. Do not "helpfully" fix
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

## Steps

### 1. Load the conventions
- Read the applicable `CLAUDE.md` file(s) and any repo conventions/linter configs (see Hard
  constraints). Hold these as the review checklist.

### 2. Gather the uncommitted changes
- List every uncommitted path: modified, added, deleted, and new/untracked files.
- Get the diff of those changes against the current commit.
- Read new/untracked files in full — they may not show up in a diff at all.
- Do all of this with the repository's own version-control tooling (see the VCS-agnostic constraint);
  do not assume git. If the repo documents how an LLM should read changes, follow that.
- If there are no uncommitted changes at all, say so and stop.

### 3. Review pass 1 — functionality (highest priority)
Look for bugs, logic errors, wrong conditions, off-by-one, nil/undefined access, unhandled or muted
errors, missed edge cases, broken invariants, concurrency issues, incorrect results, and anything that
makes the code not do what it clearly intends. Open the relevant files with `Read` to confirm.

### 4. Review pass 2 — standards, patterns, conventions
Only after pass 1: check against the loaded conventions — naming, structure, error wrapping, function
size, forbidden/allowed comments, formatting, established patterns in the surrounding code, plus plain
spelling/grammar in strings, docs, and identifiers.

### 5. Emit the report
Produce the grouped report described in **Output format**. Order the whole report by impact:
functional/correctness findings before convention/style ones, and within one block `high` before
`medium` before `low`. The finding index is global and continuous across all blocks (it does not reset
per block).

### 6. Close
End with one short line: either "No issues found" if the list is empty, or a one-line reminder that this
covers only uncommitted changes against the current commit.

## Output format

Group findings by **error type + importance**. Each block is one error type at one importance level: a
line with the error type, then a colour-coded importance label (`🟥 high:`, `🟨 medium:`, `🟩 low:`),
then the list. Every list item is: the global index, a clickable `file:line` reference, a short
description, then a newline, then the full description.

```
{{ Error type (functional, spelling, conventions, etc.) }}
{{ 🟥 high: | 🟨 medium: | 🟩 low: }}
{{ index }}. {{ file:line }}: {{ short description }}:
  {{ full description }}
```

### Example

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
