# Web workbench

The `web/` application is a React 19, TypeScript, Vite 8, HeroUI 3 and Tailwind CSS 4
workbench. The feature folders separate workspace, projects, authentication, search,
source viewing, capabilities and type inspection. Shared API adapters, domain models,
providers and shell components have focused responsibilities. It contains no demo
dataset, fallback records or seeded counts.

## Run and build

From `web/`, run `npm ci`, `npm test`, and `npm run build`. Run `npm run dev` for Vite
development; `/api` proxies to an API listening at `127.0.0.1:8090`. Dependencies are
exactly pinned with the generated npm lockfile. Node 24 is used in the build image.

Compose builds `web/Dockerfile` with the repository root as context. Its final nginx
image runs as UID 101 on port 8080. Node and nginx base images are pinned to verified
multi-platform image digests. nginx serves the SPA and forwards `/api/`, `/mcp`,
`/healthz`, and `/readyz` to `api:8090`. `/health` checks the static web server.
Access logging is disabled so query text, type names and URL scopes are not recorded
in nginx request logs. The proxy preserves the original Host including the port for
same-origin session validation. The image supplies CSP, MIME and framing headers.

## Implemented workflows

- Workspace overview shows returned catalog counts and selected view metadata.
  Without a view it shows repository/project setup and an actual agent command.
- Repository registration and project creation/editing preserve the distinction
  between repositories and projects. A project has explicit roots that can overlap
  and span repositories, with roles, versions, filters and build units.
- Source search submits the selected workspace, project and immutable view with the
  real query DSL, substring/exact/regex modes, case sensitivity and bounded pages.
  Complete/partial scans, result truncation and per-file highlight truncation are
  visible. Source files are loaded from the result's view. Original UTF-8 byte
  ranges highlight whole lines safely across Unicode and CRLF.
- Type inspection presents all exact/ambiguous candidates without choosing one.
  Cards show scope, recorded counts, fields, properties, methods, constructors,
  parameters, returns, generic constraints, modifiers, visibility, attributes and
  typed argument values, underlying types, enum members, typed constants and usages,
  array shape, and profile-specific memory-layout facts including field offsets.
  Clickable facts expose producer evidence and scope. Missing facts render as
  unresolved; unsupported facts and known zero are distinct.
- Capabilities render the API's actual features and limits. The future relationship
  graph navigation is disabled. Language-neutral metadata contracts do not imply
  availability of every extractor or compiler producer.
- Project, view, query, mode, case sensitivity, type name, build profile and selected
  source file are represented in URL parameters, with browser back/forward support.

## Authentication

The login form sends the bootstrap token once in a POST body to `/api/v1/session`.
The token field is cleared immediately on submission, including failed attempts.
No token is written to localStorage, sessionStorage or a URL. API requests use
`credentials: include` with the server's HttpOnly SameSite session cookie. A 401
returns the user to the session gate. Ending the session calls DELETE `/session`
and clears the query cache. API failures show actionable messages and retry controls.

## Validation and limits

Validated with clean `npm ci`, `npm run build`, and 12 focused Vitest tests covering
known-zero versus unresolved layout facts, evidence rendering, UTF-8/CRLF ranges,
scoped search serialization, auth token non-persistence, API error behavior, portable
root paths and credential-free repository URLs. Login layout inspected in the live
in-app browser at 1280×720. Full Compose and authenticated browser acceptance are
coordinated by the root integration task.

Source viewing is a scrollable, line-numbered text view with line highlighting;
syntax coloring and a virtualized editor are not implemented. Large source inputs
remain bounded by the API. Recorded member counts describe metadata returned by
the producer, not semantic completeness. Graph exploration, arbitrary project
deletion and ingestion uploads through the browser are not implemented. The CLI
performs ingestion. Session authentication currently follows the API's local
bootstrap model, not an identity-provider login.

HeroUI integration was checked against its official
[quick start](https://heroui.com/en/docs/react/getting-started/quick-start) and
[button API](https://heroui.com/en/docs/react/components/button). Non-root nginx
configuration follows the [official image repository](https://github.com/nginx/docker-nginx-unprivileged).
