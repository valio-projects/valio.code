# Intelligence wave: current capability and limits

`valio.code` is a local vertical slice, not a completed v1 analysis stack or a
production-scale search claim. See [language analysis](language-analysis.md),
[AI embedding providers](../getting-started/ai-providers.md), and the checked-in
[AI profile example](../../../deploy/ai-profiles.example.json).

## Current capability

An immutable project view persists source, AST chunks, symbols and relationship
facts. Go uses partial local `go/types` analysis for cross-file symbols,
references, calls, reads and writes. C, C++, C#, Java, JavaScript, TypeScript,
TSX and JSX contribute validated syntax reports and TypeDescriptors with written
members, parameters, return types, visibility, modifiers and raw attribute
syntax.

`/api/v1/types` and MCP `type_query` serve these descriptors. A non-Go type
descriptor is syntax evidence: it has no compiler-resolved symbol, ABI/layout,
effective enum constant value, overload target or cross-file type/call
resolution. Explicit enum expressions remain expressions; implicit values and
generics that the helper does not report remain unknown.

Method AST chunks from the non-Go reports support bounded parent-context
expansion. `/api/v1/structure/graph` and MCP `structure_graph` expose
declaration/member/parameter containment plus unresolved imports and call
observations. MCP now has 19 tools.

Retrieval combines lexical BM25, symbols, normalized-hash structural candidates,
exact-cosine semantic ranking, reciprocal-rank fusion and an optional reranker.
Structural matching is not general MinHash; exact cosine is not an HNSW claim.

## AI provider evidence

JSON profiles configure Ollama, LM Studio/OpenAI-compatible, remote and
Docker-host providers. Smoke used `qwen3-8b-lmstudio-docker` (Q8_0, 4096
dimensions) in project `demo-6ebf95f8fc7e`, view
`7e6eddbb4543be8e0bb330ecd5ab65f81c6ce9cdf3b1019da69a767f52f7e93f`.
It persisted 29 code embeddings and ran semantic and hybrid retrieval. This
validates that local provider path, not remote availability or retrieval quality.

## SurrealDB typed CBOR boundary

SurrealDB 3.2.4 JSON RPC has legacy coercion in which `id: number` becomes the
record ID `id:number`. Repository records now use typed CBOR through the pinned
SDK codec. A regression verifies source-like, UUID, date and Unicode strings
and 64-bit integers, so these values retain their typed representation.

## Acceptance status

| Scope | Current state | Boundary |
|---|---|---|
| Go | Local cross-file facts through `go/types` | No complete external-package or build-tag coverage. |
| C/C++/C#/Java/JS/TS/TSX/JSX | Validated syntax, type descriptors and non-Go method chunks | No compiler semantics, CFG/DFG or cross-file resolution. |
| Type and structure APIs | `/types`, `type_query`, `/structure/graph`, `structure_graph` passed Windows tests and smoke | No compiler resolution or semantic graph claim. |
| Retrieval | BM25, symbols, structural candidates, exact cosine, hybrid RRF and optional reranking | No scale, HNSW or quality benchmark claim. |
| LM Studio profile | Docker-host Q8_0, 4096 dimensions, 29 persisted code embeddings and semantic/hybrid retrieval | Model digest and remote providers are operator-managed. |

## Verification

Windows passed `go test ./... -count=1` and `go vet ./...` against real
SurrealDB 3.2.4 and Jaeger, including syntax integration. Docker built the API,
worker and migration images and `docker compose up -d --wait` reached
healthy services. Linux Docker passed `go test -race ./...`; its
database-dependent tests are skipped in that build.

`scripts/smoke.ps1` passed seven rich user types, six non-Go structure graphs,
syntax search, Go graph reads, policy handling and idempotence. The earlier
pinned view created with `87d2c0b` still returns one Go type candidate and five
lexical hits. MCP/HTTP parity passed for `structure_graph`, `type_query`,
`code_search`, `retrieval_search` and `symbol_search`; the optional-field fix
makes MCP defaults match HTTP.

## Still open

Full v1 still needs SCIP-grade semantics for five languages, CFG and
interprocedural analysis, SQL/API/event linking, and test and history
intelligence. GNN-based capabilities are post-v1 research. Contracts and
roadmap entries are not evidence that those features are running.
