# Measured compatibility baseline

Validated locally on Windows using Go 1.26.8 and SurrealDB 3.2.4 in Docker Compose.
SurrealDB is not started as a native Windows process. The current local Compose file
uses inline development values; API, worker and migration authenticate as root.

Verified mechanisms:

- JSON RPC bound query parameters, including source strings containing SQL delimiters.
- Typed CBOR repository records through `github.com/surrealdb/surrealdb.go` 1.6.0
  and `github.com/fxamacker/cbor/v2` 2.9.3, including source-like, UUID, date and
  Unicode strings plus 64-bit integers.
- Explicit namespace/database creation (3.2 does not implicitly create missing scopes).
- Database-scoped system accounts and rejection of cross-database authentication.
- Immutable create collision, stable payload retrieval and versioned migration.
- Idempotent job enqueue, exclusive lease, heartbeat, expired lease reclaim, increasing
  fencing token, stale completion rejection and transactional generation publication.
- Compose named volume initialized once for non-root UID 65532; database health check.

Compatibility findings requiring explicit design choices:

1. `DEFINE USER ... PASSWORD` in 3.2.4 requires a string literal. Provisioning serializes
   that literal with JSON escaping. It never logs SQL or raw database error strings.
2. The compound indexed OR predicate used to select queued/expired jobs returned no
   candidate for an eligible queued job in the initial integration fixture. The baseline
   claim query uses `WITH NOINDEX` for correctness. Queue scale must be benchmarked before
   claiming production throughput; index-backed selection remains an open compatibility task.
3. Lease predicates are checked on a selected record before mutation in the same transaction.
   No queue correctness assumption depends on UPDATE/WHERE evaluation of modified fields.
4. Docker Desktop's internal-only network did not expose the test database port. The test
   override makes that network non-internal and publishes only `127.0.0.1:18000`. The base
   deployment keeps the database unpublished on its internal network.
5. SurrealDB 3.2.4 JSON RPC applies legacy coercion to a value such as
   `id: number`, yielding the record ID `id:number`. Typed CBOR records preserve
   the caller's value and are the required repository codec for this boundary.

Still unverified: FTS/HNSW physical indexes, concurrent initial index creation, SQL parser
native dependencies, five SCIP compiler adapters, 10M-line throughput, storage amplification,
consistent backup/restore and the complete v1 failure-injection matrix.

Reproduce with `scripts/test-integration.ps1`. It starts its own
`valio-code-test` Compose project with the test override, then removes that project
without volumes; the base local stack does not need to be stopped. Equivalent Go
tests require `VALIO_TEST_DB_URL` and `VALIO_TEST_DB_PASSWORD` pointing to the
isolated Compose instance.

References: [SurrealDB HTTP/RPC transport](https://surrealdb.com/docs/reference/rest-api/http-protocol),
[UPDATE semantics](https://surrealdb.com/docs/reference/query-language/statements/update).
