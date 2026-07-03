# Agent Rules for DeveloperHelpTool-CLI

## Development Environment

This project uses **Podman** as the local development environment.
All build, test, and lint operations MUST be run inside the Podman container via `podman compose`.

### Prerequisites (one-time setup)

```bash
brew install podman-compose   # Install Compose provider
podman machine start          # Ensure the Podman VM is running
podman compose build          # Build the dev image
```

### Required Commands

Always use the following commands instead of running `go` directly on the host:

| Task | Command |
|---|---|
| Run tests | `podman compose run --rm dev go test -v -race ./...` |
| Build binary (Linux) | `podman compose run --rm dev go build -o developer-help-tool-cli ./src` |
| Lint (vet) | `podman compose run --rm dev go vet ./...` |
| Format check | `podman compose run --rm dev go fmt ./...` |
| Interactive shell | `podman compose run --rm dev bash` |

### Why Podman

- Ensures reproducible builds independent of the host Go installation.
- `compose.yaml` and `Dockerfile` are the single source of truth for the runtime environment.
- The `dev` image is based on `golang:1.26-bookworm` and matches the Go version in `go.mod`.

### When to rebuild the image

Rebuild the dev image (`podman compose build`) after any of the following changes:

- `go.mod` or `go.sum` is updated
- `Dockerfile` is modified
