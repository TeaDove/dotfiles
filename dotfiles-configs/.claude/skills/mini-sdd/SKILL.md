---
name: mini-sdd
description: "Lightweight single-file Spec-Driven Development for small tasks. Drives ONE spec file through requirements → research → design → code → independent review, recording every phase and its status flag inside that same file so work resumes after any interruption. The human owns only the IDEA section; the agent owns every derived artifact. Editing IDEA (or asking for a change in chat) marks the affected phases stale and cascades a minimal delta, then re-requests approval. Once everything is approved, implemented and review-passed, it switches to PR-fix mode. Never commits or pushes. Trigger phrases: '/mini-sdd', 'mini-sdd <spec-file>', 'запусти mini-sdd', 'реализуй по спеке через mini-sdd'."
---

# /mini-sdd — single-file Spec-Driven Development

## Usage

```
/mini-sdd <path-to-sdd-spec.md>
```

You are the orchestrator. Drive the whole flow yourself in this context with your normal tools.
The ONLY step that must run in a separate, fresh context is the independent review — launch it as a
fresh isolated agent (Claude Code: the `Task` tool with `subagent_type: sdd-reviewer`; Codex: a
separate `codex exec` run given `~/.codex/agents/sdd-reviewer.md`).

The path is required on the **first** run of a conversation: resolve it to one canonical absolute path
and record it in the report's `## project` header. On later `/mini-sdd` calls **in the same conversation the
path is optional** — reuse the spec path from the most recent canonical block; a bare `/mini-sdd` must
not fail just because the argument was omitted. Only if no path is given and none was established earlier
in this conversation: ask for one and stop. **If the resolved path (given or remembered) points to no
file, STOP with an error** — the spec is anchored to its directory; a missing spec is a hard failure,
never a reason to create one elsewhere or search around.

## The spec file is the single source of truth for state

All state lives in the spec file, never only in chat. Chat is an input channel; the moment you receive
an instruction you persist its consequence to the file, so a fresh agent can always recover from the
file alone. The file has this shape:

```markdown
# <title>

## IDEA            <!-- human-only -->
<free-text intent, written and edited only by the human>

## STATUS
State: <DRAFTING | AWAITING_APPROVAL | IMPLEMENTING | DoD | IN_REVIEW | DONE | PR_FIX>
Next: <one line: what you will do next, or what you are waiting for>

## REQUIREMENTS   [status: fresh | stale | provisional | approved]
<observable behaviour, acceptance criteria, edge cases, and any [NEEDS CLARIFICATION] open questions>

## RESEARCH       [status: fresh | stale | approved]
<links to all context (wiki / tickets / repo / code) + one short line: constraints / bans / what to reuse>

## DESIGN         [status: fresh | stale | provisional | approved]
<the HOW: components, contracts, key decisions. `provisional` while REQUIREMENTS has an open blocker.>

## CODE           [DoD: pending | pass | fail]
<what was implemented; the exact verification commands you ran and their results>

## REVIEW         [verdict: pending | APPROVED | REJECTED | BLOCKED]
<latest independent-review findings; iteration counter>
```

- **`STATUS` is a computed rollup**, placed right after IDEA for a quick glance. Recompute it from the
  per-phase flags on every change; it must never contradict them (no `DONE` while DESIGN is `stale`).
- **Ownership.** The human owns the `IDEA` section and all approvals; never rewrite IDEA. You own the
  *final* content of every derived section, but the human may also edit those sections directly in the
  file or drop `TODO:` notes in them — you reconcile such edits (see *Backup + human edits* below).
  Changes therefore arrive three ways: editing IDEA, editing an artifact section / leaving a `TODO:`,
  or asking in chat.
- **`fresh`** = you (re)generated it, awaiting the human's approval. **`approved`** = the human approved
  it. **`stale`** = an upstream change invalidated it, not yet reworked. **`provisional`** (design only)
  = built on a requirement with an unresolved blocker; hold its approval.

### Backup + human edits

Whenever you write an authoritative version of the file — after generating or reworking artifacts, and
before handing control back at a gate — copy it to a sibling `<spec>.backup`: the exact state you just
presented. `<spec>.backup` is scratch; keep it out of version control but keep it for the life of the project — it
is the *only* detector of later human edits, including changes made after `DONE`.

The human reviews by editing the artifact sections **directly in the file**, leaving `TODO:` notes where
they want you to change something. To pick those up, diff the current file against the backup:

```sh
diff <spec>.backup <spec>
```

Treat the human's direct edits as authoritative input to preserve, and each `TODO:` as an instruction to
fulfil. Fold them in (cascading downstream as needed), **remove every `TODO:` you handled**, refresh
`<spec>.backup`, and re-gate. If the diff shows the `IDEA` section changed, treat it as an IDEA change and
cascade from there.

