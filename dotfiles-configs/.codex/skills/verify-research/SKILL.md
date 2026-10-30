---
name: verify-research
description: "Independently verify conclusions from the current task, another task, or an exported transcript: typically an incident root cause, supporting evidence, or a proposed fix/runbook. Launches a fresh isolated read-only verifier that re-checks claims against real sources and returns a review-style report. Never edits anything or runs mutating commands. Trigger phrases: '/verify-research', 'проверь выводы в чате', 'перепроверь расследование', 'проверь, что мы правильно нашли причину', 'проверь инструкцию починки', 'verify research'."
---

# /verify-research

## Usage

```
/verify-research                                      # the last topic of the current task
/verify-research <topic hint>                         # a specific topic of the current task
/verify-research <task-id | transcript-path> [hint]   # another task or exported transcript
/verify-research --loop[=N] ...                       # revise and re-verify, at most N rounds (default 3)
```

`--loop` may be combined with any form above and is not part of the topic hint. Its value defaults to
`3`; `--loop=0` disables the loop. A value that is not a non-negative integer: ask the user and stop.

By default only the **last topic** of the task is verified, not the whole task. A topic hint is free
text naming the discussion to verify instead (e.g. `про падение воркера`).

## Flow

1. **Resolve the source.**
   - For a transcript path, use that file exactly.
   - For a task ID, use the available task-reading tool to retrieve that task. If the runtime cannot read
     another task, ask the user for an exported transcript and stop.
   - With no explicit source, use the current task history. Do not search local runtime state or guess the
     newest task because parallel tasks may exist.
2. **Launch the verifier.** Read `references/research-verifier.md`, then launch a fresh isolated subagent
   with those instructions. For the current task, let it inherit the task history without adding your own
   summary. For another task or a transcript, pass ONLY the retrieved transcript or absolute path, the
   working directory, and the user's topic hint verbatim. Add no hints or opinion about the conclusions.
   The verifier must judge the discussion on its own; your reasoning would defeat the independent check.
3. **Relay the report** as the verifier returned it. Do not soften, reorder, or drop findings. Without
   `--loop`, do not act on them: no fixes, no runbook steps. The user decides what to do next.
4. **Loop** (only when `--loop` is on and the report has findings; otherwise stop after step 3).
   - Write a new summary of the researched topic in the chat, taking the verifier's findings into
     account: corrected root cause, fixed or dropped claims, updated fix/runbook, and what was left
     unresolved and why. Text only: no file edits, no mutating commands.
   - Launch a new verifier exactly as in step 2, for the same source and topic hint. Add one line to the
     prompt: the discussion ends with revised summary number K. Nothing else about the findings.
   - Relay the new report as in step 3.
   - Repeat while the report has findings and fewer than N revised summaries were written. Stop early
     when a report has no findings. After the last round, state whether findings remain.

## Hard constraints

- Read-only end to end: neither you nor the verifier edits files, commits, or runs mutating commands.
  The only thing you write in the loop is the revised summary, as chat text.
- The verifier runs in a fresh context; never give it the coordinator's analysis or conclusions.
