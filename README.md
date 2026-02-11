# Bandcash

Go + Echo + Datastar app with SQLite, sqlc, and goose migrations.

## Setup

```bash
# Clone and setup
cp .env.example .env

# Trust this repo's mise config
mise trust

# Install toolchain
mise install
```

## Quick Start

```bash
# Development with hot reload
make dev
```

Open http://localhost:8080

## Project Structure

```
cmd/server/         # Application entrypoint
app/                # Feature modules (todo, entry, home)
  entry/            # Entry REST handlers + templates
internal/
  config/           # Environment configuration
  db/               # Database init, sqlc output, migrations, queries
  logger/           # Colored slog handler + JSON file logging
  middleware/       # Request ID middleware
  store/            # In-memory data store with client management
  utils/            # Shared state (store instance)
web/
  static/js/        # JavaScript (Datastar vendored)
  templates/        # Go HTML templates
```

## Configuration

Environment variables (see `.env.example`):

| Variable    | Default        | Description                          |
| ----------- | -------------- | ------------------------------------ |
| `PORT`      | `8080`         | Server port                          |
| `LOG_LEVEL` | `debug`        | Log level (debug, info, warn, error) |
| `LOG_FILE`  | `logs/app.log` | JSON log file path                   |
| `DB_PATH`   | `./sqlite.db`  | SQLite database path                 |

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

## Database

Migrations are handled by goose, and queries are compiled by sqlc.

Run migrations:
```bash
goose -dir internal/db/migrations sqlite3 ./sqlite.db up
```

Create a new migration:
```bash
goose -dir internal/db/migrations create add_new_column sql
```

Regenerate sqlc code:
```bash
sqlc generate
```

## Docker

```bash
make docker-run

# Or manually
docker build -t bandcash .
docker run -p 8080:8080 bandcash

# Or with docker compose
docker compose up --build
```

## How It Works

1. Browser loads page, connects to `/sse` or `/entry/sse` endpoint
2. SSE connection stays open, receives HTML patches
3. User actions trigger POST requests (e.g., `/todo`)
4. POST handler updates state, signals SSE to push update
5. Datastar morphs the DOM with new HTML

## License

MIT
