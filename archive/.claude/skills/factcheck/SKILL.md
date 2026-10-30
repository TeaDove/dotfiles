---
name: factcheck
description: "Reliable, slow, token-heavy fact-check of a claim or question found online (e.g. 'X% of fathers raise children who are not theirs'). Decomposes the claim and agrees a Definition of Done with the user, then runs several Opus researchers with different mandates (primary source, counter-evidence, origin of the number, methodology and definitions), has Sonnet verifiers check every cited quote against its source, loops on their critique, adjudicates only the sub-claims the researchers disagree on, and writes a report that starts with the cost (agents, iterations, disputes, time, tokens) and then answers the question. Trigger phrases: '/factcheck', 'фактчек', 'проверь факт', 'проверь, правда ли', 'fact-check this'."
---

# /factcheck — reliable fact-check with independent researchers

## Usage

```
/factcheck <claim, question, or link to the post that makes the claim>
```

Speed and tokens are not constraints here; correctness is. You are the coordinator: you agree the
Definition of Done (DoD) with the user, launch every agent, keep the run state and write the final
report. Researchers, verifiers and adjudicators are separate agents launched with the Agent tool
(`subagent_type: general-purpose`): researchers and adjudicators with `model: opus`, verifiers with
`model: sonnet`.

Limits (raise them only if the user asks):

| Setting | Default |
|---|---|
| Researchers | 4, one per mandate in step 2 |
| Verify-and-revise rounds per researcher | 2 |
| Adjudication rounds for disputed sub-claims | 2 |

## Run directory and resume

Keep every artifact in `~/.claude/factcheck/<YYYY-MM-DD>-<slug>/`, where the slug is a few words of the
claim. Save each agent's raw output there as soon as it arrives, plus `state.md` with:

- the approved DoD and which steps are done;
- an `## Agents` list with one line per launched agent: role, model, round and its agent ID. Write the line
  when you launch the agent and never edit it afterwards. SendMessage needs the ID for a revision and step
  5 needs it to count tokens, so put progress notes (status, verdict, problems found) on separate lines,
  never in place of the ID line.

A rate limit or an interruption must not lose finished work. On start, if a run directory for the same
claim exists and its `state.md` is not marked done, resume from the first unfinished step instead of
starting over.

Record the start time (`date -u +%FT%TZ`) in `state.md` before step 1, and the end time when the report
is written.

## 1. Decompose the claim and agree the DoD

Do this yourself before any agent runs: it is the cheapest step, and it decides whether the rest is worth
anything.

- Quote the claim exactly as stated. If the user gave a link, fetch it and quote the post.
- Split the claim into atomic sub-claims. For each, pin down what would have to be measured: population,
  place, time period, the quantity and its denominator. A number means nothing until you know "among
  whom": viral statistics are often true for one population and quoted for another.
- Write the DoD. For every sub-claim: a verdict on the scale below, the best estimate with a range where
  it is numeric, the primary sources behind it, a confidence level, and what evidence would change the
  verdict.

Verdict scale: `true`, `misleading` (true only under conditions the claim leaves out), `false`,
`unverifiable`.

Show the sub-claims and the DoD to the user with AskUserQuestion: Approve / Change (say what) / Stop. Do
not launch any agent until they approve. On Change, rework and ask again.

## 2. Research

Launch all researchers at once so they run in parallel. Give each the exact claim, the approved DoD, the
run directory and its own mandate:

1. **Primary source.** Find the original studies, datasets or official statistics behind each sub-claim,
   and report what they actually measured and found.
2. **Counter-evidence.** Make the strongest case that the claim is false or overstated: contradicting
   studies, published critiques, failed replications.
3. **Origin trace.** Find where the claim and its number first appeared, and how they changed on the way
   to the version the user saw.
4. **Methodology.** Compare estimates across populations and methods. Look for selection bias, small
   samples, mismatched denominators, and definitions that differ between sources.

Tell every researcher:

- Use WebSearch and WebFetch; load them with ToolSearch first.
- Every factual statement needs a source URL and a verbatim quote from that page that supports it. A
  statement without a quote is an opinion and must be labelled as one.
- Mark each source `primary` (the study or dataset itself) or `secondary` (news, blog, or review citing
  it), with its publication date. A secondary source counts only for what it reports, not as proof of the
  underlying finding.
