#!/usr/bin/env bash
set -euo pipefail

# Download latest production JSON log files from server as fallback/history.
#
# Usage:
#   HOST=bandcash ./scripts/download_prod_logs.sh
#   HOST=bandcash COUNT=3 OUT_DIR=logs/prod ./scripts/download_prod_logs.sh

HOST="${HOST:-bandcash}"
SSH_USER="${SSH_USER:-peti}"
SSH_PORT="${SSH_PORT:-22}"
COUNT="${COUNT:-1}"
OUT_DIR="${OUT_DIR:-logs/prod}"
REMOTE_LOG_DIR="${REMOTE_LOG_DIR:-/var/lib/docker/volumes/bandcash_data/_data/logs}"

mkdir -p "$OUT_DIR"

echo "Fetching latest $COUNT log file(s) from $SSH_USER@$HOST:$REMOTE_LOG_DIR"

latest_files=()
while IFS= read -r line; do
  latest_files+=("$line")
done < <(ssh -p "$SSH_PORT" "$SSH_USER@$HOST" "sudo sh -lc 'ls -1t $REMOTE_LOG_DIR/*.log 2>/dev/null | head -n $COUNT'")

if [ "${#latest_files[@]}" -eq 0 ]; then
  echo "No log files found at $REMOTE_LOG_DIR"
  exit 0
fi

downloaded=0
for remote_file in "${latest_files[@]}"; do
  [ -n "$remote_file" ] || continue
  local_file="$OUT_DIR/$(basename "$remote_file")"
  ssh -p "$SSH_PORT" "$SSH_USER@$HOST" "sudo cat '$remote_file'" > "$local_file"
  echo "Downloaded: $local_file"
  downloaded=$((downloaded + 1))
done

echo "Done. Downloaded $downloaded file(s)."
