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
app/                # Feature modules (entry, home)
  entry/            # Entry REST handlers + templates
internal/
  config/           # Environment configuration
  db/               # Database init, sqlc output, migrations, queries
  logger/           # Colored slog handler + JSON file logging
  middleware/       # Request ID middleware
  utils/            # Shared utilities
web/
  static/js/        # JavaScript (Datastar vendored)
  static/css/       # Layered CSS (reset/base/components/utilities)
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
mise run dev                              # Hot reload development
mise run build                            # Build binary to tmp/server
mise run start                            # Run built binary
mise run run                              # Run server directly
mise run test                             # Run tests
mise run fmt                              # Format code
mise run vet                              # Vet code
mise run check                            # fmt + vet + test
mise run sqlc                             # Generate sqlc code
mise run goose-up                         # Run migrations
mise run goose-status                     # Migration status
mise run goose-create name=add_new_table  # Create migration
mise run seed                             # Seed database
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

## Naming Conventions

Rails-style conventions are used where it makes sense:

- REST handlers follow `index`, `new`, `show`, `edit`, `create`, `update`, `destroy`.
- Templates use `index.html`, `new.html`, `show.html`, `edit.html`.

## Docker

```bash
# Or manually
docker build -t bandcash .
docker run -p 8080:8080 bandcash

# Or with docker compose
docker compose up --build
```

## How It Works

1. Browser loads page and renders templates
2. User actions trigger POST/PUT/DELETE requests
3. Handlers read/write entries in SQLite via sqlc
4. Responses re-render the requested pages

## Realtime (SSE)

- The client opens a single SSE stream at `/sse`.
- A `view` signal (matching the route, e.g. `/entry/1`) is sent on connect.
- The server renders the view and patches `main#app` via Datastar SSE events.
- Handlers trigger updates by calling the hub (`hub.Hub.Render`) and can patch signals.

## License

[O’Saasy](https://osaasy.dev/) © 2026 Peter Szarvas

See `LICENSE.md`
