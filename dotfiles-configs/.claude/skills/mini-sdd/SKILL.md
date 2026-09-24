---
name: mini-sdd
description: "Lightweight single-file Spec-Driven Development for small tasks. Drives ONE spec file through requirements → research → design → code → independent review, recording every phase and its status flag inside that same file so work resumes after any interruption. The human owns only the IDEA section; the agent owns every derived artifact. Editing IDEA (or asking for a change in chat) marks the affected phases stale and cascades a minimal delta, then re-requests approval. Once everything is approved, implemented and review-passed, it switches to PR-fix mode. When work continued in the branch outside the flow, resync mode reads what the branch actually implements, writes a DRIFT report against the spec, and lets the human fold it back in before re-running the flow. Never commits or pushes. Trigger phrases: '/mini-sdd', 'mini-sdd <spec-file>', 'запусти mini-sdd', 'реализуй по спеке через mini-sdd', '/mini-sdd --resync', '/mini-sdd <spec-file> --resync', 'восстанови SDD по ветке'."
---

# /mini-sdd — single-file Spec-Driven Development

## Usage

```
/mini-sdd [<path-to-sdd-spec.md>] [--resync]
```

`--resync` is a flag, not a positional argument: it may come with or without the spec path, in any
position. It forces the **Resync** flow (see below); without it the flow is chosen by state. The path is
resolved exactly as without the flag.

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
State: <DRAFTING | AWAITING_APPROVAL | IMPLEMENTING | DoD | IN_REVIEW | DONE | PR_FIX | RESYNC>
Next: <one line: what you will do next, or what you are waiting for>

## DRIFT          [status: fresh | decided]   <!-- only during a resync -->
<how the branch diverges from the spec, one item per divergence, each with the human's decision>

## REQUIREMENTS   [status: fresh | stale | provisional | approved]
<observable behaviour, acceptance criteria, edge cases, and any [NEEDS CLARIFICATION] open questions>

## RESEARCH       [status: fresh | stale | approved]
<links to all context (wiki / tickets / repo / code) + one short line: constraints / bans / what to reuse>

