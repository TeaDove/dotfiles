---
name: sdd
description: "Spec-Driven Development (SDD) orchestrator, implemented purely as a Claude skill (no external tool). Drives a full workflow from a specification file: strictly validate the spec, plan, implement, verify with the project's own tests/linters, then review in a FRESH independent subagent, and fix-loop until approved or blocked. The spec is the source of truth; missing product decisions stop the run with BLOCKED instead of being invented. Trigger phrases: '/sdd', 'запусти sdd', 'sdd <spec-file>', 'spec-driven development', 'реализуй по спеке через sdd'."
---

# /sdd — Spec-Driven Development

## Usage

```
/sdd <path-to-spec-file>
```

You (Claude) are the orchestrator. Drive the whole workflow yourself, in this context, using your
normal tools. The ONLY step that must run in a separate, fresh context is the independent review —
launch it with the `Task` tool (`subagent_type: sdd-reviewer`).

If no spec path is given, ask for one and stop.

## Invariants (do not violate)

- **The specification is the source of truth.** Do not begin implementation until validation passes.
- **Never invent product/behavioral decisions.** If information needed to implement the requested
  externally observable behavior is missing, contradictory, or ambiguous, STOP with `BLOCKED` and ask
  the user to update the spec — do not resolve it with a "reasonable assumption".
- **Every requirement must be implemented; every change must be justified** by the spec or by necessary
  implementation support. No opportunistic unrelated refactoring or cleanup.
- **Verification is mandatory.** Passing review is not enough — the project's tests/linters/build must
  pass.
- **Review must be independent** — a fresh subagent that never sees your implementation reasoning.
- **Safety.** Never commit, push, create branches, or run destructive VCS commands. Never discard the
  user's pre-existing uncommitted changes. Operate on the current working tree.
- **Bounded.** At most **5** independent review iterations. Verification/fix retries are also bounded
  (use ~3). The hard iteration limit is the primary safety mechanism.

## Workflow

```
VALIDATE ─BLOCKED─▶ STOP
   │READY
PLAN ─────BLOCKED─▶ STOP
   │READY
IMPLEMENT ─BLOCKED▶ STOP
   │
VERIFY ◀──────────┐
   │fail          │FIX
   ├──────────────┘
   │pass
FRESH REVIEW ─APPROVED─▶ DONE
   │         ─SPEC_GAP─▶ STOP (BLOCKED)
   │REJECTED
FIX ─▶ VERIFY ─▶ FRESH REVIEW   (≤ 5 reviews)
```

### 1. Validate (read-only, strict)

You are the last line of defense before code is written — be rigorous and skeptical. Decide whether
the spec pins down the requested **externally observable behavior** precisely enough that any competent
engineer would produce the **same observable results**, without inventing product requirements. You may
read the repo to judge sufficiency; do not modify anything.

Check every requirement individually:

1. Enumerate each requirement / command / endpoint / output the spec describes.
2. For each, verify the observable contract is fully determined:
   - Every described response/output MUST have a concrete shape — exact fields, their types, and their
     meaning. A response described only in prose ("returns the system load", "returns status") with no
     field list is **UNDERSPECIFIED → BLOCKED**.
   - When a described field maps to data that can legitimately have several values (e.g. an entity with
     multiple addresses mapped to a single field), the spec MUST say which value or how to pick.
   - Failure/edge-case semantics that change what the caller observes (error status, error body,
     empty/absent data, conflicts) must be defined where they matter.
3. Ask: could two competent engineers each satisfy the literal text yet produce **observably different**
   results? If yes → BLOCKED.
4. If requirement B depends on the output of an underspecified requirement A, B is underspecified too.

Not reasons to block (safely the implementer's choice): internal implementation details that do not
change observable behavior (concurrency primitive, file/module layout, algorithm, libraries, logging),
and stylistic latitude the spec explicitly grants ("minimalistic", "responsive"). Do not fill any gap
with a reasonable assumption; if you must invent a field, schema, status code, or selection rule the
spec does not state, that itself is a BLOCKED item.

If **BLOCKED**: print the concrete missing/ambiguous items (name the endpoint/requirement and the
missing decision), tell the user to update the spec and rerun, and STOP. Change no code.

### 2. Plan (read-only)

Enumerate every specification requirement and map each to concrete implementation work (files to
change) and verification work (tests to add / checks to run). Every requirement → at least one plan
item; every planned change must trace back to the spec or necessary support. If repository inspection
reveals the spec is actually insufficient, STOP with `BLOCKED`.

### 3. Implement

Implement the whole plan. Follow existing repository conventions, make the smallest reasonable change,
add/update meaningful tests, preserve backwards compatibility unless the spec says otherwise. Do NOT do
unrelated refactoring. If a genuine product decision is missing, STOP with `BLOCKED`.

### 4. Verify (mandatory)

Determine the project's real verification commands from its own config (Makefile, Taskfile, justfile,
package.json, pyproject.toml, go.mod, `.github/workflows/*`, README, CLAUDE.md) — prefer what CI or the
documented dev workflow uses. Run them (tests, linters, static analysis, build, race detector where
appropriate). Distinguish:

- a failure caused by the implementation → go to **Fix**, then verify again (bounded ~3 attempts);
- an environment/tooling/missing-dependency failure you must not work around → STOP with a clear
  BLOCKED/failure message.

Do not report success if a required verification step could not actually be performed.

### 5. Independent review (FRESH context — required)

Launch the `Task` tool with `subagent_type: sdd-reviewer`. Give it ONLY: the spec file path, the repo,
and an instruction to review the current working-tree changes (diff) against the spec. Do **not** pass
your implementation reasoning, plan, or a self-summary — the reviewer must judge the repository state
itself. Expect a structured verdict: `APPROVED` / `REJECTED` / `BLOCKED`, with findings classified by
severity (BLOCKER / MAJOR / MINOR) and type (IMPLEMENTATION_BUG / TEST_GAP / SPEC_GAP / OUT_OF_SCOPE).

Decide:

- **APPROVED**, or REJECTED with only MINOR findings and no spec violation → **DONE**. (Fix clear,
  low-risk MINORs if trivial, but never loop on pure style.)
- Any **SPEC_GAP** (or reviewer `BLOCKED`) → STOP with `BLOCKED`; do not let the implementer invent an
  answer. Ask the user to update the spec.
- **REJECTED** with BLOCKER/MAJOR findings → go to **Fix**.

### 6. Fix loop

Address the review's actionable findings. The spec stays authoritative: do not reinterpret it, do not
make unrelated changes, and do not blindly obey a reviewer suggestion that conflicts with the spec
(report the conflict instead). If a finding reveals a real spec gap → STOP with `BLOCKED`. After fixing,
**always** go back to **Verify**, then run a **new fresh review** (increment the counter).

Stop conditions:

- Approved → success.
- 5 reviews reached without approval → STOP: report "review limit reached" and the remaining findings.
- No progress (the same findings return after a fix) → STOP and say so; do not loop forever.

## Output

Keep progress readable: show the current phase, the commands you run for verification, and each review
iteration's verdict. End with one of: `SDD complete`, `SDD blocked` (+ what the spec must clarify), or
`SDD failed` (+ remaining findings). Use a clear non-success ending when not approved.
