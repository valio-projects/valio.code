# Code graphs and embeddings

This is a design recommendation, not a claim that the proposed graphs already exist. It records the actual boundary first, then prioritizes graph work that gives developers and agents reliable, bounded answers.

## Actual capability audit

The publication unit is an immutable workspace-scoped source view. It pins repository snapshots, file membership, project revisions and a syntax profile. Source text is persisted in SurrealDB and text search is verified against that source. Go files are parsed during ingestion; the per-file JSON report has syntax declarations, identifier occurrences, imports, Go type outlines, function signatures and diagnostics. Rich descriptors are mapped into a scoped, evidence-backed type catalog. A symbol reference can be exact, candidate or unresolved: a lexical name is not silently promoted to an exact link.

| Area | Actual state | Boundary |
|---|---|---|
| Source and text | Ready: immutable file/view publication and verified exact, substring and RE2 search | Search makes no semantic or graph inference. |
| Go syntax outline | Produced at ingestion | A parser report is stored as JSON; there is no persistent AST-edge projection or generic AST query. |
| Types and symbols | Partial: Go declarations and the rich type contract are queryable | Server ingestion uses `go-ast-syntax/v1;syntax-default`; written type uses and ordinary uses are unresolved. |
| Optional `go/types` | `AnalyzeWithOptions(CheckTypes: true)` checks one supplied Go file | It does not automatically load sibling files, build tags or module context; it maps only local definitions to local uses. Ingestion does not enable it. |
| Syntax fingerprint | `structure.Fingerprint` hashes valid Go AST shape | It is not called or persisted by ingestion. Equal hashes are candidates with `equivalence=unknown`, never clone or behavioral proof. |
| Configuration keys | Agent capture retains safe key-only JSON and `.env` locations, with values redacted | Metadata capture, not a resolved configuration/dependency graph. |
| Risk | `analytics.ScoreRisk` aggregates supplied normalized factors with an explicit sufficiency threshold | No history, coverage, dependency or graph producer supplies its inputs. |
| Calls, CFG, DFG, PDG, CPG, aliases, API/SQL/event, tests and history graphs | Not implemented | The public view correctly marks references, structural fingerprint and configuration graph unsupported. |
| Vectors, embeddings, HNSW and context generation | Not implemented | No vector store, model profile, embeddings or semantic endpoint is wired. |

[`DefaultRegistry`](../../../src/back-end/internal/projections/registry.go) describes ten families and validates dependency order, fingerprints and downstream invalidation. It is good contract design, but it is not invoked by ingestion, persistence or read APIs; it must not be advertised as an operating graph pipeline. The ingestion status map is separate and smaller.

SocratiCode is a useful comparative reference, but does not alter this audit. Its resolver intentionally uses local name matching, imported-dependency name matching and one re-export hop. Its own implementation says it has no type inference and resolves method calls by name. Its `local`, `unique`, `multiple-candidates` and `unresolved` labels are confidence labels for candidates, not compiler-backed call semantics. It is useful for candidate-edge UX, storage and coverage reporting, not proof that valio.code already has such edges.

## Terms that should remain separate

An **AST** records syntactic nesting. It supports navigation, exact spans, normalization and syntax queries; it cannot resolve a name or prove a call target.

A **CFG** links executable basic blocks by possible control transfer. A **control dependence graph (CDG)** is derived from the CFG, normally using post-dominance, and states which predicate controls whether an operation executes. A CDG is not another name for a CFG.