## On every invocation: report first, then act

1. Resolve the spec path — the argument, or if omitted the one recorded in this conversation's last
   canonical block — then read the spec file. If `<spec>.backup` exists, `diff <spec>.backup <spec>` to detect every human edit
   since you last presented the file — an edited `IDEA`, edited artifact sections, or `TODO:` notes — and
   handle them as a change (below). An `IDEA` change cascades from IDEA down; an artifact edit cascades
   from that section down. (No backup yet — the first run — means nothing to diff: generate from IDEA.)
2. **Report status before doing anything else** so a bare invocation can be used just to look: emit the
   canonical block (see *Output*) — at least `project` (with the spec path), `status` and `next step`.
   Then act according to state — every turn also closes with that same block.

## Dispatch by state

- REQUIREMENTS / RESEARCH / DESIGN missing or not all `approved` → **Forward / cascade** flow.
- REQUIREMENTS + RESEARCH + DESIGN `approved` but `CODE DoD` not `pass` → **Implement**.
- Code done but `REVIEW` not `APPROVED` → **Review**.
- Everything `approved`, code `pass`, review `APPROVED` (State `DONE`) → **PR-fix** mode.

## Forward / cascade flow

### Generate the design artifacts (bulk)

From an approved/updated IDEA, generate `REQUIREMENTS`, `RESEARCH` and `DESIGN` together in one pass,
each terse, then set them `fresh` and ask for one bulk approval.

- **REQUIREMENTS** — the disambiguation-and-verification layer, not a paraphrase of IDEA. State the
  observable behaviour and acceptance criteria unambiguously; anything you would otherwise have to
  invent becomes a `[NEEDS CLARIFICATION]` open question. For a small task this is a handful of bullets.
- **RESEARCH** — links to all relevant context plus one line on constraints / banned dependencies /
  what to reuse.
- **DESIGN** — the technical HOW, traceable to the requirements. If REQUIREMENTS has an unresolved
  **blocking** open question, mark DESIGN `provisional` and say plainly that approving it now is
  premature; do not present it as ready.

### Approval gate

Before the gate, make sure `<spec>.backup` mirrors exactly what you are presenting, and **emit the
canonical block as text (with `awaiting approval on`) immediately before calling `AskUserQuestion`** — the
block is the summary, the gate is the question. Use AskUserQuestion:
**Approve** / **Request changes** (say what, iterate in chat, re-gate) / **Look at artifacts** (the human
edited the file or left `TODO:` notes — `diff` against the backup, fold in their edits and TODOs, then
re-gate) / **Stop** (leave as is and exit). Only the human sets `approved`; never approve on the human's
behalf. On approval, set the approved sections to `approved` and refresh `<spec>.backup`.

### Handling a change (chat request, edited IDEA, or edited artifact / `TODO:`)

Changes arrive from chat, from an edited IDEA, or from the human editing an artifact section or leaving a
`TODO:` (detected by diffing against `<spec>.backup`). The classification of what a change touches is
yours to make; the human confirms by approving the delta.

1. **Journal first, then edit.** Before changing any content, set the affected phases' flags to `stale`,
   update `STATUS`, and save the file. This write-ahead record is what makes recovery-after-interruption
   work: an interrupted cascade is visible in the file.
2. **Cascade a minimal delta, never a wholesale regen.** Rework only what the change actually affects
   and preserve everything still valid; a `stale` flag means "recheck and patch", not "throw away".
   A changed IDEA can invalidate everything below it; a chat request may touch only one phase and its
   descendants.
3. Re-run the approval gate for the reworked normative sections.

## Implement

Implement the approved DESIGN. Follow the consuming repository's conventions and its `CLAUDE.md`/
`AGENTS.md`; make the smallest reasonable change; add/update meaningful tests; do NO unrelated
refactoring. Every change must trace to the spec or to necessary support. If a genuine product decision
is missing, add it as a `[NEEDS CLARIFICATION]` in REQUIREMENTS, mark the affected sections `stale`, and
stop for the human — do not invent it.

## DoD (mandatory verification)

Determine the project's real verification commands from its own config (Makefile/Taskfile/justfile,
package.json, pyproject.toml, go.mod, `.github/workflows/*`, README, CLAUDE.md/AGENTS.md) — prefer what
CI or the documented dev workflow uses. Run them (tests, linters, static analysis, build). Record the
exact commands and results in `CODE`.

- Implementation failure → fix, re-verify (bounded ~3 attempts). Set `DoD: pass` only when they pass.
- Environment/tooling failure you must not work around → STOP with a clear failure message; never report
  success for a verification step you could not actually run.

