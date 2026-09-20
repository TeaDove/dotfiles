# New project rules

These rules apply when starting a new project in any language, on top of the general rules in
[CLAUDE.md](CLAUDE.md). In an existing project, follow what the project already does. Language
specifics are in [GO.md](GO.md) and [PYTHON.md](PYTHON.md).

## Reference project

A good example to model a new project on: https://github.com/TeaDove/neko-manager

## Layout

Follow an onion architecture: split every project into modules by layer, dependencies pointing
inwards.

- `repo` — database access.
- `supplier` — calls to third-party services and APIs.
- `service` — business logic.
- `api` — presentation layer: REST API, gRPC, etc.
- `util` — shared utilities.
- `dto` — shared DTOs; this should mostly not be needed.
