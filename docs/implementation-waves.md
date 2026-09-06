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
   explicitly unknown. See type-information.md for the detailed contract and actual coverage.
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
