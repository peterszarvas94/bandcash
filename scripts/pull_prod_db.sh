#!/usr/bin/env bash
set -euo pipefail

# Pull a consistent snapshot of production SQLite from docker volume.
# Requires SSH access to the deploy host and docker permissions on host.

HOST_ALIAS="${HOST_ALIAS:-bandcash}"
DB_VOLUME="${DB_VOLUME:-bandcash_data}"
DB_PATH_IN_VOLUME="${DB_PATH_IN_VOLUME:-sqlite.db}"
OUT_DIR="${OUT_DIR:-tmp/prod-db}"

mkdir -p "$OUT_DIR"

timestamp="$(date +%Y%m%d-%H%M%S)"
out_file="$OUT_DIR/prod-$timestamp.db"

echo "Creating production DB snapshot from volume '$DB_VOLUME' on '$HOST_ALIAS'..."

ssh "$HOST_ALIAS" "docker run --rm -v $DB_VOLUME:/storage alpine:3.20 sh -lc 'apk add --no-cache sqlite >/dev/null && sqlite3 /storage/$DB_PATH_IN_VOLUME \".backup /storage/.db_snapshot.db\" && cat /storage/.db_snapshot.db && rm -f /storage/.db_snapshot.db'" > "$out_file"

if [ ! -s "$out_file" ]; then
  echo "Snapshot failed: output file is empty: $out_file" >&2
  exit 1
fi

cp "$out_file" "$OUT_DIR/prod-latest.db"

echo "Saved: $out_file"
echo "Updated: $OUT_DIR/prod-latest.db"
