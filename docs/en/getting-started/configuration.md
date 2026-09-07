# Configuration

The checked-in Compose file is self-contained for local development: no `.env`,
`setup.ps1`, or `setup.sh` is required. Run it from the repository root with:

```powershell
docker compose up -d --build --wait
```

SurrealDB has one local root account, `root`, with password `valio`; API, worker,
and migration all authenticate at root level through this account. Database traffic
stays on the internal Compose network. The API process is reachable through the
web proxy at `127.0.0.1:8080`; Jaeger is at `127.0.0.1:16686`.

The local API bootstrap token is embedded in Compose as
`valio-local-development-token-0001`. Native CLI commands default to the same
token. Browser sign-in exchanges the token for an HttpOnly cookie; do not put it
in a URL or browser storage.

An existing named volume is retained when the stack is recreated. A fresh checkout
needs no credential migration. Before recreating an older stack that has an existing
volume, rotate its root credential while the old `surrealdb` container is still
running:

```powershell
"DEFINE USER OVERWRITE root ON ROOT PASSWORD 'valio' ROLES OWNER;" | docker compose exec -T surrealdb /surreal sql --endpoint http://localhost:8000 --username root --auth-level root --hide-welcome --json
```

This preserves the stored data. If the old container is already unavailable, use
its known old root password to perform the same rotation before starting the new
configuration. An old `.env` is no longer required; the local values are defined
directly in Compose. Do not use
`docker compose down -v` when that data is needed.

The API supports `VALIO_WORKSPACE_ID`, `VALIO_API_ADDRESS`, and
`VALIO_TRUSTED_ORIGINS` for native or deployment-specific runs. Its data-store
settings are `VALIO_DB_URL`, `VALIO_DB_NAMESPACE`, `VALIO_DB_DATABASE`,
`VALIO_DB_USER`, `VALIO_DB_PASSWORD`, and `VALIO_DB_AUTH_LEVEL`. See the
[OpenAPI contract](../../../src/back-end/api/openapi/openapi.json) for HTTP limits
and session behavior.

For a native API process, `VALIO_API_ADDRESS` defaults to `127.0.0.1:8090`.
The local ingress remains `http://127.0.0.1:8080`, so native agents normally use
the ingress URL unless deliberately testing the API listener. On Windows, set a
process-only override with `$env:VALIO_API_TOKEN = '...'`; on Linux use
`export VALIO_API_TOKEN='...'`. These commands do not rewrite Compose or an
existing volume.

The agent currently reads only `VALIO_API_TOKEN`, not an agent configuration file.
The proposed base server URL, spool/watch settings, and Project profiles are
documented in [local agent configuration](../design/agent-configuration.md) and
are not implemented.

Configuration values and source-policy exclusions are applied before agent spool,
hashing, and upload. The current secret detector is heuristic; use explicit source
exclusions for confidential material.
