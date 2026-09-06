# valio.code development

Preserve the existing AGPL-3.0 license, Git remote, and user changes.

- SurrealDB is the only persistent application store and runs through Docker Compose.
- Keep Workspace, Project, Repository, Worktree, and Deployment distinct.
- Organize Go as focused packages and files with structs, methods and interfaces.
  Separate models, parsers, readers, writers, mappers, builders and adapters by
  responsibility. CLI entrypoints only wire dependencies and dispatch commands.
  Use design patterns at actual IO/algorithm boundaries; avoid giant files and
  wrappers which introduce no responsibility or invariant.
- Each important result carries concrete versions, scope, evidence and completeness.
- Filter configuration values before any upload or spool write. Do not log source payloads.
- Do not label AST heuristics as compiler-resolved facts or missing metadata as zero.
- Complete backend foundation and analysis contracts before frontend API and frontend.
- At the end of **each completed stage or development wave**, run the relevant checks,
  stage only that completed work with `git add`, then create a descriptive `git commit`.
  Record the wave's implemented capabilities, validation, and remaining limitations in docs.
- The coordinating agent integrates and commits parallel changes; contributors do not
  commit another contributor's unfinished work. Do not push unless explicitly requested.
- Use `gpt-6-astra` with `high` reasoning for requested parallel development agents,
  respecting the available concurrency limit. Tasks own disjoint paths.

Tools installed for this Windows workspace are ignored under `.tools/`; use an installed
Go 1.26 toolchain or `.tools/go/bin/go.exe`. Do not commit tool binaries or local credentials.