## Review (fresh, independent)

Launch the reviewer as a fresh isolated agent (see Usage). Give it ONLY the spec file path and, if any
paths held pre-existing user changes, that note. Do NOT pass your reasoning — it must judge the
repository state itself. Expect `APPROVED` / `REJECTED` / `BLOCKED` with findings classified by severity
(BLOCKER/MAJOR/MINOR) and type (IMPLEMENTATION_BUG/TEST_GAP/SPEC_GAP/OUT_OF_SCOPE).

- **APPROVED**, or REJECTED with only trivial MINORs → set `verdict: APPROVED`, State `DONE`. (Fix
  trivial MINORs; never loop on pure style.)
- Any **SPEC_GAP** / reviewer **BLOCKED** → record it as a `[NEEDS CLARIFICATION]` in the owning phase,
  mark it `stale`, and STOP for the human. Do not invent the answer.
- **REJECTED** with BLOCKER/MAJOR → fix. If a finding only touches code, patch code and re-review. If it
  touches a normative phase, journal it `stale`, cascade the delta, re-gate, then re-review.

Bounded: at most **5** review iterations. If reached without approval, or the same findings return after
a fix (no progress), STOP and report the remaining findings.

## PR-fix mode (State DONE)

Entered only when everything is `approved`, code `pass`, review `APPROVED`. Gather actionable fixes from
**both** sources (reading only — never commit or push):

- **PR review comments** for the current change, via the repository's connected tooling.
- **`TODO:` markers the human left in the code**, limited to the **uncommitted** working-tree changes
  (use the repository's own diff of uncommitted changes and scan the added lines for `TODO:`). Only the
  ones newly added in the working tree — never act on committed baseline TODOs.

Handle each fix the same way, by classifying its scope:

- **Touches a normative artifact** (requirements/research/design — e.g. a banned dependency, a changed
  contract) → do NOT touch code. Instead write down in the affected section exactly what must change,
  mark it `stale`, set State `PR_FIX`, and wait until the artifacts converge (human updates IDEA /
  approves the delta). Only after they are consistent again do you touch code, then remove a code
  `TODO:` that drove the fix.
- **Pure code** (rename, local fix that does not diverge from DESIGN) → apply it, **remove the handled
  `TODO:` from the code**, then re-review. If a fix you are about to apply contradicts the current
  DESIGN, flag it and ask rather than silently diverging.

## Invariants (do not violate)

- The spec file is the source of truth for state; persist every consequence to it immediately.
- The human owns IDEA and approvals; you own the final derived content and the change classification.
  The human may edit artifacts or leave `TODO:` notes; reconcile them via the `<spec>.backup` diff.
- Never invent a product decision — surface it as `[NEEDS CLARIFICATION]` and stop.
- Cascade minimal deltas; never regenerate an approved chain wholesale.
- Verification is mandatory; review is a fresh independent agent.
- **VCS is read-only.** Never commit, push, create branches, or run destructive VCS commands, and never
  discard the human's pre-existing uncommitted changes. Operate on the current working tree.
- Bounded: ≤5 review iterations, ~3 DoD fix attempts.

## Output

Chat is a pointer to the file, not a copy of it. The specifics live in the spec; never restate
requirement/research/design content in chat — no acceptance-criteria dumps, no tech specifics (language
versions, library names, contracts).

**Every turn ends with the same canonical block** — the markdown structure below, sections in this fixed
order. Always include `project` (with the canonical absolute spec path in its header), `status` and
`next step`; include the rest only when it applies. The path in the `## project` header is what a later
bare `/mini-sdd` uses to recover the spec. It is the last thing you write before yielding control. If the turn ends by calling a tool — in particular the
approval gate `AskUserQuestion` — **emit the block as text FIRST, then make the call**; never let a tool
call swallow the block. Any working narration goes before it, kept terse.

## project <what this project is, one line, no more than 5 words> (<canonical absolute path to the spec file>)

### status
<the STATUS state>
### next step
<one line, or "—">

### changes
- <short delta, one item per change — "added endpoint X"; omit on first generation / no change>

### blockers
- <open [NEEDS CLARIFICATION] items needing your decision; omit if none>

### awaiting approval on
<sections at the gate, e.g. requirements, research, design; omit when not gating>

### verification
<exact commands + pass/fail; only right after DoD>

### review
<verdict + iteration; only right after a review>

### result
<mini-sdd complete | mini-sdd blocked | mini-sdd failed; only on the terminal turn>

The approval gate (AskUserQuestion) still fires as its own interactive step; `awaiting approval on`
names what it covers. `blockers` and the gate are the two things that must always surface, because they
need the human.
