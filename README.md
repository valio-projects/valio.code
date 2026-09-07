# valio.code

**Ask a codebase a focused question and keep the evidence behind the answer.**
`valio.code` turns a pinned project view into searchable source, typed
declarations, relationships and bounded context for developers and coding
agents.

It is AGPL-3.0 software and an evolving local vertical slice, not completed v1.
Read the [intelligence-wave status](docs/en/reference/intelligence-wave.md) and
[language-analysis reference](docs/en/reference/language-analysis.md) before
relying on a capability.

[English documentation](docs/en/README.md) · [Русская документация](README.ru.md)

## What it helps with

- Find source, symbols, declarations and error text while retaining the
  repository and immutable view that supplied the result.
- Give agents bounded, cited source and structural context instead of an entire
  repository dump.
- Combine lexical BM25, symbols, normalized-hash structural candidates, exact
  cosine embeddings, hybrid RRF and an optional reranker.
- Navigate containment from a type to its members and parameters, and inspect
  unresolved import and call observations without mistaking them for facts.
- Use the same view-scoped capabilities through HTTP and 19 MCP tools.

## Current intelligence layer

Go has partial local cross-file analysis through `go/types`. C, C++, C#, Java,
JavaScript, TypeScript, TSX and JSX supply validated syntax reports and typed
descriptors for written classes, structs, interfaces, enums, members,
parameters, return types, modifiers, visibility and raw attribute syntax.
`/api/v1/types` and MCP `type_query` serve them for all of these languages.

Non-Go descriptors are syntax evidence. They do not assert compiler-resolved
types, imports, calls or overloads; ABI/layout; data/control flow; or evaluated
enum values. A written expression such as `1 << 2` remains an expression.
`/api/v1/structure/graph` and MCP `structure_graph` expose declaration/member/
parameter containment with explicitly unresolved imports and calls. Non-Go
method chunks support bounded parent-context expansion.

## Representation contracts

Eight versioned representations keep retrieval inputs distinct: code, symbol,
context, documentation, architecture, change, error and API. Current producers
create code, symbol, context, documentation, error and API facts. Architecture
and change remain contracts without a producer, so they are not returned as
available evidence.

JSON profiles support Ollama, LM Studio/OpenAI-compatible, Docker-host and
remote embedding endpoints. The verified local smoke used
`qwen3-8b-lmstudio-docker` (Q8_0, 4096 dimensions), persisted 29 code
embeddings, and ran semantic/hybrid retrieval. This is integration evidence,
not a quality benchmark or remote-provider guarantee. Begin with the
[AI profile example](deploy/ai-profiles.example.json).

## Status and remaining work

The second integration wave is verified on Windows with the full Go test suite,
`go vet`, real SurrealDB 3.2.4 and Jaeger, and a Compose smoke scenario. Docker
also built the API, worker and migration images and reached healthy Compose
services. Linux Docker `go test -race ./...` passed; database-dependent tests
are skipped in that Linux build.

Full v1 still needs SCIP-grade semantics for five languages, CFG and
interprocedural analysis, SQL/API/event linking, and test/history intelligence.
GNN capabilities are post-v1 research.

## Run locally

With Docker and Docker Compose installed, run from the repository root:

```powershell
docker compose up -d --build --wait
```

Open <http://127.0.0.1:8080> and use the local development token:

```text
valio-local-development-token-0001
```

Jaeger is at <http://127.0.0.1:16686>. For installation, native agent and
frontend development, use the [getting-started guide](docs/en/getting-started/installation.md).
