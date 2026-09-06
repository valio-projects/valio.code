# syntax=docker/dockerfile:1.7
FROM golang:1.26.8-bookworm@sha256:9fdc884aacc3bec89b20ffc69f4bb369c78210e3e4f600387b5128b12c199f81 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=development
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/...

FROM build AS test
RUN go test -race ./...

FROM debian:bookworm-slim@sha256:88200866dfff7ea7f5cbcb6ec7c8a701889efe6fe859fe64d6990e4b07ea4171 AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
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
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s CMD ["valio-api", "health"]
ENTRYPOINT ["valio-api"]
