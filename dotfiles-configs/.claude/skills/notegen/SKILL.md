---
name: notegen
description: "Use when the user asks to generate, create, refresh, or update a NOTES.md file that explains a directory/module — its purpose, architecture (presentation vs business logic), business value, key files, and external integrations. Trigger phrases: 'сделай NOTES.md', 'задокументируй эту папку', 'что происходит в этой папке', 'объясни модуль', 'актуализируй документацию', 'обнови NOTES.md'."
---

# /notegen

## Usage

```
/notegen <path>                        # create or update NOTES.md in <path>
/notegen <path> <related-path>         # also account for a second directory that holds this module's core logic (e.g. cmd/foo whose logic lives in internal/foo) — skips auto-detection for it
/notegen <path> --tree                 # also (re)generate an annotated file/directory tree, depth 2, inside NOTES.md
/notegen <path> --tree=<N>             # same, but tree depth N (like `tree -L N`) instead of the default 2
```

`--tree` and `<related-path>` can be combined in any order, e.g. `/notegen <path> <related-path> --tree=3`.

If no path is given, ask the user which directory to document. Do not default to the current directory silently — a wrong guess in a large repo wastes a full investigation pass.

Use `--tree` for the "I just landed in this project, what even is here" case: it walks the directory and attaches a one-line, grounded explanation to every file and subdirectory, in the style of an annotated `tree -L 2` output. This is significantly more expensive in tool calls/tokens than the base flow — only do it when explicitly requested via `--tree`, never infer it from a plain `/notegen <path>`.

## Goal

Produce a `NOTES.md` inside the target directory that lets a human or an LLM understand the module in one read instead of re-deriving it from scratch every time (reading it should replace ~15-20 exploratory tool calls with 1-2). Optimize for the doc being **trustworthy**: every claim must be traceable to something actually read (code, comments, config, existing README/tests) — not inferred from the directory's name or generic priors about what a module "probably" does. When something is a genuine inference rather than a directly observed fact, say so explicitly instead of stating it as settled.

## Hard constraints

- **Output language: Russian.** All prose/headings in `NOTES.md` must be written in Russian. File paths, identifiers, commands, and inline code stay unchanged (usually English).
- **If `NOTES.md` already exists in the target directory, this is an update, not a regeneration:**
  - Read the existing file first, in full.
  - Keep its existing section headings, order, wording style, and any custom/manual notes exactly as-is, unless they are now factually wrong.
  - Only refresh the factual content under each section to match the current state of the code.
  - Do not replace it with "your" preferred template just because it differs from the default one below — minimal diff applies to docs too.
  - If something is now stale (renamed/removed file, outdated claim), correct it in place, in the same section it already lived in.
  - If the existing file is missing something from the checklist below, append a new section at the end rather than reshuffling what's already there.
- **If `NOTES.md` does not exist**, create it fresh using the template in "Default section set" below.
- **Respect scope.** Do not attempt to read the whole repository. Investigation budget: the target directory itself, plus at most one hop into a directly-imported internal package that clearly holds the "real" business logic for this module (see Step 3). If you find yourself wanting to go deeper than that, stop and note the gap in the doc ("детали реализации X не разбирались, см. <path>") instead of silently expanding scope.

## Steps

### 0. Resolve target
Get the target directory from the invocation argument. If a second argument is given, treat it as an explicitly known "logic" counterpart directory — include it, skip the auto-detection heuristic for it in step 3.

### 1. Read what already exists
- Read `NOTES.md` if present — this determines create vs update mode.
- Read any `README.md` / other `*.md` files already in the target directory. Treat these as primary sources: summarize/reconcile with them rather than re-deriving the same facts independently from raw code.

### 2. Get structural facts without bulk-reading everything
- Use `ls`, `grep`/`rg`, and Read — but prefer file headers/outlines over full reads for anything long. For a file over ~500 lines, skim its top-level declarations first (e.g. `grep` for `func`/`class`/`type`/`def` signatures) and read only the relevant slice, rather than reading it in full.
- List the files and group them by rough role: entrypoint/CLI vs internal logic vs config/example data vs tests.

