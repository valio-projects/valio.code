# Local bootstrap HTTP contract

`openapi.json` is generated from the actual Go request, response and rich type
models. From `src/back-end`, regenerate it with `go run ./api/openapi`.

The API uses one database-bound workspace (`VALIO_WORKSPACE_ID`, default
`workspace-main`). A caller explicitly registers repository IDs, then creates
project definitions and their repository source roots before uploading sanitized
agent snapshots. Project keys and identity are immutable; display names and roots
can change. Published views retain their original definitions and memberships.

The local stack defaults to the root SurrealDB user and password `valio` at
`http://127.0.0.1:18000`, namespace `valio`, database `workspace_main`, and
the development token `valio-local-development-token-0001`. `VALIO_DB_*` and
`VALIO_API_TOKEN` override these values. HTTP listens at `127.0.0.1:8090` unless
`VALIO_API_ADDRESS` is set. By default, browser commands from
`http://localhost:8080` and `http://127.0.0.1:8080` are trusted;
`VALIO_TRUSTED_ORIGINS` accepts a comma-separated replacement list. Bearer
authentication supports CLI clients; browser login sets
an expiring, signed, HttpOnly, SameSite=Strict cookie without persisting the
bootstrap token in browser storage. Cookie mutations require a trusted Origin.
Sessions expire after eight hours and reset when the API restarts. Logout clears
the browser cookie; it is not a server-side revocation service.

Only health/readiness and the credential-verifying session exchange are reachable
without an authenticated principal. `/mcp` uses the same bearer authorization
boundary and invokes the same application queries directly.

The current bounded slice limits upload JSON, combined source and expanded
analysis artifacts separately to 16 MiB, combined files to 10,000, catalog lists
to fewer than 1,000 entries, and search pages to 1,000 matches. Exceeding a scope
budget returns `SCOPE_TOO_LARGE`; search never silently caps candidates.

Go uploads produce AST syntax evidence. The build profile is `syntax-default`;
no server compiler, toolchain or filesystem reader runs. Rich facts that require
semantic resolution or ABI information remain unresolved, and unsupported
projections are explicit. Source text is ready; symbols/types are partial.
Views combining repositories report `mixed: true`. Project source fingerprints,
compiler fact import, durable search indexes and multi-user authorization are
not implemented by this slice.

Publication uses a single SurrealDB transaction for immutable snapshots, blobs,
files, artifacts, type descriptors, project revisions, 256 manifest shards and
the latest-view pointer. An immutable mismatch or concurrent head/configuration
change rolls back everything. A retry of an existing view returns its original
receipt without rewinding the latest pointer.

Validation includes HTTP authentication/Origin/strict JSON/body-limit tests,
unsafe snapshot rejection before any repository access, Fx dependency-graph
validation, and real SurrealDB tests for duplicates, mixed repositories,
root/name changes with old-view queries, rich Go types and staged-write rollback.
Set `VALIO_TEST_DB_URL` and `VALIO_TEST_DB_PASSWORD` to run storage integration
tests against an isolated temporary database.
