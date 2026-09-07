# valio.code

**Ask a codebase a focused question and keep the evidence behind the answer.**
`valio.code` turns a pinned project view into searchable source, symbols,
relationships and bounded context for developers and coding agents.

It is AGPL-3.0 software. The current implementation is a tested local vertical
slice, not a completed v1. Read the [intelligence-wave status](docs/en/reference/intelligence-wave.md)
before relying on a particular analysis result.

[English documentation](docs/en/README.md) · [Русская документация](README.ru.md)

## What it helps with

- Find the implementation, symbol or error message behind a question without
  losing the project and immutable view that supplied it.
- Give an agent a small, scoped set of cited source and relationship evidence
  instead of an undifferentiated repository dump.
- Combine exact and lexical search with symbols, structural candidates,
  semantic ranking, hybrid RRF and an optional reranker when a suitable model
  profile is available.
- Traverse local code facts to investigate impact: declared symbols, references,
  calls, reads, writes, AST structure and bounded parent/child context.
- Use the same scoped capabilities through HTTP and 18 MCP tools, so a person
  and an agent can work from the same evidence.

## Current intelligence layer

Go has local cross-file symbol, reference, call, read and write analysis based
on `go/types`; it is deliberately partial for external packages and build-tag
combinations. C, C++, C#, Java, JavaScript, TypeScript, TSX and JSX are parsed
by a WASM syntax helper. Their declarations, members, parameters, enum values
and attributes are persisted for navigation.

Retrieval stores AST chunks for Go and a fallback for the other parsers. It can
rank lexical BM25, symbols, normalized-hash structural candidates and exact
cosine embeddings, fuse candidates with RRF and optionally rerank them. The
result keeps its source/view scope. Structural hashes are not general MinHash;
exact cosine is not an HNSW claim; non-Go syntax facts are not compiler-resolved
cross-file calls or data flow.

There are eight versioned representation contracts: code, symbol, context,
documentation, architecture, change, error and API. Current producers create
the first four plus error and API facts. Architecture and change have no
producer yet.

JSON AI profiles support local Ollama and LM Studio/OpenAI-compatible services,
Docker-host and remote endpoints. The supplied Qwen3 4B/8B quantized aliases
are guarded by model/dimension identity. A configured profile does not prove a
remote model is running; load or import it and use the explicit probe. Start
with the [AI profile example](deploy/ai-profiles.example.json).

## Evidence and remaining work

SurrealDB integration validation passed graph, retrieval, synthetic 3-D
embeddings and six non-Go language parsers. The Docker build/race validation was
still running when the wave record was written, so it is not claimed as passed.
Implementation validation is recorded in `9277db0`; its documentation record is
`641db16`.

Full v1 work remains: SCIP-grade semantics for five languages, CFG and
interprocedural analysis, SQL/API/event links, test and history intelligence,
and GNN capabilities. See the [acceptance matrix and limits](docs/en/reference/intelligence-wave.md).

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
