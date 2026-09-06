# valio.code

Software Intelligence Graph with explicit **Workspace → Project → source roots**,
immutable source views, evidence and honest partial-analysis states. Go backend and
Windows/Linux agent; React/Vite/HeroUI web; SurrealDB is the only persistent
application store. Licensed under AGPL-3.0.

This repository currently implements the first working vertical slice, **not the
complete v1 specification**. See [implementation status and backlog](docs/implementation-waves.md).

## Run locally

Requirements: Docker with Compose, Git; Go 1.26.8 for native agent development.

```powershell
./scripts/setup.ps1
docker compose up -d --build
```

On Linux, generate settings with `sh scripts/setup.sh` instead. Setup creates an
ignored `.env` with independent random credentials. Existing settings are retained.

Open http://127.0.0.1:8080. Sign in using `VALIO_API_TOKEN` from your local `.env`.
The browser exchanges it for an HttpOnly session cookie. Jaeger is available at
http://127.0.0.1:16686. API/worker use distinct database accounts; root credentials
are restricted to provisioning. The base Compose file exposes no database port.

Create a repository and a Project with its source roots in the UI, then ingest it:

```powershell
# Set VALIO_API_TOKEN in this process from your local settings first.
go run ./cmd/valio-agent index --root C:/path/to/repository --server http://127.0.0.1:8080 --repository repo-id
go run ./cmd/valio-agent watch --root C:/path/to/repository --server http://127.0.0.1:8080 --repository repo-id
```

`watch` combines fsnotify with reconciliation and upload retries. `index` without
`--server` emits a sanitized snapshot. Spool data remains available for retry.
Tracked files honor explicit policy; `.env` values are excluded before spooling.
Some configuration formats are conservatively excluded until safe adapters exist.
The basic secret detector for ordinary source is heuristic, not a universal secret
classifier; apply explicit source exclusions to confidential material.

For an isolated demonstration, run `./scripts/smoke.ps1`: it creates a tiny Git
fixture under ignored `.valio/`, registers a demo Project, uploads through the
native agent, checks search/type lookup, redaction and duplicate delivery.

## Implemented operations

- Project/source-root registration, shared and multi-repository membership, immutable
  view pinning and transactional publication in SurrealDB.
- Verified boolean exact/substring/RE2 search with project scope, Unicode folding,
  conservative trigram candidates and explicit scan/highlight limits.
- Go syntax extraction, rich type lookup: fields, methods, parameters, generic
  constraints, struct tags, named underlying types and constants. An optional local
  Go compiler provider adds supported semantic facts; server ingestion uses syntax.
- Language-neutral type metadata contracts cover visibility, modifiers, attributes,
  enums and ABI-dependent layout. Unavailable values remain explicitly unknown.
- Durable fenced job queue, bounded worker concurrency, Fx dependency injection,
  OpenTelemetry HTTP/storage/job spans and Jaeger OTLP export.
- Authenticated HTTP, Streamable HTTP MCP and native stdio bridge. Implemented MCP
  tools: `project_list`, `project_get`, `code_search`, `type_query`, `index_status`.

```powershell
go run ./cmd/valio-mcp serve --server http://127.0.0.1:8080
```

The bridge reads `VALIO_API_TOKEN`; it has no database credentials. See the generated
[OpenAPI contract](api/openapi/openapi.json), [type metadata](docs/type-information.md),
[architecture](docs/architecture.md) and [observability](docs/observability.md).

## Validation

```powershell
go vet ./...
go test ./...
./scripts/test-integration.ps1
docker build --target test -t valio-code-test .
cd web
npm ci
npm test
npm run build
```

The integration script temporarily uses `deploy/compose.test.yaml` to expose
SurrealDB on loopback 18000 and OTLP on 14318. Return to the base Compose topology
with `docker compose -f compose.yaml -f deploy/compose.test.yaml down` followed by
`docker compose up -d`; volumes are retained. Never use `down -v` to preserve data.

## Current boundaries

The bootstrap deployment supports one configured Workspace and a shared local
admin token. Multi-user roles, token lifecycle, production TLS configuration,
backup/restore tooling and multi-tenant control-plane provisioning remain backlog.

Persistent FTS/HNSW, five-language SCIP/compiler ingestion, call/CFG/interprocedural
flow, API/SQL/event linking, test reports, history analytics, structural/graph-role
retrieval, embeddings and context generation are not complete. Current search
rebuilds in-memory candidates from persisted source and has bounded scan sizes;
10-million-line performance is not established. The worker currently registers a
health processor; analysis imports are synchronous. Jaeger uses ephemeral storage.
The backlog records these gaps rather than advertising unsupported tools.
