# Intelligence wave: current capability and limits

This is a tested local vertical slice, not a completed v1 analysis stack or a
production-scale search claim. See [AI embedding providers](../getting-started/ai-providers.md)
and the checked-in [AI profile example](../../../deploy/ai-profiles.example.json).

## What is available

An immutable project view can persist source, AST chunks, symbols and
relationship facts. Go uses `go/types` for local cross-file symbols,
references, calls, reads and writes. It does not promise resolution through
external packages or every build-tag combination.

The WASM syntax helper parses C, C++, C#, Java, JavaScript, TypeScript, TSX and
JSX. Names, members, parameters, enum values and attributes are persisted in
SurrealDB. Query these non-Go syntax facts with `syntax_query` or HTTP
`POST /syntax/query`; the legacy `/types` endpoint remains Go-only. They are
navigation facts, not compiler-resolved cross-file calls, CFG, or data flow.

Retrieval combines lexical BM25, symbol lookup, normalized-hash structural
matches, exact-cosine semantic ranking, reciprocal-rank fusion and an optional
reranker. Structural matching is not general MinHash; exact cosine is not an
HNSW/vector-index claim. Parent/child context expansion is bounded.

There are versioned contracts for `code`, `symbol`, `context`, `documentation`,
`architecture`, `change`, `error`, and `api`. Current producers create code,
symbol, context, documentation, error and API facts. Architecture and change
have no producer.

HTTP has POST endpoints for retrieval/search, graph/query, context,
embeddings/index, AI/probe and syntax/query. MCP exposes 18 scoped tools,
including retrieval, context, graph, reads/writes, embedding index, AI probe
and syntax queries alongside the original project/search/type/status tools.

## AI profiles

JSON profiles configure Ollama, LM Studio/OpenAI-compatible, remote and
Docker-host providers. The example includes quantized Qwen3 embedding 4B/8B
aliases and validates dimensions before records can mix. Configuration does not
confirm reachability or model availability: load/import the model, then probe.
Native LM Studio doctor and Docker-host API probes passed with
`text-embedding-qwen3-embedding-8b` (Q8_0), dimension 4096. Eight representations
from a synthetic project were stored in SurrealDB and used by semantic and
hybrid retrieval. Ollama was absent on its default local port. Unit tests still
use synthetic vectors; these checks are not a retrieval-quality benchmark.

## Acceptance matrix

| Scope | Validated result | Boundary |
|---|---|---|
| Go | AST, members and local cross-file references/calls/reads/writes | Partial `go/types`; no complete external-package or build-tag coverage. |
| C/C++/C#/Java/JS/TS/TSX/JSX | Parsed AST, declarations and members persisted | No compiler semantics, CFG/DFG or cross-file resolution. |
| Text and symbols | Lexical BM25 and scoped symbol retrieval | No production FTS scale claim. |
| Structural retrieval | Normalized-hash candidates | Not general MinHash or semantic equivalence. |
| Semantic retrieval | Exact cosine, hybrid RRF, optional reranking; live LM Studio embeddings | No HNSW or measured retrieval-quality claim. |
| Linux Docker test target | `go test -race` and syntax-helper fixtures for every listed language passed | A test-target result, not runtime smoke. |
| Windows and Compose | Full Go tests, `go vet`, real SurrealDB integration and multi-language runtime smoke passed | Small fixtures; no scale claim. |
| Persistence | Graph/retrieval integration, synthetic 3-D test vectors and live 4096-dimensional LM Studio vectors passed | Exact model artifact digest remains operator-managed. |

These are the exact validation results currently recorded. Commits `9277db0`
and `641db16` predate this intelligence wave and are not evidence for it.

## Still open

Full v1 still needs SCIP-grade semantics for five languages, CFG and
interprocedural analysis, SQL/API/event linking, and test and history
intelligence. GNN-based capabilities are post-v1 research. Contracts and
roadmap entries are not evidence that those features are running.
