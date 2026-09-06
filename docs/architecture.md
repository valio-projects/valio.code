# Backend component boundaries

The architecture centers on Project (service, library, application or tool). A Workspace
is the access boundary, a Repository is a source location, and a Worktree is an observed
checkout. Many projects can share one source root; one project can span repositories.

```text
CLI / HTTP / MCP
      -> application services
         -> repositories (interfaces)
            -> SurrealDB repository adapters
         -> providers (Git, compiler, AI; explicit capabilities)
         -> processors (bounded analysis transformations)
            -> readers -> parsers -> models -> mappers -> builders -> writers
workers -> durable job repository -> processors -> fenced publication
```

Go uses structs with methods and small interfaces for these boundaries. Packages group
coherent responsibilities; files separate models from algorithms and IO. There is one
Go module for the initial deployable system, with several command binaries. Independent
modules are justified only when separate versioning and dependency graphs are needed.

Patterns used where they enforce a real boundary:

- Repository: fixed database/table, typed records, immutable vs mutable operations.
- Adapter/provider: Git CLI, filesystem rooted IO, SurrealDB transport, compiler APIs.
- Strategy: registered job processors and language/compiler capabilities.
- Builder: manifests, indexes, views, captured snapshots.
- Mapper: language-specific analysis facts to scoped domain type descriptions.
- Dependency injection: processors and IO can be tested without live external systems.

No service constructs SQL. No compiler provider silently claims support when its toolchain
or dependencies are absent. Libraries from the Go standard distribution implement parsing,
type checking, HTTP, JSON, cryptography, rooted filesystem access and process execution.
External protocol/SDK dependencies must be pinned and covered by compatibility fixtures.

All persistent state belongs to SurrealDB. Agent spool is a transport recovery journal,
not another searchable database. A source policy runs before spool, hashing and upload.

Application lifecycle: capture immutable input -> validate sanitized bundle -> write staged
facts -> validate version/scope -> publish a pointer atomically. An old view remains readable.
Expensive compiler/provider work never runs inside a database transaction.

Type facts distinguish identity, version and occurrence. Metadata includes attributes on
types/members/parameters, modifier/accessibility semantics, enum members and usages, and
layout facts tied to architecture/ABI/compiler/build profile. Unknown size or alignment is
not represented as zero. The rich-type contract does not imply five-language extraction
has already been implemented.
