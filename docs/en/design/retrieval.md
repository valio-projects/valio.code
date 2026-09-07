# Retrieval design: current state and proposed expansion

## Status and scope

This is a **design proposal**, not an implemented feature. It distinguishes
what the current backend proves from a possible retrieval system. It does not
promise a universal quality multiplier, a fixed best model, or a fixed number
of results for every repository.

The proposal is constrained by the existing source policy: only sanitized,
immutable view inputs may be indexed; the system must keep the selected
workspace, immutable view, project, repository, and build-profile scope fixed
through retrieval and context assembly. A retrieval result is evidence, not a
claim that a model has established a program fact.

## Audit of the current backend

| Area | Present behavior | Important limit |
|---|---|---|
| Search unit | The search package scans complete files from the selected immutable view, verifies every positive match against original content, then pages. | It is not passage or semantic retrieval. |
| Candidate index | `BlockBytes` is 16 KiB, with two-rune overlap and folded trigrams. Blocks are source-byte-aligned filtering aids. | A 16 KiB block is **not** a semantic chunk and has no declaration, function, or parent-child meaning. |
| Exactness | `exact` compares an entire selected field; default content/path/symbol matching is verified substring matching; regex is RE2. | BM25 or vectors cannot replace this exact/substring contract. |
| Parser | Go uses `go/parser` and `go/ast`; its output records syntax ranges, declarations, imports, occurrences, and syntax diagnostics. | Tree-sitter is not currently used. |
| Type information | Optional `go/types` receives only the one parsed file. The mapper makes exact identities only for local declarations. | Imports, cross-file declarations, and compiler/package resolution remain unresolved; `PaymentService` cannot be followed through the repository. |
| Project/version scope | Manifests and immutable views pin content and project membership; type lookup preserves ambiguity. | There is no global import/export symbol map or graph traversal. |
| Projection contracts | `vector` and `context` are declared projection families with dependencies. | They are contracts only; no embeddings, BM25 index, reranker, or context builder exists. |

The evidence is in [the search README](../../../src/back-end/internal/search/README.md),
[the trigram index](../../../src/back-end/internal/search/index.go),
[the Go parser](../../../src/back-end/internal/analysis/go_parser.go),
[the single-file checker](../../../src/back-end/internal/analysis/type_checker.go),
and [the projection registry](../../../src/back-end/internal/projections/registry.go).

## Proposed retrieval records

### 1. Context-aware syntax chunks

