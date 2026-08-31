---
name: sdd-implement
description: "Implementation stage for the Spec-Driven Development (SDD) workflow. Implements an approved specification according to its plan, making the smallest reasonable change that follows existing repository conventions, and adds meaningful tests. Stops with BLOCKED if a product decision is missing rather than inventing behavior. Used by 'dotfiles sdd <spec-file>' and invokable manually as /sdd-implement."
---

# /sdd-implement

Implement the approved specification according to the implementation plan.

## Rules

- The specification is authoritative.
- Follow existing repository conventions.
- Make the smallest reasonable change consistent with the repository's architecture.
- Do NOT perform unrelated refactoring or opportunistic cleanup.
- Add or update meaningful tests where appropriate.
- Preserve backwards compatibility unless the specification says otherwise.
- Do not silently change the specification.
- Implement the COMPLETE plan — every requirement must be covered.

## Result

If implementation requires a product or behavioral decision absent from the specification, stop
and return BLOCKED with the open questions rather than inventing behavior. Otherwise return DONE
with a short summary of what changed. Return only the structured result once the code is done.
