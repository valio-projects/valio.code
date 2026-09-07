# Contributor guide and development rules

Read the repository [contribution policy](../../../CONTRIBUTING.md) and
[`AGENTS.md`](../../../AGENTS.md) before changing code. Keep the AGPL-3.0 license,
the remote, and unrelated user changes intact.

Work from the component that you change:

```powershell
# Run this block from the repository root.
cd src/back-end
go test ./...
go vet ./...
go run ./api/openapi
```

```powershell
# Run this block from the repository root.
cd src/front-end
npm ci
npm test
npm run build
```

Use `gofmt` on Go changes. For storage or Compose changes, run the relevant
integration checks from the repository root:

```powershell
.\scripts\test-integration.ps1
docker build -f src/back-end/Dockerfile --target test -t valio-code-test .
```

`test-integration.ps1` isolates its services with
`docker compose -p valio-code-test -f compose.yaml -f deploy/compose.test.yaml up -d --wait surrealdb jaeger`
and tears down that project without volumes. Its database, OTLP, and Jaeger UI
ports are 18000, 14318, and 16687; it leaves the base Jaeger UI on 16686. Docker
Compose 2.24.4 or later is required for the port override.

Run root-level orchestration from the repository root and module-level checks
from their component directory; do not rely on the current shell having changed
directories earlier in a session. On Windows the bundled toolchain can be invoked
as `& .\.tools\go\bin\go.exe -C src\back-end test ./...`; on Linux use
`./.tools/go/bin/go -C src/back-end test ./...`. The frontend equivalent is
`cd src/front-end; npm ci; npm test; npm run build` on either platform.

Keep Workspace, Project, Repository, Worktree, and Deployment as separate
concepts. SurrealDB is the only persistent application store. Domain code stays
independent of HTTP, database SDKs, compiler SDKs, and dependency injection;
repository interfaces are implemented by infrastructure adapters. Services do not
construct SurrealQL. Operational logs must not contain source or raw configuration
payloads. Policy-accepted, sanitized source files may be captured, uploaded, and
spooled for retry; configuration values and excluded paths must never enter them.

State what is implemented, tested, and still unknown. Do not present AST heuristics
as compiler facts or unknown metadata as zero. Record capabilities, validation,
and limitations in the appropriate documentation after each completed wave.
