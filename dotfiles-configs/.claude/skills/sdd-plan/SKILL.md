---
name: sdd-plan
description: "Planning stage for the Spec-Driven Development (SDD) workflow. Read-only. Builds an implementation plan from an already-validated specification, mapping every specification requirement to concrete implementation and verification work. Returns READY with a traceable plan, or BLOCKED if repository inspection reveals the spec is actually insufficient. Used by 'dotfiles sdd <spec-file>' and invokable manually as /sdd-plan."
---

# /sdd-plan

Build an implementation plan from the approved specification.

## Rules

- Inspect the repository to ground the plan in real files and conventions.
- Do NOT modify any file. This stage is read-only.
- Avoid introducing work that cannot be traced back to the specification or necessary
  implementation support.

## Requirements

- Enumerate every specification requirement.
- Map each requirement to concrete implementation work (files to change) and verification work
  (tests / checks). Every requirement must map to at least one plan item.
- Make traceability explicit, e.g.:

  ```
  SPEC-1: duplicate request_id returns the same operation
    -> modify internal/service/create.go (idempotency lookup)
    -> add TestCreate_Idempotent
  ```

## Result

Return READY with the full traceable plan and the requirement count. If repository inspection
reveals the specification is actually insufficient, return BLOCKED with the gaps instead of
guessing. Return only the structured result.
