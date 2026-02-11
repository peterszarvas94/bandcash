# Stargazing

Go + Echo + Datastar bootstrap for real-time web apps with SSE.

## Quick Start

```bash
# Clone and setup
cp .env.example .env

# Development with hot reload
make dev
```

Open http://localhost:8080

## Project Structure

```
cmd/server/         # Application entrypoint
internal/
  config/           # Environment configuration
  handlers/         # HTTP handlers (index, sse, health)
  logger/           # Colored slog handler + JSON file logging
  middleware/       # Request ID middleware
  store/            # In-memory data store with client management
  utils/            # Shared state (store instance)
web/
  static/js/        # JavaScript (Datastar vendored)
  templates/        # Go HTML templates
examples/
  todo.go           # Example todo handler
```

## Configuration

Environment variables (see `.env.example`):

| Variable    | Default        | Description                          |
| ----------- | -------------- | ------------------------------------ |
| `PORT`      | `8080`         | Server port                          |
| `LOG_LEVEL` | `debug`        | Log level (debug, info, warn, error) |
| `LOG_FILE`  | `logs/app.log` | JSON log file path                   |

## Make Commands

```bash
make dev        # Hot reload development
make build      # Build binary to bin/
make run        # Build and run
make test       # Run tests
make fmt        # Format code
make vet        # Vet code
make check      # fmt + vet + test
make clean      # Remove build artifacts
make docker     # Build docker image
make docker-run # Build and run docker image
```

## Docker

```bash
make docker-run

# Or manually
docker build -t stargazing .
docker run -p 8080:8080 stargazing

# Or with docker compose
docker compose up --build
```

## How It Works

1. Browser loads page, connects to `/sse` endpoint
2. SSE connection stays open, receives HTML patches
3. User actions trigger POST requests (e.g., `/todo`)
4. POST handler updates state, signals SSE to push update
5. Datastar morphs the DOM with new HTML

## License

MIT
