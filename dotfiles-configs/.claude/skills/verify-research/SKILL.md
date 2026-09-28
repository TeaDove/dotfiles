---
name: verify-research
description: "Independently verify the conclusions of the last topic discussed in the current chat (or a topic named by a hint, or another chat given by session id / transcript path): typically an on-call incident investigation — is the root cause right, is the evidence solid, will the fix/runbook actually work. Launches a fresh, isolated, read-only research-verifier agent that reads the chat transcript itself and re-checks the claims against real sources, then returns a review-style report of problems and how to fix them. Never edits anything and never runs mutating commands. Trigger phrases: '/verify-research', 'проверь выводы в чате', 'перепроверь расследование', 'проверь, что мы правильно нашли причину', 'проверь инструкцию починки', 'verify research'."
---

# /verify-research

## Usage

```
/verify-research                                        # the last topic of the current chat
/verify-research <topic hint>                           # a specific topic of the current chat
/verify-research <session-id | path.jsonl> [topic hint] # another chat
```

By default only the **last topic** of the chat is verified, not the whole chat. A topic hint is free
text naming the discussion to verify instead (e.g. `про падение воркера`).

## Flow

1. **Find the transcript.**
   - If the first argument is a path to a `.jsonl` file, use it as is; if it is a session id, it
     resolves to `~/.claude/projects/*/<session-id>.jsonl`. The rest of the arguments is the topic hint.
   - Otherwise the current chat is `~/.claude/projects/*/${CLAUDE_SESSION_ID}.jsonl`, and all
     arguments, if any, are the topic hint.
   - If the file does not exist or the session id is not known, ask the user for the session id or path
     and stop. Never guess by picking the newest transcript — parallel chats live in the same directory.
2. **Launch the verifier** as a fresh isolated agent: the `Agent` tool with
   `subagent_type: research-verifier`. Pass ONLY the absolute transcript path, the working directory,
   and the user's topic hint verbatim if there is one. Pass nothing else: no summary of the findings,
   no hints on what to check, no opinion on whether the conclusions are right. The verifier must judge
   the chat on its own; your reasoning would defeat the independent check.
3. **Relay the report** as the verifier returned it. Do not soften, reorder, or drop findings, and do
   not act on them: no fixes, no runbook steps. The user decides what to do next.

## Hard constraints

- Read-only end to end: neither you nor the verifier edits files, commits, or runs mutating commands.
- The verifier runs in a fresh context; never let it inherit this chat's reasoning other than via the
  transcript it reads itself.
