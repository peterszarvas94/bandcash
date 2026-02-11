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
mise run dev
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

## Mise Tasks

```bash
mise run dev         # Hot reload development
mise run build       # Build binary to tmp/server
mise run start       # Run built binary
mise run run         # Run server directly
mise run test        # Run tests
mise run fmt         # Format code
mise run vet         # Vet code
mise run check       # fmt + vet + test
mise run sqlc        # Generate sqlc code
mise run goose-up    # Run migrations
mise run goose-status# Migration status
mise run goose-create name=add_new_table
mise run seed        # Seed database
```

## Database

Migrations are handled by goose, and queries are compiled by sqlc.

Run migrations:
```bash
mise run goose-up
```

Create a new migration:
```bash
mise run goose-create name=add_new_column
```

Regenerate sqlc code:
```bash
mise run sqlc
```

## Docker

```bash
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
