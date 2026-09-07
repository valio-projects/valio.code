# Language analysis

Valio indexes C, C++, C#, Java, JavaScript, and TypeScript as source files and
adds a syntax report for each supported file when the syntax helper is enabled.
The report is an AST observation, not a compiler result. It makes written
declarations useful across repositories while keeping unresolved information
explicit.

## What happens during indexing

1. `valio-agent index` captures a Git worktree, applies the source and
   configuration policy, and uploads the accepted snapshot.
2. The API assigns immutable file and view identities. Supported source files
   are sent to the bundled Tree-sitter WebAssembly helper.
3. The report is validated against the submitted path and original UTF-8 byte
   length, then stored with the immutable artifact in SurrealDB.
4. A pinned view serves ordinary source and symbol search, plus the dedicated
   syntax query. The query returns the report that was stored for that view;
   it does not parse a developer's current working tree.

The source-search DSL remains useful for exact source and symbol discovery. For
example, `symbol:User` searches written symbols in the selected immutable view.
Use the syntax query when the caller needs a declaration outline or written
member metadata.

## Supported paths

| Extension | Indexed language | Syntax grammar |
| --- | --- | --- |
| `.c`, `.h` | C | C |
| `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.hh` | C++ | C++ |
| `.cs` | C# | C# |
| `.java` | Java | Java |
| `.js`, `.mjs`, `.cjs`, `.jsx` | JavaScript | JavaScript; JSX for `.jsx` |
| `.ts`, `.mts`, `.cts`, `.tsx` | TypeScript | TypeScript; TSX for `.tsx` |

`.h` deliberately defaults to C. A C++ header is ambiguous without a compiler
configuration, so Valio does not guess that it is C++. Other extensions remain
source text with the `text` language fallback and receive no syntax report.

## Syntax report evidence

Reports use original, half-open UTF-8 byte ranges. They include classes,
interfaces, structs, enums and enum members, fields, properties, functions,
methods, their parent relationship, declared types where written, parameters,
visibility, modifiers, and direct attributes. Attribute lists preserve their
written literal syntax, including C# attribute targets and arguments. C# enums
also expose `underlyingType` only when the declaration explicitly writes a base
type such as `byte`.

Identifier observations and syntactic calls have `resolution: "unresolved"`.
C# `using` directives and TypeScript import statements are retained as
syntax-only imports. This is useful navigation evidence, but it is not a claim
that an import, call, type name, or enum use resolves to one particular symbol.

The report does not infer layout, offsets, alignment, overload selection,
inheritance, aliases, cross-file symbols, call targets, exact enum usages, or
control-flow/data-flow graphs. An absent field is unknown rather than evidence
that it does not exist.

## Querying a pinned view

From the repository root, create or refresh a view with the native agent:

First register `example-repository` and `example-project` (with its source root)
in the running API or web interface. The path must be an existing Git worktree.

```powershell
$env:VALIO_API_TOKEN = 'valio-local-development-token-0001'
go -C src/back-end run ./cmd/valio-agent index `
  --root C:\work\example `
  --server http://127.0.0.1:8080 `
  --repository example-repository
```

Keep the returned `viewId` and use it in every read. This PowerShell request
finds files whose stored syntax report contains the written name `User`:

```powershell
$headers = @{ Authorization = 'Bearer valio-local-development-token-0001' }
$body = @{
  scope = @{ workspaceId = 'workspace-main'; projectIds = @('example-project'); viewId = 'pinned-view-id' }
  name = 'User'
} | ConvertTo-Json -Depth 6
Invoke-RestMethod -Method Post -Uri http://127.0.0.1:8080/api/v1/syntax/query `
  -Headers $headers -ContentType 'application/json' -Body $body
```

The MCP tool is `syntax_query`. Supply the same `scope` and optional `name` or
`fileId`; for example, a client sends
`{"scope":{"workspaceId":"workspace-main","projectIds":["example-project"],"viewId":"pinned-view-id"},"name":"User"}`.
The legacy `GET /api/v1/types` endpoint remains Go-only. It is not a fallback
for C, C++, C#, JavaScript, or TypeScript type resolution.

## Enabling the native helper

Docker Compose includes the Node runtime and helper dependencies. For a native
API process, install Node.js 26.8.1, then prepare the helper before starting
the API:

```powershell
Push-Location src/back-end/analyzers/syntax
npm ci
npm test
Pop-Location
$env:VALIO_SYNTAX_HELPER = (Resolve-Path src/back-end/analyzers/syntax/index.js).Path
$env:VALIO_NODE_BINARY = 'node'
go -C src/back-end run ./cmd/valio-api
```

`VALIO_SYNTAX_HELPER` is read at API startup. The startup probe loads the
TypeScript grammar, so a missing Node executable, incompatible grammar asset,
or unavailable parser stops startup. During ingestion, a helper failure stops
publication rather than creating a view that pretends to contain syntax
evidence. Parse diagnostics from invalid source remain explicit in a stored
partial report.

Views created before syntax analysis was enabled have no retroactive reports.
Index the source again to create a new view after enabling the helper.

## Acceptance checks

Run the repository checks from the root:

```powershell
.\scripts\test-integration.ps1
.\scripts\smoke.ps1
```

The integration script installs the helper dependencies and runs its grammar
fixtures. The smoke scenario uploads a small multi-language fixture and checks
source and symbol search, syntax reports, retrieval/context, and local Go graph
evidence against its returned pinned view.
