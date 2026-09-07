# valio.code

**See how a codebase fits together before you change it.** `valio.code` helps
developers and software agents explore a project, find the code behind a
question, understand the type information that is available, and keep every
answer tied to a specific source version.

The product direction is a software intelligence graph: repositories, projects,
source, symbols, and their relationships become useful context for development,
reviews, incident response, and automation.

`valio.code` is licensed under AGPL-3.0. It has a working local foundation and
a broader product roadmap. The sections below keep those two states separate.

[English documentation](docs/en/README.md) · [Русская документация](README.ru.md)
· [Implementation status](docs/en/reference/implementation-waves.md)

## What you can do today

### Work from a stable view of the code

Create projects from one or more repositories and source roots. Each published
snapshot creates an immutable view, so an investigation, review, or agent task
can continue to refer to the exact version that supplied its result. Shared
repositories can belong to more than one project without losing that context.

The native agent can capture a repository once or watch it for changes. It keeps
retryable work locally, reconciles filesystem events, and records Git metadata.
Its capture policy removes configuration values before source is stored or sent.

### Find the code you need

Search supports exact text, substring matching, regular expressions, and a
boolean query language with `AND`, `OR`, and `NOT`. Results stay scoped to the
chosen project and version and point back to the matching source. This makes it
useful for finding an identifier, checking a migration, locating an error
message, or narrowing down a change before opening an editor.

### Explore available Go type information

The current Go analysis and type lookup show declarations, members, parameters,
generic constraints, tags, named types, and constants. The type model is also
ready to represent visibility, modifiers, attributes, enums, and physical layout
when a language adapter can establish them. Values that are not known are shown
as unknown instead of being guessed.

Today this is partial Go support, rather than whole-program semantic analysis.
Support for other languages and compiler-grade cross-file resolution belongs to
the roadmap.

### Use the same knowledge from the web, API, or an agent

The web app, authenticated HTTP API, and MCP use the same application
operations. MCP currently exposes `project_list`, `project_get`, `code_search`,
`type_query`, and `index_status`. A developer can inspect a result in the web
app while an agent requests the same version-scoped information through MCP.

The local stack includes SurrealDB, durable-job foundations, OpenTelemetry, and
Jaeger tracing.

## The planned intelligence layer

The roadmap adds retrieval and relationships without treating generated text as
the source of truth. Canonical code remains the evidence; richer retrieval helps
people and agents get to the right evidence faster.

### RAG that preserves the path back to code

The planned RAG experience combines these variants:

1. **Hybrid retrieval** combines verified exact matches with lexical and
   semantic results, so rare identifiers remain easy to find while natural
   language queries gain useful recall.
2. **Parent-child retrieval** finds a focused function, method, or declaration
   and can add only the relevant parent or child context.
3. **Summary-assisted retrieval** uses documentation and generated summaries as
   helpful clues; canonical code remains available beside them.
4. **Graph and context retrieval** follows selected symbols and relationships
   within a fixed project and version instead of returning a whole repository.
5. **Reranked evidence** orders candidate context for a question and retains
   links back to the source and version used to support it.

These capabilities are planned. There is no implemented RAG context builder,
embedding service, BM25 index, vector search, or cross-encoder reranker today.

### Eight planned embedding views

Different questions need different views of a codebase. The planned embedding
profiles are complementary, so teams can measure which combination helps their
repositories rather than assuming one model fits every query.

| Planned view | Developer benefit |
|---|---|
| Source code | Finds implementation intent while preserving identifiers and syntax. |
| Symbols and signatures | Improves API, type, method, and service discovery. |
| Local context | Connects a result to its surrounding module or component. |
| Documentation | Matches product language, comments, and developer questions. |
| Architecture | Helps navigate components and their responsibilities. |
| Change history | Connects current code with the reason and impact of a change. |
| Errors and diagnostics | Speeds up investigation of recurring failures. |
| APIs and events | Connects contracts, integrations, messages, and boundaries. |

General multilingual models, code-specialized models, structural or graph
representations, and contrastive retrieval models can overlap in value. They
are planned inputs to evaluate together, not mutually exclusive editions.

### Relationships that answer development questions

The planned graph is grouped by the question it answers:

- **What is this, and where is it used?** Syntax trees (AST), symbols, imports,
  exports, references, aliases, and external resources help locate ownership and
  usage.
- **What does a type contain?** Type and member relationships connect classes,
  structs, fields, properties, methods, and parameters to their types,
  attributes, and modifiers. They will help find interface implementations and
  individual enum-value usages, inspect accessibility, and understand layout.
- **What could a change affect?** Calls, control flow (CFG), data flow (DFG),
  dependencies, and bounded interprocedural paths help assess impact.
- **Why does an operation depend on this condition or value?** Control
  dependence (CDG) and program dependence (PDG) graphs isolate relevant chains
  of conditions and data. A Code Property Graph (CPG) connects syntax, control
  flow, and dependencies for combined queries.
- **Where has a similar problem been solved?** Structural and graph-role
  similarity will help locate similar algorithms, handlers, and components
  even when their names and descriptions differ.
- **How does the system behave at its boundaries?** API contracts,
  configuration, SQL, events, deployments, and external services explain how
  code meets the rest of the system.
- **Why should I trust this change?** Tests, documents, errors, change history,
  ownership, and risk signals provide context for review and maintenance.

These graph families—including call graphs, CFG/data-flow analysis, API/SQL/event
links, and history analytics—are roadmap work. Current search and Go type lookup
do not claim to provide them.

Read the [retrieval design](docs/en/design/retrieval.md),
[code-graph and embedding design](docs/en/design/code-graphs-and-embeddings.md),
[local agent configuration proposal](docs/en/design/agent-configuration.md),
[architecture](docs/en/contributing/architecture.md), and
[type-information contract](docs/en/reference/type-information.md).

## Benefits for developers and agents

Developers get results they can reproduce: a source answer carries its project,
repository, and immutable view. That makes reviews and investigations easier to
share and reduces the risk of confusing a changed working tree with the version
that was actually examined.

Agents get a bounded working context instead of an undifferentiated code dump.
They can search and inspect types through MCP today; the roadmap adds selected
relationship and retrieval context while retaining provenance and explicit
unknown states. This supports more useful agent assistance without hiding the
evidence a developer needs to verify a suggestion.

## Run locally

Docker Compose is enough to run the local stack. You do not need a host SDK,
setup script, or `.env` file.

```powershell
docker compose up -d --build --wait
```

Open <http://127.0.0.1:8080> and sign in with:

```text
valio-local-development-token-0001
```

Jaeger is available at <http://127.0.0.1:16686>. The local Compose database
uses the example root account `root` / `valio` inside the Compose network. The
token and database credentials are for local development only.

Go 1.26.8 is needed to run the native agent from source. Node.js 26.8.1 is
needed only for frontend development outside Compose. For project setup, follow
the [getting-started guide](docs/en/getting-started/installation.md), then run:

```powershell
go -C src/back-end run ./cmd/valio-agent index --root C:/path/to/repository --server http://127.0.0.1:8080 --repository repo-id
go -C src/back-end run ./cmd/valio-agent watch --root C:/path/to/repository --server http://127.0.0.1:8080 --repository repo-id
go -C src/back-end run ./cmd/valio-mcp serve --server http://127.0.0.1:8080
```

For configuration and operating guidance, see the [English documentation](docs/en/README.md)
or [Russian documentation](docs/ru/README.md).
