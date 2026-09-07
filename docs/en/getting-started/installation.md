# Install and run

Choose the tools for the work you are doing:

- [Docker Desktop with Compose](https://docs.docker.com/compose/install/) is
  required to run the local stack and the isolated integration topology (Compose
  2.24.4 or newer).
- [Git](https://git-scm.com/downloads) is required to clone the repository and for
  agent worktree capture.
- [Go 1.26.8](https://go.dev/doc/install) is required for native-agent and backend
  development.
- [Node.js 26.8.1](https://nodejs.org/en/download) is required for frontend
  development.

A clean checkout does not contain `.tools/go`. Check the normally installed Go
toolchain first; use the optional portable toolchain only after it has separately
been provisioned:

```powershell
docker compose version
go version
node --version
```

On Windows, a provisioned portable toolchain is
`& .\.tools\go\bin\go.exe version`; on Linux it is
`./.tools/go/bin/go version`. These are optional alternatives to `go version`.

From the repository root, start the local stack:

```powershell
docker compose up -d --build --wait
```

Open the web app at <http://127.0.0.1:8080> and Jaeger at
<http://127.0.0.1:16686>. SurrealDB is internal to Compose and has no base-stack
host port.

The local login token is `valio-local-development-token-0001`. The browser
exchanges it for an HttpOnly session cookie. It is a local bootstrap credential,
not a production identity system.

For native backend development:

```powershell
# Run this block from the repository root.
cd src/back-end
go test ./...
go vet ./...
go run ./api/openapi
```

For frontend development and verification:

```powershell
# Run this block from the repository root.
cd src/front-end
npm ci
npm test
npm run build
```

Build the backend test image from the repository root:

```powershell
docker build -f src/back-end/Dockerfile --target test -t valio-code-test .
```

Run the integration and native-agent smoke scenarios from the repository root:

```powershell
.\scripts\test-integration.ps1
.\scripts\smoke.ps1
```

The integration script starts only its isolated database and Jaeger project, then
removes that project in `finally`; it does not stop the base local stack. It requires
Docker Compose 2.24.4 or later for the test port override:

```powershell
docker compose -p valio-code-test -f compose.yaml -f deploy/compose.test.yaml up -d --wait surrealdb jaeger
```

The isolated endpoints are SurrealDB `127.0.0.1:18000`, OTLP
`127.0.0.1:14318`, and Jaeger UI `127.0.0.1:16687`. The base Jaeger UI remains
at `127.0.0.1:16686`.

To ingest a repository during native development, run the agent from
`src/back-end`. Its local default token matches the Compose token:

```powershell
go run ./cmd/valio-agent index --root C:/path/to/repository --server http://127.0.0.1:8080 --repository repo-id
go run ./cmd/valio-agent watch --root C:/path/to/repository --server http://127.0.0.1:8080 --repository repo-id
```

`index` without `--server` emits a sanitized snapshot. `watch` reconciles file
changes and retries uploads using its spool.

## First Project and repository upload

The current CLI cannot create a Workspace, Repository, Project, or source-root
attachment. In the web app, create the Repository and Project, attach its source
roots, and copy the returned repository ID. Then run the supported agent flags
from `src/back-end`:

```powershell
go run ./cmd/valio-agent index --root C:/src/users-api --server http://127.0.0.1:8080 --workspace workspace-main --repository repo-users
go run ./cmd/valio-agent watch --root C:/src/users-api --server http://127.0.0.1:8080 --workspace workspace-main --repository repo-users --interval 5s
```

On Linux, an absolute root can be `/work/users-api`. `--server` enables upload;
without it the agent emits a local sanitized snapshot. `--spool` defaults to
`<root>/.valio/spool`, and `--max-file-bytes` is capped at 2 MiB. Set a different
local token only when needed:

```powershell
$env:VALIO_API_TOKEN = 'valio-local-development-token-0002'
```

```bash
export VALIO_API_TOKEN='valio-local-development-token-0002'
```

Use that override only with an API/server configured with the same token. For a
Compose change, update the API token value and run
`docker compose up -d --force-recreate api` before sending the request. The base
Compose token already matches the agent default, so a normal local stack needs no
override. For a future base-file/profile workflow, see
[local agent configuration](../design/agent-configuration.md); `--config` and
`--profile` are not valid current flags.

## Verify and troubleshoot

From the repository root, inspect the service that is failing:

```powershell
docker compose ps
docker compose logs --tail 100 api worker
Invoke-WebRequest http://127.0.0.1:8080/health
Invoke-WebRequest http://127.0.0.1:8080/readyz
```

On Linux, use `curl -fsS http://127.0.0.1:8080/health` and
`curl -fsS http://127.0.0.1:8080/readyz`. `/health` verifies nginx; `/readyz`
is proxied to the API. If an upload fails, inspect the agent's JSON diagnostic and
retry from the same spool; do not copy configuration values into commands or logs.
