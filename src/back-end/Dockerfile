# syntax=docker/dockerfile:1.7
FROM node:26.8.1-bookworm-slim@sha256:367679cf9792759492a486e4aa4b421764d71a9546a6dae8aab81a99eb797b3e AS syntax-build
WORKDIR /helper
COPY src/back-end/analyzers/syntax/package*.json ./
RUN npm ci --ignore-scripts
COPY src/back-end/analyzers/syntax/ ./
RUN npm test

FROM golang:1.26.8-bookworm@sha256:9fdc884aacc3bec89b20ffc69f4bb369c78210e3e4f600387b5128b12c199f81 AS build
WORKDIR /src
COPY src/back-end/go.mod src/back-end/go.sum ./
RUN go mod download
COPY src/back-end/ ./
COPY deploy/ai-profiles.example.json /deploy/ai-profiles.example.json
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=development
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/...

FROM build AS test
COPY --from=syntax-build /usr/local/bin/node /usr/local/bin/node
COPY --from=syntax-build /helper /opt/valio/syntax
ENV VALIO_TEST_SYNTAX_HELPER=/opt/valio/syntax/index.js
RUN go test -race ./...

FROM debian:bookworm-slim@sha256:88200866dfff7ea7f5cbcb6ec7c8a701889efe6fe859fe64d6990e4b07ea4171 AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libstdc++6 libatomic1 && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/ /usr/local/bin/
USER 65532:65532
WORKDIR /tmp
ENV TMPDIR=/tmp

FROM runtime AS migrate
ENTRYPOINT ["valio-admin", "migrate"]

FROM runtime AS worker
HEALTHCHECK --interval=30s --timeout=5s CMD ["valio-worker", "health"]
ENTRYPOINT ["valio-worker"]

FROM runtime AS api
COPY --from=syntax-build /usr/local/bin/node /usr/local/bin/node
COPY --from=syntax-build /helper /opt/valio/syntax
ENV VALIO_SYNTAX_HELPER=/opt/valio/syntax/index.js
RUN node --version
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s CMD ["valio-api", "health"]
ENTRYPOINT ["valio-api"]
