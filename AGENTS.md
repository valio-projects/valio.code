# valio.code development

Preserve the existing AGPL-3.0 license, Git remote, and user changes.

- SurrealDB is the only persistent application store and runs through Docker Compose.
- Keep Workspace, Project, Repository, Worktree, and Deployment distinct.
- Organize Go as focused packages and files with structs, methods and interfaces.
  Separate models, parsers, readers, writers, mappers, builders and adapters by
  responsibility. CLI entrypoints only wire dependencies and dispatch commands.
  Use design patterns at actual IO/algorithm boundaries; avoid giant files and
  wrappers which introduce no responsibility or invariant.
- Access databases through typed repository interfaces in domain and SurrealDB repository
  adapters. Services/processors must not assemble SQL. Separate providers (external
  tools/models), processors (bounded transformations) and workers (leased execution).
- Place repository implementations in infrastructure. Keep domain independent of
  application, HTTP, database SDKs, compiler SDKs and the DI container.
- Use focused entity/enum/model files grouped by feature. Application exposes explicit
  commands/queries/handlers where mutation and reading have different invariants (CQRS).
- Use Uber Fx at executable composition roots for constructor injection and lifecycle.
  Admin HTTP, query HTTP and MCP are adapters to the same application use cases.
- Instrument application and infrastructure boundaries with OpenTelemetry. Export to
  Jaeger in Compose; never attach raw source, query text, credentials or config values.
- Prefer maintained standard-library and external packages to reimplementing protocols,
  compiler semantics, parsers or cryptography. Pin dependencies and test adapters.
  Security and scale claims require evidence; a clean abstraction alone is not proof.
- Each important result carries concrete versions, scope, evidence and completeness.
- Filter configuration values before any upload or spool write. Do not log source payloads.
- Do not label AST heuristics as compiler-resolved facts or missing metadata as zero.
- Complete backend foundation and analysis contracts before frontend API and frontend.
- At the end of **each completed stage or development wave**, run the relevant checks,
  stage only that completed work with `git add`, then create a descriptive `git commit`.
  Record the wave's implemented capabilities, validation, and remaining limitations in docs.
- The coordinating agent integrates and commits parallel changes; contributors do not
  commit another contributor's unfinished work. Do not push unless explicitly requested.
- Use `gpt-5.6-terra` with `high` reasoning for new delegated tasks (the user's latest
  cost preference supersedes the initial Astra request),
  respecting the available concurrency limit. Tasks own disjoint paths.
- Document public types, interfaces, fields, methods, parameters and enum values with
  useful GoDoc: meaning, units, invariants, unknown states, side effects and guarantees.

Tools installed for this Windows workspace are ignored under `.tools/`; use an installed
Go 1.26 toolchain or `.tools/go/bin/go.exe`. Do not commit tool binaries or local credentials.
