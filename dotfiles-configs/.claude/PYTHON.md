# Python rules

These rules apply only to Python projects, on top of the general rules in [CLAUDE.md](CLAUDE.md)
and, for new projects, [NEWPROJECTS.md](NEWPROJECTS.md).

## Preferred stack

- In an existing project, use what the project already uses. Do not migrate it to the stack
  below unless I ask.
- In a new project, use:
  - `asyncio` for concurrency
  - `sqlalchemy` (2.x, async) for database access
  - `fastapi` for HTTP APIs
  - `pydantic` for models, validation and settings
  - `uv` for dependency and environment management
  - Python 3.14
  - Logging: the standard `logging` module.
  - HTTP client: `httpx`.
  - Linting and formatting: `ruff`.
  - Configuration and secrets: `pydantic-settings`.
  - Data modelling: DTOs are pydantic models, database models are SQLAlchemy models, everything
    else (services, repositories, etc.) is a `@dataclass`.
  - Layout: sources live in `src/`; modules follow [NEWPROJECTS.md](NEWPROJECTS.md).

## Definition of done

After every code change, before reporting the task as complete, you **must** run whatever the project ships:

1. Run the code — confirm it executes.
2. Run the tests (e.g. `pytest`) — confirm they pass.
3. Run the formatters/linters (`ruff`, `black`, `pre-commit run -a`, depending on the project) — confirm they pass.

Do not skip these steps, and don't rely on CI to catch what you missed.

## Code style

- Always annotate types.
- When working with JSON, always use pydantic.
- Define classes with `@dataclass`.
- In `.ipynb` notebooks: keep all imports in a single cell at the very top of the file; put all settings (e.g. chart colors) in the second cell.