A **DFG** here means def-use/data-dependence edges under an explicit analysis model. A **program dependence graph (PDG)** contains *both data and control dependences*; it is not merely a CFG and DFG placed side by side. A slice must declare its seed, direction, dependence kinds, interprocedural and alias policies, and treatment of exception/async paths. The original definition makes both dependence kinds explicit. [Ferrante, Ottenstein and Warren (1987)](https://bears.ece.ucsb.edu/class/ece253/papers/ferrante87.pdf)

A **code property graph (CPG)** is a property-graph representation combining AST, CFG and PDG layers at shared program nodes. It is a query representation, not a guarantee of precise analysis, a GNN, or semantic equivalence. Quality remains bounded by the language front end, build configuration, dispatch/alias policy and source coverage. [Yamaguchi et al. define the composition](https://www.sec.cs.tu-bs.de/pubs/2014-ieeesp.pdf); the [Joern CPG specification](https://cpg.joern.io/) documents CFG and control-dependence layers.

## Common facts, derived projections

Do not create one database per graph or make a CPG the source of truth. SurrealDB remains the only application store. Persist versioned **facts** with evidence, then materialize narrow, replaceable projections and adjacency/index records. Every node and edge needs its immutable view, producer/profile/version, evidence IDs, status (`ready`, `partial`, `stale`, `unsupported`, etc.) and completeness counters. A projection should reference source spans instead of duplicating source text.

```text
immutable source view + project/build profile
  -> syntax facts, compiler/SCIP facts, external-contract facts
  -> symbol/reference/type/member facts
  -> call + CFG + def-use facts
  -> PDG/slices, effects, system/test/history projections
  -> retrieval candidates and cited agent context
```

Thus a symbol graph, call graph, CPG view, impact view and agent-context view are projections over common facts, not competing stores. Each producer needs a profile fingerprint: compiler/SCIP version, language/build configuration, dependency-lock identity where used, schema version and source fingerprint. A profile change invalidates its output and declared descendants. The existing registry is a starting mechanism, but actual dependencies must be revised with real producers: a blanket `data_flow -> call` prerequisite is too coarse for intraprocedural def-use.

## Prioritized roadmap

### 1. Build-resolved symbols, references, types, members and calls

This is the first graph to build. Providers must receive an explicit build profile and return declarations, imports/modules, exact and candidate references, type/member selection and call sites. For Go, add a package/module-aware front end before calling a cross-file use exact; the existing `go/types` path demonstrates the right object-identity primitive but is deliberately insufficient as a single-file run. SCIP can be an import format/provider if it supplies stable identity and occurrence evidence, never a substitute for source/version provenance.

Persist `declares`, `refers_to`, `has_type`, `member_of`, `imports`, `calls` and `overrides/implements` only where substantiated. Candidate calls retain all targets and reasons; external/dynamic calls retain a distinct unresolved/external state. This enables find-references, caller/callee exploration, rename preflight, API discovery and bounded impact traversal.

Dependencies are package loading, build tags/targets, lockfile/dependency resolution, language adapters, stable symbol IDs and incremental per-package rebuilds. Invalidate the changed package and reverse import/call consumers, not every workspace. Fixtures need shadowing, overloads, generics, import aliases, promoted members, generated-code policy, missing dependencies, dynamic dispatch and partial builds. Publish exact/candidate/unresolved counts by language and view.

### 2. Intraprocedural CFG, def-use, PDG and bounded slices

For each resolved function body, produce basic blocks and explicit edge kinds: `normal`, `true`, `false`, `return`, `panic/throw`, `defer/finally` and `exception` when supported. Derive CDG from CFG and def-use from explicit SSA/IR or an equivalent representation. Form a PDG only when both inputs exist.

The first products are backward slices (what can affect this write, return or sink?), forward slices (what can this input affect?), dead/unreachable candidates with explanation, and control-path context for review. Each answer carries seed, direction, maximum nodes/edges/depth, timeout, included edge kinds and omitted/unknown behavior. Start intraprocedurally. Interprocedural summaries later require SCC/fixpoint/widening, context sensitivity and an alias model; local slices must not silently become whole-program conclusions.

### 3. Points-to/alias, effects, resources and concurrency

Alias precision determines whether reads/writes, side effects and interprocedural flow deserve trust. Add a configurable points-to/escape model before claiming precise heap flow. Then derive effect summaries: field/global read-write, I/O, transaction, network, filesystem, lock/channel/task creation and resource acquire/release. Async scheduling, await/join and lock order are separate facts; `happens-before unknown` is a valid answer.

Cache function/package summaries, use budgeted fixed points, and expose widening, timeout and unknown aliases. This tier enables safer impact estimates and answers about effects, leaks and deadlock candidates. Validate it first with curated positive and negative fixtures.

### 4. System contracts and operational paths

Derive API route/RPC/service, database schema/query, event producer/consumer, configuration-key/provider and infrastructure-resource relations from parsers and framework adapters. Each edge needs a contract/version, exact source/configuration span or imported artifact, and resolution status. Matching strings do not establish deployment reachability. Protocol/state-machine facts require declarations, annotations or verified patterns; heuristic transitions stay candidates.

This enables: “which handlers can reach this table?”, “what publishes this event?”, “which config keys affect this endpoint?” and “what resources does this job use?” It makes current configuration capture useful without retaining secret values.

### 5. Tests, coverage, history and risk

Import test discovery and versioned coverage reports, map them to the same source view, then derive `tests`, `covers`, failure/stack and changed-symbol relations. Add commit DAG, semantic diff, ownership and co-change only with a stated history policy. Feed the existing risk function only versioned, normalized facts, preserving its knownness/sufficiency result. “No coverage” remains unknown until a complete report for the view proves it.

### Later: deterministic structure, then GNN evaluation

Keep the current AST fingerprint as a labeled candidate filter. Add deterministic features first: normalized syntax, control/data motifs, symbol/type roles, path/context features and locality-sensitive candidate buckets. They are explainable and cheap to invalidate, but still do not prove semantic equivalence. A GNN is an optional ranking model over a stable evaluated graph, not a prerequisite for CPG queries or a substitute for analysis. Version its labels, graph snapshot/profile, repository/time split, cost, calibration and fallback.

## Agent query contract

Graph operations should return evidence before prose.

| Operation | Required bounds | Required evidence |
|---|---|---|
| `symbol_resolve` / `references_find` | `viewId`, project/profile, symbol or span, candidate limit | Exact ID or all candidates, resolution reason, spans, producer/profile and completeness. |
| `impact_trace` / `call_trace` | direction, edge kinds, depth, node/edge/time budget, external policy | Paths with edge evidence/confidence, truncated frontier, omitted dynamic/external paths. |
| `slice` | function/view seed span, forward/backward, edge kinds, interprocedural and alias policy, budget | CFG/PDG producer version, included nodes/edges, unknown aliases and cutoff reason. |
| `system_trace` | endpoint/event/table/config/resource seed, relation kinds and depth | Contract/source evidence, exact versus candidate edges, unresolved adapters and snapshot consistency. |
| `context_build` | task, token/edge/file budget, ranking profile and citation policy | Snippets with view ID, spans, relation paths, scores/features and exclusions. |

Each operation resolves `latest` once at request start and returns that concrete view ID. Reject mixed views unless an explicitly marked composite view supports the operation. Budget exhaustion is a result state, not an empty answer.

## Embeddings: choose by task, not label

Embeddings complement lexical search and graph traversal. They rank candidates; they do not prove type, call, flow or equivalence. Use hybrid retrieval: exact identifier/text search for precision, vector retrieval for vocabulary mismatch, then reranking and graph expansion within a pinned view. The BGE authors recommend hybrid retrieval plus reranking for BGE-M3. [BGE-M3 model card](https://huggingface.co/BAAI/bge-m3)

| Axis/class | Meaning | Role in valio.code |
|---|---|---|
| General multilingual text embedding | Training/task focus is ordinary multilingual retrieval | Baseline for issues, documentation and Russian/English query-to-code retrieval; it may miss identifier and syntax signals. |
| Code-specialized embedding | Trained/evaluated for code or text-to-code retrieval | Comparison candidate for code chunks and NL-to-code. Jina v2 base code declares 30 programming languages and Apache-2.0 licensing. [Model card](https://huggingface.co/jinaai/jina-embeddings-v2-base-code) |
| Deterministic structural features | Derived from syntax/semantic facts, not a neural vector | First choice for explainable structural similarity and graph-role ranking. |
| Graph/GNN representation | A model consumes a specified graph snapshot | Later ranking option after graph facts and relevance labels are reliable. |
| Contrastive training | Learning objective bringing positives together and negatives apart | Can train general-text, code-specialized or graph models; it is not a competing storage or retrieval modality. Text/code contrastive pre-training is used for code search. [Neelakantan et al.](https://cdn.openai.com/papers/Text_and_Code_Embeddings_by_Contrastive_Pre_Training.pdf) |

Ollama is an inference/runtime choice, not a model family. It suits a local baseline only when the exact model artifact is pinned and evaluated. BGE-M3 is a reasonable multilingual baseline to test: its published card reports 1,024 dimensions, 8,192-token inputs, dense/sparse/multi-vector modes and MIT licensing. That does **not** establish code-search quality for this repository. Ollama says vector width depends on the model, so collection schema must be probed, never assumed. [Ollama embeddings documentation](https://docs.ollama.com/capabilities/embeddings)

For every vector generation store: publisher/name and immutable revision/digest, license/terms review, runtime and quantization, tokenizer/context limit, native and stored dimensions, normalization, query/document prefixes, chunker and path/symbol context format, source-view/profile fingerprints, fusion/reranker version and access policy. A change to model, dimension, chunking or prefix creates a new generation and requires reindexing; never compare vectors from different spaces. Apply upload/redaction policy before considering a remote provider, and never put raw source in telemetry.

## Evaluation gates

1. **Truth fixtures:** language-specific positive, negative, ambiguous, incomplete-build and dynamic/framework cases; assert edge evidence and candidate status, not counts alone.
2. **Incrementality:** edit one file/package/profile and assert exactly which generations become stale and rebuild; cover deletion, rename and dependency-lock change.
3. **Query correctness:** exact/candidate precision, reference/call recall against compiler or curated oracle, slice agreement with hand-checked small programs, and no cross-view joins.
4. **Retrieval:** held-out repository/time split; separate exact identifier, NL-to-code, bug/issue, cross-lingual and structural-similarity sets; report Recall@K, MRR/nDCG, latency, token/context cost and citation coverage. Compare lexical, vector, hybrid, reranked and graph-expanded variants at the same budget.
5. **Operational bounds:** index/reindex time, storage per source line, traversal fan-out, peak memory, timeout rate, unresolved/candidate rates and a 10M-line benchmark before scale claims.

The release gate is not “a graph was stored” or “an embedding was returned.” A result must name its pinned view, explain the evidence path, expose unknowns and remain inside its declared budget.
