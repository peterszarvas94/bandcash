#!/usr/bin/env bash
set -euo pipefail

: "${RELEASE_TAG:?RELEASE_TAG is required}"

APP_DIR="${APP_DIR:-/opt/bandcash/app}"
BIN_DIR="${BIN_DIR:-/opt/bandcash/bin}"
DATA_DIR="${DATA_DIR:-/opt/bandcash/data}"
BACKUP_DIR="${BACKUP_DIR:-/opt/bandcash/backups}"
DB_PATH="${DB_PATH:-$DATA_DIR/sqlite.db}"
SERVICE_NAME="${SERVICE_NAME:-bandcash}"
PORT="${PORT:-2222}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:${PORT}/health}"

LOCK_FILE="/tmp/${SERVICE_NAME}-deploy.lock"

mkdir -p "$BIN_DIR" "$DATA_DIR" "$BACKUP_DIR"

exec 9>"$LOCK_FILE"
flock -n 9 || {
  echo "Another deploy is already running." >&2
  exit 1
}

echo "[deploy] starting ${RELEASE_TAG}"

cd "$APP_DIR"

git fetch --tags origin
git checkout --force "refs/tags/${RELEASE_TAG}"
git clean -fd

if [ -f "$DB_PATH" ]; then
  TS="$(date +%Y%m%d-%H%M%S)"
  cp "$DB_PATH" "$BACKUP_DIR/sqlite-${TS}.db"
  echo "[deploy] backup created: $BACKUP_DIR/sqlite-${TS}.db"
fi

go build -o "$BIN_DIR/server" ./cmd/server/main.go

sudo systemctl restart "$SERVICE_NAME"

sleep 2
curl -fsS "$HEALTH_URL" >/dev/null

echo "[deploy] success: $(git rev-parse --short HEAD)"
