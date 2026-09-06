# Contributing

Use Go 1.26.8, Git, Docker with Compose, and Node.js for the later web component.
Run `go vet ./...`, `go test ./...`, and the Compose integration tests for storage changes.
Format Go with gofmt. Preserve AGPL-3.0 notices and document external dependencies.

Each finished development stage/wave receives a separate conventional commit after
verification. Include the changed behavior, evidence/limitations, and checks in its
description. Unimplemented analyzers must be reported as unsupported, not ready.

Do not add credentials, raw configuration artifacts, toolchains, corpora or runtime
databases to Git. Fixtures must be small, synthetic, and independently checked.