## DESIGN         [status: fresh | stale | provisional | approved]
<the HOW: components, contracts, key decisions. `provisional` while REQUIREMENTS has an open blocker.>
```

The spec holds no CODE or REVIEW section: the code lives in the repository and verification and review
results are reported in chat. Their progress is tracked only by `State` (`IMPLEMENTING` → `DoD` →
`IN_REVIEW` → `DONE`); while `IN_REVIEW`, `Next` carries the review iteration (e.g. `review 2/5`).

- **`STATUS` is a computed rollup**, placed right after IDEA for a quick glance. Recompute it from the
  per-phase flags on every change; it must never contradict them (no `DONE` while DESIGN is `stale`).
  Any normative section leaving `approved` resets State to the spec phase; a passed approval gate sets
  it to `IMPLEMENTING`.
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

- `--resync` flag, a chat request to resync, or State `RESYNC` → **Resync** flow (takes precedence).
- REQUIREMENTS / RESEARCH / DESIGN missing or not all `approved` → **Forward / cascade** flow.
- REQUIREMENTS + RESEARCH + DESIGN `approved`, State `IMPLEMENTING` or `DoD` → **Implement** / **DoD**.
- State `IN_REVIEW` → **Review**.
- Everything `approved`, State `DONE` → **PR-fix** mode.

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
exact commands and results in chat (`verification` in the canonical block).

- Implementation failure → fix, re-verify (bounded ~3 attempts). Move State to `IN_REVIEW` only when they pass.
- Environment/tooling failure you must not work around → STOP with a clear failure message; never report
  success for a verification step you could not actually run.

## Review (fresh, independent)

Launch the reviewer as a fresh isolated agent (see Usage). Pass it ONLY: the spec file path (the full
spec — every requirement and design — lives there), the repo root / working directory, and, if any paths
held pre-existing user changes, that note. Pass **nothing else**, and keep it short.

In particular do NOT tell it which requirements passed a previous review, where to focus, what the delta
since last time is, or what to confirm — that is your reasoning, and feeding it defeats the independent
review: the reviewer must re-derive every requirement from the spec and judge the whole current state
fresh. Do NOT restate VCS/tooling or code conventions either; those live in the repo's own
AGENTS.md/CLAUDE.md, which the reviewer reads itself. Expect `APPROVED` / `REJECTED` / `BLOCKED` with
findings classified by severity (BLOCKER/MAJOR/MINOR) and type
(IMPLEMENTATION_BUG/TEST_GAP/SPEC_GAP/OUT_OF_SCOPE).

- **APPROVED**, or REJECTED with only trivial MINORs → State `DONE`. (Fix
  trivial MINORs; never loop on pure style.)
- Any **SPEC_GAP** / reviewer **BLOCKED** → record it as a `[NEEDS CLARIFICATION]` in the owning phase,
  mark it `stale`, and STOP for the human. Do not invent the answer.
- **REJECTED** with BLOCKER/MAJOR → fix. If a finding only touches code, patch code and re-review. If it
  touches a normative phase, journal it `stale`, cascade the delta, re-gate, then re-review.

Bounded: at most **5** review iterations. If reached without approval, or the same findings return after
a fix (no progress), STOP and report the remaining findings.

## Resync (the branch drifted from the spec)

For when work went on in the branch outside this flow. For example, after `DONE` the human kept going by
hand or through ad-hoc agent sessions: exploratory changes, experiments, fixes. The code no longer
matches the spec, and the spec has to catch up before the flow can drive the branch again. You never
edit IDEA here either; you *propose* the IDEA delta and the human applies it.

1. **Journal first.** Set State `RESYNC` and Next `drift report`, then save.
2. **Read what the branch really implements.** Get the whole branch relative to its base, both committed
   and uncommitted, through the repo's own VCS. Do not limit it to the working tree or to changes since
   `DONE`. Read the changed code itself.
3. **Write `## DRIFT`** right after `STATUS`. Compare the code against every REQUIREMENTS item and DESIGN
   decision. Write one item per divergence, numbered `D1`, `D2`, …, each with:
   - **kind**: `added` means the code does something the spec does not describe. `removed` means a spec
     item that the code no longer satisfies. `changed` means both describe it, but differently.
   - **evidence**: `file:line` references.
   - **spec delta**: which `R`/design items would be added, reworded, or deleted to match the code.
   - **IDEA proposal**: the exact line(s) to add to or remove from IDEA if the human keeps the change.
     Leave it empty for design-only drift that changes no observable behaviour.
   - **decision**: `keep` (the spec adopts the code), `revert` (the code goes back to the spec), or
     `[NEEDS CLARIFICATION]`. You always write `[NEEDS CLARIFICATION]`; only the human replaces it.

   Mark debug scaffolding, experiment knobs, and one-off scripts as such in their item, so the human
   can pick `revert` for them. Leave REQUIREMENTS / RESEARCH / DESIGN and the code untouched.
   Set DRIFT `fresh`.
4. **Gate.** Refresh `<spec>.backup` and emit the canonical block. `changes` gets the drift counts per
   kind, and `awaiting approval on` is `drift`. Then use AskUserQuestion with these options:
   - **Apply decisions**: the human has set every `decision` in the file and edited IDEA where they want.
   - **Request changes**: the drift report itself is wrong or incomplete; iterate on it, then re-gate.
   - **Stop**.
5. **Apply.** If any `decision` is still `[NEEDS CLARIFICATION]`, list those items as blockers and stop.
   Do not guess. Then diff against `<spec>.backup` as usual and fold in the human's edits. After that:
   - A `keep` item that changes observable behaviour must be covered by the new IDEA. If it is not, it
     is a blocker: ask the human to update IDEA or switch the item to `revert`.
   - Journal the affected phases `stale`. Cascade a minimal delta so that `keep` items become part of
     REQUIREMENTS / RESEARCH / DESIGN. `revert` items leave the spec as is.
   - Set DRIFT `decided`. Then run the normal
     approval gate.
6. **Converge.** After approval, continue with **Implement**. Bring every `revert` item back to the spec,
   and do not rewrite `keep` code that already satisfies it. Then run DoD and the fresh review as usual.
   On `DONE`, delete the `DRIFT` section and refresh `<spec>.backup`.

## PR-fix mode (State DONE)

Entered only when everything is `approved` and State is `DONE` (DoD passed, review `APPROVED`). Gather actionable fixes from
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
- Resync never edits IDEA and never decides a DRIFT item; it proposes, the human decides.
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