### 3. Identify the architecture split
- Determine whether this directory is a thin presentation/orchestration layer over logic that actually lives elsewhere (common pattern: `cmd/X` wraps `internal/X` or a sibling package). Find this from the imports of the entrypoint file(s) — don't guess from naming alone.
- If such a counterpart exists and wasn't given as an explicit second argument, take one hop into it: read enough of its key files (exported types/funcs) to describe what it actually does. Don't fully document it — that belongs in its own `NOTES.md` if/when someone runs `/notegen` on it.

### 4. Business value / purpose
- Ground this in: existing docs, code comments, config field names/comments, test names, and observed behavior — not in the directory name alone.
- If you can't ground the "why" in anything concrete, say so explicitly rather than inventing a plausible-sounding justification.

### 5. External integrations
- Identify real external dependencies from imports/client construction: datastores, queues, other internal services (gRPC/HTTP clients), object storage, etc.
- For each, name the specific file/package responsible — not just "uses X" with no pointer.

### 6. How to run/use it (if applicable)
- Only include this section if the directory is runnable/invokable (has a `main.go`, a CLI, or exported entrypoints used elsewhere). Pull actual commands from the existing README/tests rather than inventing flags.

### 7. Optional: annotated file tree (`--tree`)

Only do this if `--tree`/`--tree=<N>` was passed. Skip this step entirely otherwise — it's the expensive path.

- Get the literal structure first: `tree -L <N>` (default `N=2`) on the target directory. If `tree` isn't installed, reconstruct the same shape with `find <path> -maxdepth <N>` or `ls -R` and format it as a tree yourself.
- For every file and directory entry, append a short one-line explanation, grounded in something actually looked at — not guessed from the name:
  - **Files**: a quick skim is enough — package/file doc comment, top-level signatures, or the first ~30-50 lines otherwise. Reuse whatever you already read in steps 1-6 instead of re-reading those files.
  - **Directories** (at the deepest listed level): one line summarizing what's inside, based on its file names/roles, not a full recursive read of everything under it.
- **Budget/collapsing rules** (this is what keeps `--tree` affordable):
  - If a directory contains many files that are clearly one pattern (e.g. 15 `*_test.go` files, a directory of dated run folders, generated/example data), collapse them into one annotated line covering the pattern (`*_test.go (12 файлов) - unit-тесты по одному на файл сценария`) instead of explaining each individually. Call out only genuine outliers by name.
  - Do not open binary, generated, or data files (`.json.gz`, `.html` reports, snapshots, `db-structure.json.gz`-style dumps) — label them by what they evidently are from name/size/extension and move on.
  - If depth `N` would expose more entries than is reasonably explainable (rule of thumb: >~60 annotated lines), stop descending further and note it (`... (глубже не разбиралось, см. также при необходимости)`), rather than silently truncating without saying so.
- Output format — literal tree characters (`├──`, `└──`, `│`) with ` - <объяснение>` appended per line, e.g.:
  ```
  .
  ├── README.md - документация по запуску проекта
  ├── go.mod - зависимости
  ├── pkg/ - основная часть исходников
  │   ├── cli/ - исходники CLI
  ```
- Put this as its own `## Дерево файлов` section (see template below). On update, this section follows the same override rule as everything else: only touch it if `--tree` was passed this time; otherwise leave an existing `## Дерево файлов` section untouched even if the rest of the doc gets refreshed.

### 8. Write the file

Default section set — use these headings when creating fresh; when updating an existing file, keep its existing headings if they differ from this list:

```markdown
# NOTES.md

_Обновлено: <YYYY-MM-DD>. Составлено/актуализировано ассистентом (skill notegen) на основе чтения кода — часть формулировок может быть неполной или требовать проверки перед использованием в критичных решениях._

## Назначение
...

## Архитектура и ключевые файлы
...(презентационный слой vs бизнес-логика, ключевые файлы с однострочным описанием)

## Как запускать / использовать
...(только если применимо)

## Внешние интеграции
...

## Дерево файлов
...(только если запускалось с --tree; см. Step 7 — иначе секцию не добавлять)

## Открытые вопросы / что не разбиралось
...(явно перечисленные пробелы, вытекающие из ограничения по scope выше)
```

### 9. Report back to the user (in chat, not in the file)
State: create vs update mode, which sections changed, and — most importantly — which specific claims are inferred/low-confidence and worth the user double-checking themselves.
