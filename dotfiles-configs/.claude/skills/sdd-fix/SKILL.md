---
name: sdd-fix
description: "Fix stage for the Spec-Driven Development (SDD) workflow. Fixes supplied verification failures or independent review findings without reinterpreting the specification or making unrelated changes. Rejects reviewer suggestions that conflict with the spec, and returns BLOCKED if a finding reveals missing specification information. Used by 'dotfiles sdd <spec-file>' and invokable manually as /sdd-fix."
---

# /sdd-fix

Fix the supplied verification failures or independent review findings.

## Rules

- The specification remains authoritative.
- Address every finding.
- Do NOT reinterpret the specification.
- Do NOT introduce unrelated changes.
- Do not blindly follow reviewer suggestions that conflict with the specification; if a reviewer
  request conflicts with the spec, report the conflict.

## Result

If a finding reveals missing specification information (a real spec gap), return BLOCKED with the
open questions instead of inventing behavior. Otherwise return DONE. Return only the structured
result once the fixes are done.
