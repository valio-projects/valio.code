# Implementation waves and release gates

The approved v1 is a product-scale target. A foundational package or declared projection
contract is not a completed analyzer. The status of executable features is documented
separately from the target architecture. SurrealDB must run through Docker Compose.

## Execution order

1. **Foundation / backend wave**: repository setup, Go toolchain, Compose, migrations,
   database isolation, lease fencing, native capture and configuration filtering,
   project membership, immutable manifests and views, verified text search and Go AST.
2. **Rich type metadata wave**: query types by name and versioned project context;
   members, parameters, accessibility, modifiers, attributes, enum values and uses,
   underlying types, dimensions and compiler-backed layout. Unknown information remains
   explicitly unknown. See [the type-information contract](type-information.md) for the detailed contract and actual coverage.
3. **Frontend API wave**: application services, authenticated HTTP contracts, immutable
   ingestion/publication, search/view/type APIs and integration tests.
4. **Frontend wave**: React/Vite/HeroUI pages consuming those contracts, real empty/error/
   partial states, workspace/project/version selection and rich type inspection.

API and frontend are last in this initial implementation sequence at the user's request.
The first waves used five GPT-6 Astra / high agents across waves, subject to the
runtime limit of four simultaneously active agents including the coordinator.
Subsequent tasks use GPT-5.6 Terra / high per the user's cost preference.

## Completion and commits

Go implementation uses focused packages and files: models and validation; readers
and writers for IO; parsers for syntax; mappers for domain facts; builders for
manifests/indexes/views; adapters for storage and transports. Struct methods and
interfaces provide the class-like boundaries requested by the user. Dependency
injection isolates side effects, and CLI entrypoints only compose components.

For **every completed stage and every completed wave**:

1. Integrate the assigned paths and inspect the diff for unrelated work and credentials.
2. Run relevant unit, compatibility and integration checks; record unexecuted checks.
3. Update the implementation status, limitations and acceptance evidence.
4. `git add` the completed paths and `git commit` a descriptive conventional commit.
5. Record the resulting commit in the next status update. Publishing is a separate action.

A failing acceptance check means the stage remains in progress. Independent completed
stages can be committed without including unfinished changes from another agent.

## Full v1 backlog

| Approved stage | Required outcome | Initial status |
|---|---|---|
| 0 Compatibility | SurrealDB 3.2.4, FTS/HNSW, fencing, five SCIP adapters, SQL and routing baselines | In progress |
| 1 Bootstrap | Go, frontend build, CI, Docker targets, Compose, health, migrations | In progress |
| 2 Workspace/Project/version model | Access, shared roots, revisions, pinned views, durable jobs | In progress |
| 3 Agent/policy | Watch reconciliation, Git, ignore, sanitized resumable delivery | In progress |
| 4 Text/AST | Verified DSL search, persistent candidate indexes, AST, search UI | In progress |
| 5 SCIP/symbols/types | Five compiler adapters, identity/bindings, rich type metadata and usage | In progress: Go AST and contracts |
| 6 Calls/CFG | Confirmed calls, dispatch, reads/writes, CFG and evidence | Not implemented |
| 7 Interprocedural flow | Summaries, aliasing, SCC/fixpoint, widening and invalidation | Not implemented |
| 8 Dependency/configuration | Dependency resolution, provider precedence, configuration traces | Metadata capture only |
| 9 API/database graph | REST/gRPC, PostgreSQL/T-SQL, schema revisions, mapping and impact | Not implemented |
| 10 Event graph | Protocol-specific Kafka/RabbitMQ/NATS/Redis routing, contracts and flow | Not implemented |
| 11 Documents/tests/errors | Sections, test discovery, versioned report import and stack mapping | Not implemented |
| 12 History/analytics | DAG, semantic diff, co-change, ownership, risk, retention/GC | Risk function only |
| 13 Structural/role similarity | Normalization, MinHash, persistent buckets, independent evaluation | Syntax fingerprint only |
| 14 Embeddings/context | Eight profiles, HNSW, fusion, cited context, optional generation | Not implemented |
| 15 MCP/UI | Shared typed application operations, stdio/HTTP MCP, accessible workflows | Initial Project/search/type slice implemented; full graph workflows pending |
| 16 Release | 10M-line benchmarks, recovery, restored backups, security, distributions, SBOM | Not implemented |

Every backlog feature must specify input entities, output facts/projections, dependencies,
invalidation unit, edge cases, independent fixtures, partial-result diagnostics and cost/
quality measurements. No v1 completion or performance claim is made by this document.

## Completed increments

- `44b190e`: Compose SurrealDB 3.2.4, schema provisioning, database isolation and fenced
  durable queue. Real container integration tests passed.
