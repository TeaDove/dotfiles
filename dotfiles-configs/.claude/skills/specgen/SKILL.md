---
name: specgen
description: "Use after finishing code on a feature branch, before running a review, to reverse-engineer a concise ticket-level spec (\"ТЗ\") describing the problem and goal the branch addresses, relative to its base branch (master/main) — written like a Jira ticket or Confluence doc, not a file-by-file changelog. Trigger phrases: 'собери ТЗ по бранчу', 'опиши что сделано в этой ветке', 'сгенерируй ТЗ по коду', 'summarize this branch as a spec', 'what did this branch actually do'. Produces a short recap (goal(s), problem/cause if knowable, what was solved, out-of-scope/unrelated changes, open questions) that the human edits and then hands to a review agent (/code-review, /review) as the reference spec."
---

# /specgen

## Usage

```
/specgen                    # review only what this branch adds over its base branch (master, fallback main)
/specgen <base-branch>      # compare against an explicit base branch/ref instead
/specgen --save[=<path>]    # also write the recap to a markdown file (default: SPEC.md in repo root)
```

If the repo has no `master` or `main` branch and none was given explicitly, ask the user which base ref to diff against — do not guess.

## Goal

The user has just finished writing code on a branch and wants to sanity-check it against the original
task **before** asking anyone (human or agent) to review it. This skill reads the diff and writes a
ticket-level spec of what problem/goal the branch addresses and, at a high level, how — the way a Jira
ticket or a Confluence design doc would state it, e.g. "users occasionally hit a nil pointer dereference
because User isn't always initialized; add explicit initialization and cover it with tests." It never
turns into a changelog: no file paths, no function/test names, no line-by-line inventory of the diff.
Implementation detail belongs in the code and the PR diff itself, not in this recap — the human already
knows exactly which files changed and cares about the *why*, not a redundant list of the *where*.

The human compares this recap against their own original intent to answer three questions:

1. Is everything from the original task actually done?
2. Is there anything in the diff that isn't related to the original task?
3. Is the branch trying to do too much and should it be split into multiple PRs/MRs? (A recap listing
   several unrelated goals is itself the signal here — no size metrics needed to make that case.)

The resulting recap can also be edited by the human and handed to a separate review agent as the
"intended spec," so that review checks the code against stated intent instead of inventing its own.
The user of this skill is fully aware an LLM can't reverse-engineer a perfect spec from code alone — the
recap is a draft they will edit themselves, not a claim of ground truth.

## Hard constraints

- **Read-only.** Never edit files, stage, commit, or push. This is a reporting skill, not an implementation one.
- **VCS-agnostic — never assume git.** Read history and diffs through whatever version control the repo
  uses; big monorepos ship their own VCS and their own LLM instructions for it (for example `arc`).
  Think in terms of actions ("the history this branch adds", "the diff against the base"), and let the
  repo's own tooling/instructions supply the exact commands.
- **Ticket-level abstraction, not a changelog.** The recap describes the problem/goal and, in general terms,
  what kind of solution was applied — never "with the precision of files": no file paths, no
  package/function/test names, no line-level inventory. If a section would otherwise read like a
  per-file change summary with commentary, it's too detailed — rewrite it one level up (what capability/behavior changed,
  not where).
- **Ground everything in the diff**, even though the wording stays abstract. Read enough of the diff and
  surrounding context (via `Read`, not guessing) to correctly state the problem and the shape of the fix.
  If something is genuinely unclear from the code, say so plainly — but don't hedge routine, clearly-stated
  facts with "предположительно"/"похоже, что"/"судя по всему". State what the diff shows directly; reserve
  uncertainty language for things that are actually uncertain.
- **Do not fetch or fabricate the "original" task.** This skill only reconstructs a spec from code; it does
  not know what was actually asked for. Never claim the recap matches or doesn't match some assumed
  requirement — that judgment belongs to the human.
- **Write in an imperative/infinitive register**, like a ticket describing work to be done, not a past-tense
  narration of a diff — e.g. "исправить баг, из-за которого..." / "добавить явную инициализацию и покрыть
  тестами", not "в диффе исправлен баг..." / "добавлена инициализация в файле X".
