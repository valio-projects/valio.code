# Initial vertical-slice validation — 2026-09-07

Validated locally on Windows with Go 1.26.8 and Node 26.8.1, and in Docker Desktop
Linux containers. These results establish fixture-level behavior only.

| Check | Result |
|---|---|
| `go test ./...`, `go vet ./...` | Passed |
| Linux Docker `go test -race ./...` | Passed; external integrations run separately |
| Real Compose SurrealDB 3.2.4 | Provisioning, database isolation, immutable writes, job leases/fencing and publication passed |
| Real application persistence | Duplicate ingestion, pinned old membership, project rename/root changes, multi-repository view and atomic rollback passed |
| Real Jaeger OTLP export | Span exported and retrieved by trace ID |
| MCP official SDK roundtrip | Tool discovery and Project invocation passed; redirects blocked |
| Actual nginx → API → MCP | Initialize and authenticated `project_list` passed |
| Native agent → nginx → API → SurrealDB | Demo source upload, search and rich Go type lookup passed |
| Configuration fixture | `.env` value excluded from searchable source; sanitized duplicate upload preserved view identity |
| Container recreation | Existing pinned view and demo Project remained readable after Compose down/up without removing volumes |
| Base network | Only web 127.0.0.1:8080 and Jaeger 127.0.0.1:16686 published; no database/worker host port |
| Web clean install/build | Passed; lockfile retained |
| Web Vitest | 12 tests passed: auth, API errors, scope, source ranges, form validation, metadata/evidence |
| Browser | Login layout at 1280×720 and rejected-token clearing checked |

Authenticated UI navigation was not exercised end-to-end in the browser. Its
application operations were exercised through real HTTP, and rendering helpers
through component tests. Full graph capabilities, five-language extraction,
10-million-line benchmarks, production authentication and backup restoration are
not covered by these results. Refer to the backlog before treating a feature as
release-ready.

Reproduce the container integration with `scripts/test-integration.ps1`, and the
native-agent demonstration with `scripts/smoke.ps1`. The latter retains its small
demo Project and ignored source fixture for inspection. No user repository was
uploaded as part of this smoke scenario.
