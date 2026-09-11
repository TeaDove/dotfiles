---
name: long-review
description: "Long review of a whole branch relative to its base branch (master/main/the repo default), not just the uncommitted working copy: every commit the branch adds on top of the base. Reviews functionality/bugs/correctness first, then standards, patterns, and conventions. When the branch maps to a ticket and a ticket-system MCP (Jira, Linear, GitHub Issues, etc.) is connected, also reads the ticket and checks the branch against its goal — flagging scope creep and missed requirements. Always reads CLAUDE.md and any project conventions before reviewing. VCS-agnostic — never assumes git; gathers changes via whatever version control the repo uses. Trigger phrases: 'сделай длинное ревью', 'отревьюй всю ветку', 'ревью ветки относительно мастера', 'long review', '/long-review'. Read-only: reports a grouped list of findings and never edits, stages, or commits."
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

When the branch maps to a ticket and the ticket system is reachable (see **Ticket alignment** below),
the review also judges the branch against *what the ticket actually asked for*: every change should
serve the ticket's goal, and the ticket's goal should be fully met. Changes that don't belong and
requirements that were missed are findings in their own right.

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

## Ticket alignment (optional pass)

If — and only if — the branch is tied to a ticket **and** a ticket-system MCP is connected, add a pass
that checks the branch against the ticket's stated goal. This is best-effort: when either condition is
missing, skip it silently and review the code as usual. Never block or delay the code review waiting on
a ticket.

- **Find the ticket id.** Look for a ticket key in the branch name, then in the branch's commit
  messages / PR title (e.g. `PROJ-1234`, `ABC-42`). If nothing looks like a ticket key, skip this pass.
  Do not invent or guess an id, and do not ask the user unless they explicitly requested ticket
  alignment.
- **Find the ticket system.** Use a connected MCP that fronts the issue tracker (Jira, Linear,
  GitHub Issues, etc.) — discover it via the available tools rather than assuming one. If no
  such MCP is available, skip this pass; mention in the closing line that ticket alignment was skipped
  for lack of a reachable tracker.
- **Read, don't write.** Only fetch/read the ticket (title, description, acceptance criteria,
  comments, linked subtasks). Never create, edit, comment on, or transition a ticket — the read-only
  constraint covers the tracker too.
- **Extract the intent.** From the ticket, distill *what actually needs to change* — the concrete goal
  and any acceptance criteria — before comparing against the diff. Judge the ticket's intent, not its
  wording; a vaguely worded ticket with a clear goal still counts.
- **Compare both directions.**
  - *Scope creep* — changes in the branch that don't serve the ticket's goal (unrelated refactors,
    drive-by edits, files touched for no stated reason). Flag them so the user can split or justify
    them. A genuinely necessary incidental change (e.g. updating a caller a required change forced) is
    not creep — only flag what the ticket's goal doesn't explain.
  - *Missed requirements* — parts of the ticket's goal or acceptance criteria the branch does not
    appear to address.
- Treat ticket-alignment findings as **functional** (they affect whether the branch does the right
  thing), and rank them accordingly in the report — see the added finding types in Output format.

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

### 2b. Resolve the ticket (optional)
Follow **Ticket alignment** above: try to extract a ticket id from the branch name / commits, and check
for a connected ticket-system MCP. If both are present, read the ticket and distill its goal and
acceptance criteria to compare against in pass 1. If either is missing, skip this and note it in the
closing line only if a tracker was expected but unreachable.

### 3. Review pass 1 — functionality (highest priority)
Look for bugs, logic errors, wrong conditions, off-by-one, nil/undefined access, unhandled or muted
errors, missed edge cases, broken invariants, concurrency issues, incorrect results, and anything that
makes the code not do what it clearly intends. Open the relevant files with `Read` to confirm — and
because the scope is larger than one working copy, also watch for cross-file/cross-commit issues:
inconsistent changes to a caller and its callee, a rename applied in some places but not others, dead
code left behind by a mid-branch pivot.

If a ticket goal was resolved in step 2b, also check the branch against it here: report *scope creep*
(changes the ticket's goal doesn't explain) and *missed requirements* (parts of the goal the branch
doesn't address), as described in **Ticket alignment**.

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
covers the whole branch against `<base-ref>`. If ticket alignment ran, note the ticket id it was checked
against; if it was expected (a ticket id was found) but the tracker was unreachable, say so.

## Output format

See **Code review → Output format** in `CLAUDE.md` — the grouped, impact-ordered report format is
shared across all review skills. If a review turns up nothing, output no blocks and just write
`No issues found`.

Ticket-alignment findings use two extra error-type blocks, grouped with the functional findings (they
rank as functional): `Scope creep` (changes the ticket doesn't explain) and `Missed requirement`
(ticket goals the branch doesn't address). Reference the concrete files/lines for scope creep; for a
missed requirement, reference the ticket criterion and, where possible, where it should have landed.
