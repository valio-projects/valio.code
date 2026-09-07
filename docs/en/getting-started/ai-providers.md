# AI embedding providers

Copy [`deploy/ai-profiles.example.json`](../../../deploy/ai-profiles.example.json)
to an operator-managed location and select a profile explicitly. The file is
strict JSON: unknown fields, trailing data, duplicate IDs, invalid endpoints,
dimensions, token-environment names, and timeouts are rejected. It contains no
token value; a remote token is read only from the named `tokenEnv` variable.

The API loads `VALIO_AI_CONFIG` once during startup. In the supplied Compose
configuration, the example is mounted read-only at
`/etc/valio/ai-profiles.json`; changing the file requires an API restart.
Loading configuration does not contact a provider, start a model, pull a model,
or send source text.

| Provider mode | Profile base URL | Request path and width behavior |
| --- | --- | --- |
| Ollama on the API host | `http://127.0.0.1:11434` | `POST /api/embed`; Valio sends the configured `dimensions` and `truncate: false`. |
| LM Studio or another OpenAI-compatible host | `http://127.0.0.1:1234/v1` | `POST /embeddings`; the server's native output must match the profile dimension. |
| Host model server reached from Compose | `http://host.docker.internal:1234/v1` | Same OpenAI-compatible request; the example sets `allowInsecureHTTP: true` deliberately. |
| Remote provider | `https://…/v1` | `POST /embeddings`; use HTTPS and a token environment variable where required. |

`allowInsecureHTTP` defaults to false. Set it only for an operator-controlled
local, Docker-host, or LAN endpoint. It is not a substitute for transport
security on a remote provider.

Qwen publishes Qwen3 Embedding 4B with output dimensions from 32 through 2560
and 8B from 32 through 4096. The profile fixes one width, model revision,
quantization, representation kind, and cosine vector space. The six local
aliases in the example use the configured widths for their selected GGUF
artifacts; they are not claims that every quantization is an Ollama library tag.
See Qwen's official [4B model card](https://huggingface.co/Qwen/Qwen3-Embedding-4B)
and [8B model card](https://huggingface.co/Qwen/Qwen3-Embedding-8B).

For Ollama, `/api/embed` documents both `dimensions` and `truncate`; a false
`truncate` value makes an over-context input fail instead of silently changing
the input. [Ollama API documentation](https://docs.ollama.com/api/embed).
OpenAI-compatible servers do not share a universal output-width contract, so
the LM Studio profile remains pinned to its expected native 4096-dimensional
output. Vectors from different profiles are never mixed.

To import a particular official GGUF artifact, create a local alias yourself:

```text
# Modelfile: choose exactly one downloaded official GGUF filename
FROM ./Qwen3-Embedding-4B-Q4_K_M.gguf
```

Run `ollama create valio-qwen3-embedding-4b-q4km -f Modelfile`; repeat for
`Q5_K_M`, `Q8_0`, and the 8B artifact. Record the exact filename, digest, and
revision in the operator profile.

Use the admin doctor deliberately after loading a model:

```text
valio-admin ai doctor --config PATH --list
valio-admin ai doctor --config PATH --profile qwen3-8b-q8
```

The doctor first performs separate, read-only two-second discovery requests to
`http://localhost:11434/api/tags` and `http://localhost:1234/v1/models`. The
selected profile then receives one fixed synthetic embedding request under that
profile's configured timeout, up to 120 seconds. It verifies the returned
vector width, but it never starts a service or downloads a model.

After startup, the API exposes the same configured profiles through these
operations. All JSON examples use ordinary text; callers cannot provide a
provider URL or arbitrary source text to the index operation.

```text
POST /api/v1/ai/probe
{"modelProfile":"qwen3-8b-q8"}

POST /api/v1/embeddings/index
{"scope":{"workspaceId":"workspace-main","viewId":"PINNED_VIEW"},"modelProfile":"qwen3-8b-q8","representation":"code","limit":16}

POST /api/v1/retrieval/search
{"scope":{"workspaceId":"workspace-main","viewId":"PINNED_VIEW"},"mode":"semantic","modelProfile":"qwen3-8b-q8","query":"where is User.ID written?","limit":20}
```

`/ai/probe` checks the configured provider with fixed synthetic text. Indexing
uses only persisted, policy-approved representations in the pinned view.
Retrieval's `query` is plain text and its result retains the selected profile
and ranking evidence.

Embedding indexing accepts at most 16 inputs per request (default 8). If
`complete` is false, repeat with the returned `nextOffset` as `offset`, keeping
the same view, profile, representation and project scope. Each completed input
is cached durably; retries reuse its immutable result.

The LM Studio profiles use the model identifier
`text-embedding-qwen3-embedding-8b`: `qwen3-8b-lmstudio-local` for a native API,
and `qwen3-8b-lmstudio-docker` for Compose. Adjust the identifier to match
`GET /v1/models` in your installation. A local server without authentication
does not require `tokenEnv`.

For an end-to-end check, run `scripts/smoke.ps1`, then pass its printed synthetic
project and view to `scripts/smoke-ai.ps1 -ProjectId PROJECT -ViewId VIEW`.
The latter probes the model, persists that project's embeddings, and verifies
semantic and hybrid retrieval. It requires a running model provider and is
separate from provider-independent CI checks.