- **Output language:** Russian, since this is a working recap for the user, not a committed doc artifact.
- **Keep it short.** A human should be able to read the whole recap in under a minute for a normal-sized
  branch. Prefer grouped bullets over prose paragraphs.

## Steps

### 1. Resolve the base ref
- Default to `master`; if it doesn't exist, fall back to `main`; if neither exists, ask the user.
- If the user passed an explicit base, use that instead without asking.
- Diff from the point where this branch diverged from the base, not from the base branch tip, so the
  recap only covers what this branch actually added. Use the repository's own version-control tooling to
  find that divergence point and produce the diff — do not assume git.

### 2. Get the shape of the change first
- Get the commit history this branch adds over the base — a one-line-per-commit list.
- Get a per-file change summary (files touched), purely to budget how you read the diff in step 3.
  This is internal bookkeeping only — none of these numbers belong in the final recap.

### 3. Read the diff
- Get the full diff of what this branch adds over the base.
- For a large diff (rule of thumb: over ~1500 changed lines or ~30 files), don't read it as one blob —
  go file-by-file or directory-by-directory, and summarize each group before moving to the next, to avoid
  losing the smaller changes among the large ones.
- Where a diff hunk is not self-explanatory (e.g. a changed condition, a new field with no context), open
  the file with `Read` to understand the surrounding function/struct rather than guessing from the `+`/`-`
  lines alone.
- Pure renames/moves with no content change (the VCS status usually flags these) should be collapsed
  into a single mention, not itemized file by file.

### 4. Group changes by independent goal, not by file order
Look for the natural problem/goal boundaries in the diff — e.g. "a bug where X" / "add capability Y" /
"restructure Z" — and mentally bucket every changed file under one of them, rather than walking the diff
top to bottom. This grouping is what step 5 turns into numbered goals; a diff with N clearly independent
buckets is itself the evidence for "this should probably be N branches," without needing to say so
explicitly.

### 5. Write the recap

Produce a Russian-language report, written the way a Jira ticket or Confluence spec would state the work
*before* it existed — problem/goal and shape of the solution, never file paths, package/function/test
names, or a diff inventory. Sections:

- **Цель** — one or more independent goals the branch addresses (number them if there's more than one,
  matching the buckets from step 4). State each directly, in imperative/infinitive form: "исправить баг,
  из-за которого...", "реализовать команду для...", "провести рефакторинг структуры...". No hedging
  qualifiers, no "(предположительно)".
- **Причина / как воспроизвести** — only for goals that are bugfixes or have a knowable root cause: describe
  the underlying problem and, if derivable from the diff (e.g. from the condition being fixed), how it
  manifests/reproduces. Skip this section for goals where there's no bug or cause to state (new features,
  pure refactors).
- **Что сделано** — for each goal, the shape of the solution in capability/behavior terms: what now happens
  that didn't before, what class of safeguard was added (e.g. "добавлена явная инициализация и тестовое
  покрытие для этого сценария") — still no file/function/test names.
- **Что не затронуто** — only if something adjacent to a stated goal was conspicuously *not* covered (e.g.
  a class of input left untested, a related code path not updated) — described as a behavior/coverage gap,
  not as a named file. Skip the section if nothing stands out.
- **Потенциально не по теме** — changes that don't belong to any of the goals in "Цель": unrelated fixes,
  formatting-only diffs, stray scratch files, changes to unrelated tooling/config. Concrete pointers (file/
  path) are fine *here specifically*, since the point is to help the human locate and decide on them — this
  is the one section where naming things is useful rather than noise.
- **Открытые вопросы** — anything genuinely ambiguous from the diff alone that the user should double-check
  against their own intent. Omit this section if there's nothing worth flagging — don't pad it.

### 6. Save, if requested
Only if `--save` was passed: write the recap to the given path, or `SPEC.md` in the repo root if no path
was given. If the target file already exists, show a short diff of what would change and ask for
confirmation before overwriting — never overwrite silently.

### 7. Close with a pointer
End the chat response with one line pointing at the intended next step, e.g.: "Проверьте ТЗ, при
необходимости поправьте, затем можно передать его ревью-агенту (`/code-review` или `/review`) как
референс того, что должно быть сделано."