- Agent/search increment: Git-safe capture, mandatory configuration projection, atomic
  spool, HTTPS upload, validation on reloaded bundles; verified boolean text search and
  modular Go AST/compiler provider. Focused tests and vet passed on Windows; the initial
  Docker build/test target also passed on Linux. Watch polling is still the baseline at
  this increment. Five-language compiler coverage and production search scale remain open.

- `3ea1d95`: Project/roots/revision/evidence contracts and rich type model with
  attributes, modifiers, enum metadata and explicitly unknown ABI-dependent layout.
- `ec9aa20`: repository interfaces in domain, Surreal adapters in infrastructure,
  application workers, Uber Fx and OpenTelemetry. Real Jaeger export test passed.
- `4ea3456`: fsnotify plus reconciliation/retry, bounded repository metadata and
  spool validation, public model and method documentation.
- Integrated HTTP/MCP and React slice: real agent upload through Compose ingress,
  immutable publication, Go type lookup, search and duplicate receipt verified.
  A `.env` fixture's value is absent from searchable source. Full v1 remains open.
- `73677f5`: authenticated HTTP and MCP, domain repository ports, transactional
  source/view publication, generated OpenAPI and API readiness command.
- `386c1f3`: bounded search highlights with exact matched-file totals, plus public
  GoDoc across agent, search, MCP, telemetry and watcher contracts.
- `9aef8ea`: React/HeroUI Project forms, search, source and type inspection. Twelve
  component/unit tests passed; the live login layout and rejection path were checked.

The current foundation passes Windows Go unit/vet checks, real Docker SurrealDB
integration and Linux race tests. These are correctness checks on small fixtures,
not evidence for 10-million-line scale, five-language compiler support or full v1.

## Repository and documentation wave

- `e6b426b`: moved tracked application sources with `git mv` into `src/back-end`
  and `src/front-end`; updated Docker, scripts and CI. The backend module import
  path stays unchanged.
- Local Compose now uses one explicit `root` / `valio` database account across
  migration, API and worker. This replaces the earlier separate-password decision
  for the local example. Setup scripts and the environment-file prerequisite were
  removed. Source policy and immutable/evidence contracts still apply.
- Existing local data survived credential rotation. Isolated Compose integration
  checks and the native-agent smoke scenario passed after the move. Twelve web
  tests and the production build passed from the new source directory.

## Design follow-ups, not implemented capabilities

The [intelligence wave](intelligence-wave.md) implements the first bounded
version of the chunk, local Go graph, BM25, embedding and hybrid increments
listed below. This table describes their broader acceptance goals; it is not
a statement that their initial implementations are still absent.

The reference-porting queue now includes `codebase-memory-mcp` at revision
`aa44c28ea5ea82a5f811f0bace4f7857a68cac80` (MIT). Candidates are targeted index
coverage checks, architecture summaries, structural graph queries, HTTP service
linking, infrastructure graphs, call tracing, ADRs and AST/LSP integration.
Each needs implementation-level review and fixtures before porting; README
performance claims are not valio.code benchmarks. See the
[porting queue](../../ru/design/codebase-memory-porting.md).

`chunkhound` is also queued at `2ab775354d8d378bdacc8ea76b51a588362cfb99` for
implementation-level review of chunking, hybrid retrieval, model providers,
Git research and cited context. See its [queue](../../ru/design/chunkhound-porting.md).

The documentation wave adds English and Russian product READMEs, installation,
configuration, contribution and observability guides, and architecture with ten
diagrams per language. Design notes cover graph extensions, eight embedding
representations, retrieval strategies and separate agent/project configuration.
They distinguish existing behavior from proposed features. Markdown links and
all 24 Mermaid diagrams are checked; this wave adds no new analysis capability.

| Priority | Increment | Required foundation and acceptance |
|---|---|---|
| 1 | Base agent configuration and separate Project profiles | Validated file reader, explicit precedence, multiple source roots, linked-project scope and deterministic reload; no accidental remote/local Project identity conflation |
| 2 | AST chunk records, generated metadata headers and bounded parent expansion | Immutable source spans, chunk/profile fingerprints, oversized-function fixtures, budget and policy tests |
| 3 | Compiler-backed import/export, symbols, references and calls | Compilation-unit analysis, unresolved candidates, dependency invalidation, cross-file and cross-project fixtures |
| 4 | CFG, def-use, control dependence and PDG slicing | Versioned IR, branch/exception semantics, alias policy and independently checked slices |
| 5 | BM25, dense embeddings, hybrid fusion and reranking | SurrealDB compatibility, held-out Russian/English queries, latency/cost budgets, baseline ablations and model-space isolation |

These increments refine the approved stages rather than adding separate storage
systems. See [graphs and embeddings](../design/code-graphs-and-embeddings.md),
[retrieval](../design/retrieval.md) and [agent configuration](../design/agent-configuration.md).
