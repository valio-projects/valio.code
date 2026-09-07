# Contributing

Use Go 1.26.8, Node.js 26.8.1, Git, and Docker with Compose. Read the
[contributor guide](docs/en/contributing/development-rules.md),
[architecture](docs/en/contributing/architecture.md), and
[observability rules](docs/en/contributing/observability.md).

Run backend checks from `src/back-end`:

```powershell
go vet ./...
go test ./...
go run ./api/openapi
```

Run frontend checks from `src/front-end`:

```powershell
npm ci
npm test
npm run build
```

For storage or Compose changes, run the relevant integration checks from the
repository root, including `.\scripts\test-integration.ps1` and
`docker build -f src/back-end/Dockerfile --target test -t valio-code-test .`.
The integration script uses `docker compose -p valio-code-test -f compose.yaml -f
deploy/compose.test.yaml up -d --wait surrealdb jaeger`, then tears down only the
isolated project without volumes. It needs Docker Compose 2.24.4 or later for the
port override and leaves the base local stack untouched.
Format Go with `gofmt`. Preserve AGPL-3.0 notices and document external
dependencies.

Do not add private credentials, sensitive configuration artifacts, toolchains,
corpora, or runtime databases to Git. The public local example credentials in
Compose are intentional. Fixtures must be small, synthetic, and independently
checked. Report unsupported analyzers as unsupported, not ready. Each finished
development stage or wave needs validation and a descriptive conventional commit
recording changed behavior, evidence, limitations, and checks.