Introduce a language-neutral `SyntaxChunker` port with language adapters. For
Go, the first adapter should use the existing Go AST; Tree-sitter is a future
option for languages without a suitable compiler parser or for tolerant,
incremental parsing. Tree-sitter is a parser generator and incremental parser,
and it exposes concrete-tree byte ranges, but adopting it would be an explicit
new dependency and grammar/version decision, not a description of the current
implementation. [Tree-sitter’s documentation](https://tree-sitter.github.io/)
describes its incremental concrete syntax trees and byte-oriented positions.

Each chunk should record:

* immutable `viewId`, workspace, project, repository, file, language, and
  build-profile identifiers;
* a stable `chunkId`, `parentId` where applicable, ordinal, kind, and original
  half-open UTF-8 byte range;
* the content digest and chunker/grammar/profile versions used to create it;
* declaration identity when it is known, plus an explicit unresolved or
  ambiguous state when it is not; and
* a generated retrieval representation separate from the original source.

Chunk at useful syntax boundaries: file/package, type, function or method,
and selected nested bodies. A small function can be one chunk. A function that
exceeds the representation budget must split on AST body boundaries, retain a
shared `parentId`, preserve an ordered sequence, and carry only a bounded
signature/header overlap. The child must never silently imply that it contains
the whole parent. Boundary and overlap sizes are parameters to calibrate on
fixtures; they are not product invariants.

Header injection may add a generated header such as language, qualified
declaration, signature, package/path context, and parent identity to an
embedding or reranking representation. It must be stored and labelled as
`generated-representation`, with a version and the original byte range. It is
not source text, does not move source coordinates, and does not make a model
output trustworthy. Displayed source, highlights, and evidence continue to use
the original immutable bytes and their recorded ranges.

### 2. Eight representation profiles, including source code

Do not make summary text the only vector input. Code identifiers, literals
that policy permits, API spellings, and uncommon error text are often the
reasons a developer searches. A summary can omit those details, drift from the
versioned source, and adds generation cost and invalidation work.

Use independently versioned profiles alongside canonical source code. The
following eight are a starting schema, not a claim that every profile will be
enabled in every deployment:

| Profile | Input | Purpose and caution |
|---|---|---|
| `code-body/v1` | Sanitized canonical chunk source | Preserves identifiers and syntax; still subject to capture policy. |
| `declaration/v1` | Name, qualified name, signature, kind | Strong retrieval surface for symbol and API queries. |
| `documentation/v1` | Attached comments/docstrings | Useful for natural-language intent; may be absent or stale. |
| `header-context/v1` | Generated bounded header plus body | Adds scope without pretending the header was source. |
| `identifier-lexicon/v1` | Tokenized identifiers and aliases | Helps rare names; does not replace exact matching. |
| `parent-summary/v1` | Explicitly generated, versioned summary | Optional retrieval aid, marked generated and never sole evidence. |
| `neighborhood/v1` | Bounded resolved relations and selected signatures | Available only when graph resolution is ready; no whole-class expansion. |
| `path-metadata/v1` | Repository-relative path, package/module and language | Helps navigation; must not widen access scope. |

Every profile fingerprint includes source content, profile schema, tokenizer,
embedding model/version, prompt/header template where used, and upstream graph
versions. A source change invalidates its children; a profile/model change
invalidates only that profile and downstream vectors. This fits the existing
projection fingerprint principle rather than treating vectors as timeless data.

### 3. Import/export and symbol graph

Build a versioned, build-profile-specific import/export map before promising
cross-file navigation. Each adapter should emit declaration/export records,
imports, references, and resolution outcomes: `exact`, `candidate`,
`unresolved`, or `unsupported`. Resolution must retain the chosen scope and
its evidence; equal names in two packages must remain distinct candidates.

Only after that graph exists can a query such as “follow `PaymentService`”
resolve its declaration, imports/re-exports, references, implementations, or
selected members. Current import strings and local Go type facts are useful
inputs, but are not such a graph. Compiler-backed resolution needs the
appropriate package/module/build inputs; it must not be inferred from spelling.

Context expansion starts from an already selected chunk or resolved symbol. It
is bounded by a fixed scope, depth, bytes/tokens, node count, and selected
member policy. It should add a declaration signature, directly relevant call
or reference edge, and only the member/body selected by the query or reranker.
It must not expand an entire large class, package, or repository merely because
one symbol matched.

## Proposed retrieval pipeline

Keep three independently observable channels:

1. **Verified exact channel.** The existing exact/substring/regex search stays
   authoritative for literal identifiers, paths, and source spans. It returns
   verified matches even if semantic retrieval is unavailable.
2. **Lexical ranked channel.** Add a persisted, view-scoped BM25-style index
   over approved chunk/profile fields. BM25 ranks lexical evidence; it does not
   provide the current exact-substring guarantee.
3. **Dense ranked channel.** Add embeddings for selected profiles, with model,
   tokenizer, dimension, distance, and profile versions pinned in the record.
   Dense vectors are useful for paraphrase and intent, but are not universally
   bad at identifiers and are not universally good at code. Measure both query
   classes separately.

Fuse lexical and dense *ranks*, rather than adding incomparable raw scores.
Reciprocal Rank Fusion is a reasonable initial baseline because it combines
ranked lists without assuming score calibration; its constant and channel
depths still need offline tuning. The original RRF paper evaluates a fusion
method over ranked systems, not this product or its corpus. [Cormack, Clarke,
and Büttcher (SIGIR 2009)](https://doi.org/10.1145/1571941.1572114).

A cross-encoder reranker can score the query with a bounded candidate
representation after fusion. Cross encoders typically score query/document
pairs with higher quality than a bi-encoder but at a per-pair cost, which is
why retrieve-then-rerank is common. [Sentence Transformers documents this
trade-off](https://www.sbert.net/docs/cross_encoder/usage/usage.html).

Start experimental calibration with up to 50 fused candidates and return up to
5 reranked chunks only when that fits the caller’s context budget. These are
not permanent `top 50 → 5` rules: query type, candidate recall, corpus size,
latency, and an approximately 8k-token downstream context budget decide the
actual thresholds. The API should report candidate count, channel membership,
profile/model versions, truncation, and whether reranking was skipped.

For RU/EN queries and code context, evaluate language coverage, code-query
quality, context limit, deployment constraints, cost, and license before
selection. Cohere’s current rerank documentation states multilingual support
but also states that combined query/document tokens count against a model
context limit; that is a vendor capability, not a code-search benchmark.
[Cohere rerank details](https://docs.cohere.com/docs/rerank) and
[language support](https://docs.cohere.com/docs/rerank-overview) are useful
operational references. A local candidate is BAAI’s multilingual
`bge-reranker-v2-m3`; its model card identifies it as a reranker and lists an
Apache-2.0 license, which must still receive project legal and deployment
review. [BAAI model card](https://huggingface.co/BAAI/bge-reranker-v2-m3).

Neither provider should receive source outside the existing capture policy or
the caller’s approved boundary. A provider outage, rejected content, exceeded
context, or incompatible license leaves exact and lexical results available and
reports the semantic channel as unavailable or partial.

Dense retrieval has demonstrated value in its own evaluated setting, but that
does not transfer as a guaranteed result for this repository. DPR reported
improvements over a BM25 baseline for open-domain QA, while CodeSearchNet was
created precisely because natural-language-to-code retrieval has a distinct
semantic gap and requires code-specific evaluation. [DPR](https://arxiv.org/abs/2004.04906)
and [CodeSearchNet](https://arxiv.org/abs/1909.09436) support evaluating rather
than assuming this design.

## Implementation and evaluation plan

| Priority | Needed capability | Dependencies and invalidation | Acceptance fixtures and measurements |
|---|---|---|---|
| P0 | Retrieval record, scope guard, profile registry, provenance response | Immutable views, policy filter, projection fingerprints | Cross-workspace denial; source range unchanged after header injection; profile/version shown in result. |
| P1 | AST chunker and source/header representations | Language adapter; source digest; parent/child records | Empty/error-tolerant file, nested declarations, oversized function split, Unicode byte ranges, signature overlap bound. |
| P2 | Persisted exact and BM25 channels | View-scoped postings and deterministic analyzer | Exact rare identifier/path recall is 100% within scanned scope; BM25 does not claim exactness; latency and index-size measurements. |
| P3 | Import/export map and selected graph expansion | Per-language resolver, build profile, symbol/reference projection | Two `PaymentService` names, aliases, re-exports, unresolved import, version change; no whole-class expansion. |
| P4 | Dense profiles and vector index | P0–P3, chosen model/tokenizer, operational and license review | RU/EN natural-language and identifier sets, code-only queries, stale-summary detection, recall@K/NDCG@K, embedding cost and rebuild time. |
| P5 | Fusion and reranking | Candidate telemetry, bounded representations, provider/local adapter | RRF parameter sweep; recall before rerank; MRR/NDCG@5, p50/p95 latency, rejected/oversize documents, per-query cost. |

Create a reviewed relevance set from real, sanitized repository questions,
stratified by language (RU/EN), query type (identifier, exact phrase, natural
language, path, symbol-following), project size, and ambiguous names. Freeze
the view and expected evidence for each judgement. Report Recall@K before
reranking, MRR and NDCG@5 after ranking, exact-channel coverage, resolution
coverage/ambiguity, p50/p95 latency, bytes/tokens sent to each provider, index
size, rebuild duration, and monetary cost per indexed byte and query. Compare
against the current verified search baseline; do not publish an aggregate
“X-times better” claim without the fixture, corpus, version, and confidence
interval that support it.

## Decision summary

The next safe implementation step is P0/P1: durable, scoped syntax chunks with
original coordinates and versioned representations. BM25, vectors, global
symbol resolution, and reranking depend on those records and on evaluation
fixtures. The existing verified search remains a required exact evidence path
through every phase.
