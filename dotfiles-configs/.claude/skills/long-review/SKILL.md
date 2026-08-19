---
name: long-review
description: "Long review of a whole branch relative to its base branch (master/main/the repo default), not just the uncommitted working copy: every commit the branch adds on top of the base. Reviews functionality/bugs/correctness first, then standards, patterns, and conventions. Always reads CLAUDE.md and any project conventions before reviewing. VCS-agnostic — never assumes git; gathers changes via whatever version control the repo uses. Trigger phrases: 'сделай длинное ревью', 'отревьюй всю ветку', 'ревью ветки относительно мастера', 'long review', '/long-review'. Read-only: reports a grouped list of findings and never edits, stages, or commits."
---

# /long-review

## Usage

```
/long-review                 # review the whole branch vs its base (master, then main, then repo default)
/long-review <base-ref>      # compare against an explicit base branch/ref instead
```

Scope is everything the current branch adds on top of its base branch — all commits since it diverged,
plus any uncommitted working-copy changes on top. Unlike `/short-review` (which looks only at
uncommitted changes vs the current commit), this reviews the branch as a whole, the way a reviewer
would see it in a merge/pull request.

Base-branch resolution when no `<base-ref>` is given: prefer `master`, then `main`, then the
repository's configured default branch. If none of these exist and none was passed explicitly, ask the
user which base ref to diff against — do not guess.

## Goal

Give the user a focused review of their entire branch before it goes up for merge, so they can fix
issues while it is still theirs to rewrite. A review always goes in order of impact: **first**
correctness — does it work, are there bugs, logic errors, missed edge cases, broken invariants — and
**only then** the softer layers: project standards, established patterns, naming conventions, style,
spelling. A subtle functional bug matters more than a style nit, and the report must reflect that
ordering.

## Hard constraints

See **Code review → Hard constraints** in `CLAUDE.md` — they are shared across all review skills and
authoritative here. Scope for this skill is the whole branch vs its base branch (see Usage). Two
scope-specific additions:

- **Review the branch's net effect, not its history.** Judge the final state of the code the branch
  produces (the cumulative diff against the base), not each intermediate commit. A bug introduced in an
  early commit and fixed in a later one is not a finding; a bug still present in the final diff is.
- **Mind the base ref.** Diff against the merge-base of the branch and its base (i.e. only what the
  branch itself adds), not against the raw tip of the base branch — otherwise unrelated changes that
  landed on the base after divergence would leak into the review.

## Steps

### 1. Load the conventions
- Read the applicable `CLAUDE.md` file(s) and any repo conventions/linter configs (see Hard
  constraints). Hold these as the review checklist.

### 2. Resolve the base and gather the branch changes
- Determine the base ref: the explicit `<base-ref>` argument, else `master` → `main` → repo default (see
  Usage). If none can be resolved, ask and stop.
- Get the cumulative diff of the current branch against the merge-base with that base ref — every file
  the branch adds, modifies, or deletes across all its commits.
- Include uncommitted working-copy changes on top, and read any new/untracked files in full — they may
  not show up in a diff at all.
- Do all of this with the repository's own version-control tooling (see the VCS-agnostic constraint);
  do not assume git. If the repo documents how an LLM should read changes, follow that.
- If the branch adds nothing over its base, say so and stop.

### 3. Review pass 1 — functionality (highest priority)
Look for bugs, logic errors, wrong conditions, off-by-one, nil/undefined access, unhandled or muted
errors, missed edge cases, broken invariants, concurrency issues, incorrect results, and anything that
makes the code not do what it clearly intends. Open the relevant files with `Read` to confirm — and
because the scope is larger than one working copy, also watch for cross-file/cross-commit issues:
inconsistent changes to a caller and its callee, a rename applied in some places but not others, dead
code left behind by a mid-branch pivot.

### 4. Review pass 2 — standards, patterns, conventions
Only after pass 1: check against the loaded conventions — naming, structure, error wrapping, function
size, forbidden/allowed comments, formatting, established patterns in the surrounding code, plus plain
spelling/grammar in strings, docs, and identifiers.

### 5. Emit the report
Produce the grouped report described in **Output format** in `CLAUDE.md`. Order the whole report by
impact: functional/correctness findings before convention/style ones, and within one block `high`
before `medium` before `low`. The finding index is global and continuous across all blocks (it does
not reset per block).

### 6. Close
End with one short line: either "No issues found" if the list is empty, or a one-line reminder that this
covers the whole branch against `<base-ref>`.

## Output format

See **Code review → Output format** in `CLAUDE.md` — the grouped, impact-ordered report format is
shared across all review skills. If a review turns up nothing, output no blocks and just write
`No issues found`.
