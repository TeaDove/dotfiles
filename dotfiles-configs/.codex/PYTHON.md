# Python rules

These rules apply to all Python code, on top of the general rules in [AGENTS.md](AGENTS.md).
I write Python only as standalone scripts: single-file CLI tools for duty work and one-off automation.

## Preferred stack

- In existing code, use what it already uses. Do not migrate it to the stack below unless I ask.
- Shebang `#!/usr/bin/env python3` (Python 3.14); dependencies come from that interpreter, no venv or `uv` project.
- CLI: `typer`.
- Logging: `loguru`.
- JSON: the standard `json` module into plain dicts and lists, no pydantic.
- HTTP client: `httpx`.
- Linting and formatting: `ruff`.

## Definition of done

After every code change, before reporting the task as complete, you **must**:

1. Run the code — confirm it executes. Stub external commands that change anything (prod CLIs, ssh/pssh) and run
   the script against the stubs.
2. Run the tests, if there are any (e.g. `pytest`) — confirm they pass.
3. Run `ruff check --config ~/dotfiles/ruff.toml` and `ruff format --config ~/dotfiles/ruff.toml` (or the code's own
   ruff config, if it has one) — confirm they pass.

Do not skip these steps, and don't rely on CI to catch what you missed.

## Code style

- Keep it simple: one file, plain functions, no layers or abstractions beyond what the task needs.
- All configuration (constants, paths, templates, etc.) goes at the top of the file, right after the imports.
- Always annotate types.
- CLI with `typer`: one `typer.Typer()` app, one `@app.command()` per subcommand (add an empty `@app.callback()` so a
  single command still runs as a subcommand). Declare parameters in the default-value style, not `Annotated`:
  `ticket: str = typer.Option(..., "-t", "--ticket")`, `label: list[str] | None = typer.Option(None, "--label")`.
  The first argument is the default (`...` means required).
- Log every external command with `loguru` before running it.
- Errors: raise plain `Exception` with a clear message, no custom exception classes; catch it once in the command,
  log it and exit with code 1.
  Log with `logger.exception` so the traceback is kept:

  ```python
  try:
      ...
  except Exception:
      logger.exception("failed to ...")
      raise typer.Exit(1)
  ```
- Multi-line text: adjacent f-strings inside parentheses with explicit `\n`, not `"\n".join([...])`.
- `@dataclass` only for the few structured values passed between functions.
- In `.ipynb` notebooks: keep all imports in a single cell at the very top of the file; put all settings
  (e.g. chart colors) in the second cell.