- Return only a JSON block: for each sub-claim, the verdict, estimate and range, confidence, caveats, and
  the list of statements as `{statement, url, quote, kind, date}`.

## 3. Verify and revise

When a researcher's result arrives, launch a verifier for it. Its job is mechanical: it checks the
evidence, not whether it likes the conclusion.

- Open every cited URL and check that the quote is really on the page and says what the statement claims.
  Check verbatim when possible (`curl -sL <url>` and search the text). For pages that need a browser, and
  for PDFs, use WebFetch asking for the exact surrounding sentences, compare them yourself, and mark the
  check `by-summary` instead of `verbatim`.
- Flag quotes that are not on the page, quotes taken out of context, dead links, secondary sources
  presented as primary, statements without a quote, and verdicts that do not follow from the cited quotes.
  A paywalled or unreachable source is `unverifiable`, not false.
- Return `PASS`, or `REVISE` with the list of problems.

On `REVISE`, send the problems to the same researcher with SendMessage, so it revises with its earlier
context, then verify again. Stop after the configured number of rounds. Statements that still fail are
dropped from that researcher's result and listed in the report as rejected.

Some problems are not about a statement: an estimate, range, caveat or summary that says more than the
passing statements support. They cannot be dropped, so keep them as notes for the report. When you write
the report, word those parts only as far as passing statements support them. A statement that passed in
another agent's result counts.

## 4. Compare and adjudicate

Build an agreement table: one row per sub-claim, one column per researcher, each cell holding that
researcher's verdict and estimate. A sub-claim is disputed when the verdicts differ or the estimate ranges
do not overlap.

Do not re-run the whole research for a dispute, and do not average it away: the minority result may be
the correct one. For each disputed sub-claim, launch one adjudicator with only that sub-claim, the
competing positions and their verified sources. It must decide from the evidence which position holds, or
under which definitions each one holds, or declare the dispute unresolved and say what evidence would
settle it. Verify the adjudicator's result exactly as in step 3, with the same round limit: send the
problems to the same adjudicator with SendMessage, then verify again. Adjudicators add unsupported details
and mislabel sources as readily as researchers do, so do not skip this because they reuse statements that
already passed. Repeat for the disputes that remain, up to the configured number of adjudication rounds.
Anything still disputed goes into the report as an open dispute, with both positions.

## 5. Report

Write the report in the language of the user's request and save it as `report.md` in the run directory.
It starts with the cost, then answers the question:

```
## Cost
Time: <start> – <end> (<duration>)
Agents: <n> researchers, <n> verifiers, <n> adjudicators (<n> opus, <n> sonnet)
Iterations: verify-and-revise rounds per researcher <list>; adjudication rounds <n>
Disputes: <raised> raised, <resolved> resolved, <open> open
Tokens: agents <input> in (<cache read> cache read, <cache write> cache write), at least <output> out;
coordinator <input> in, at least <output> out

## Verdict
<one line per sub-claim: verdict, best estimate with range, confidence>

## Why
<what the evidence shows, and where the claim's version comes from if it differs>

## Key sources
<the primary sources the verdict rests on: title, year, URL, and the quote>

## Open disputes and unknowns
<each unresolved dispute with both positions; what no source could establish>

## Rejected statements
<statements dropped because their quote did not check out>
```

Count tokens from the transcripts. Do not use the number in an agent's completion notification: it
tracks the agent's context size, not what it spent, and a resumed agent reports it again. Each agent's
transcript is `subagents/agent-<agentId>.jsonl` in the session's transcript directory
(`~/.claude/projects/<project>/<session-id>/`; find it by searching for one of the IDs in `state.md`).
The coordinator's own transcript is `<session-id>.jsonl` next to that directory; count it from the run's
start time. In each transcript, take the assistant entries and group them by `message.id`. Keep the entry
with the largest `usage.output_tokens` in each group, then sum `input_tokens`,
`cache_creation_input_tokens`, `cache_read_input_tokens` and `output_tokens`. Output counts in transcripts
can come out low, so report output as a lower bound. If an agent's transcript is missing, say how many
agents the total leaves out. Never estimate.

Show the report in chat, then mark `state.md` done.
