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

See **Code review → Hard constraints** in `CLAUDE.md` — they are shared across all review skills and
authoritative here. Scope for this skill is the uncommitted changes vs the current commit (see Usage).

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

See **Code review → Output format** in `CLAUDE.md` — the grouped, impact-ordered report format is
shared across all review skills. If a review turns up nothing, output no blocks and just write
`No issues found`.
