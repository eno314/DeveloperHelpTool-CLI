# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# dev: Development environment for running tests, linting, and local builds.
#      Source code is mounted as a volume at runtime (see compose.yaml).
# ---------------------------------------------------------------------------
FROM golang:1.26-bookworm AS dev

WORKDIR /app

# Download dependencies with a cache mount to speed up repeated builds.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# Default command: run tests (can be overridden via compose).
CMD ["go", "test", "-v", "-race", "./..."]

# ---------------------------------------------------------------------------
# builder: Produces a static binary for Linux (used in CI or manual builds).
# ---------------------------------------------------------------------------
FROM dev AS builder

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /developer-help-tool-cli ./src

# ---------------------------------------------------------------------------
# final: Minimal distroless image containing only the compiled binary.
# ---------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS final

COPY --from=builder /developer-help-tool-cli /developer-help-tool-cli

ENTRYPOINT ["/developer-help-tool-cli"]
