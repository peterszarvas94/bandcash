#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "$SCRIPT_DIR/.." && pwd)"

read_assignment_from_file() {
  local file_path="$1"
  local key="$2"
  local line value

  [ -f "$file_path" ] || return 0

  line="$(grep -E "^${key}=" "$file_path" | tail -n 1 || true)"
  [ -n "$line" ] || return 0

  value="${line#*=}"
  value="${value%\"}"
  value="${value#\"}"
  value="${value%\'}"
  value="${value#\'}"
  printf '%s' "$value"
}

load_op_config_from_file() {
  local file_path="$1"
  local val

  if [ -z "${OP_ACCOUNT:-}" ]; then
    val="$(read_assignment_from_file "$file_path" "OP_ACCOUNT")"
    if [ -n "$val" ]; then
      export OP_ACCOUNT="$val"
    fi
  fi
  if [ -z "${OP_FROM_DEVELOPMENT:-}" ]; then
    val="$(read_assignment_from_file "$file_path" "OP_FROM_DEVELOPMENT")"
    if [ -n "$val" ]; then
      export OP_FROM_DEVELOPMENT="$val"
    fi
  fi
  if [ -z "${OP_FROM:-}" ]; then
    val="$(read_assignment_from_file "$file_path" "OP_FROM")"
    if [ -n "$val" ]; then
      export OP_FROM="$val"
    fi
  fi
  if [ -z "${OP_FROM_LOCALHOST:-}" ]; then
    val="$(read_assignment_from_file "$file_path" "OP_FROM_LOCALHOST")"
    if [ -n "$val" ]; then
      export OP_FROM_LOCALHOST="$val"
    fi
  fi
}

fetch_local_secrets_from_1password() {
  local op_account op_from secrets_blob value
  local keys

  op_account="${OP_ACCOUNT:-}"
  if [ "${APP_ENV:-development}" = "development" ] && [ -n "${OP_FROM_LOCALHOST:-}" ]; then
    op_from="${OP_FROM_LOCALHOST}"
  else
    op_from="${OP_FROM_DEVELOPMENT:-${OP_FROM:-}}"
  fi
  if [ -z "$op_account" ] || [ -z "$op_from" ]; then
    return 0
  fi

  if ! command -v kamal >/dev/null 2>&1; then
    return 0
  fi

  keys=(
    APP_ENV
    HOST
    PORT
    URL
    DB_PATH
    LOG_LEVEL
    LOG_FOLDER
    LOG_PREFIX
    DISABLE_RATE_LIMIT
    SUPERADMIN_EMAIL
    EMAIL_PROVIDER
    EMAIL_FROM
    MAILTRAP_HOST
    MAILTRAP_PORT
    MAILTRAP_USERNAME
    MAILTRAP_PASSWORD
    LEMON_WEBHOOK_SECRET
    LEMON_HOSTED_URL
  )

  if ! secrets_blob="$(kamal secrets fetch --adapter 1password --account "$op_account" --from "$op_from" "${keys[@]}")"; then
    if [ -z "${LEMON_HOSTED_URL:-}" ]; then
      echo "with_dev_secrets: could not load local secrets via 1Password (check 'op signin' and OP_FROM_LOCALHOST/OP_FROM_DEVELOPMENT)." >&2
    fi
    return 0
  fi

  for key in "${keys[@]}"; do
    value="$(kamal secrets extract "$key" "$secrets_blob" 2>/dev/null || true)"
    if [[ "$value" == *"ERROR (RuntimeError)"* ]] || [[ "$value" == *"Could not find secret"* ]]; then
      value=""
    fi
    if [ -n "$value" ]; then
      export "$key=$value"
    fi
  done
}

load_op_config_from_file "$ROOT_DIR/.kamal/secrets.development"
load_op_config_from_file "$ROOT_DIR/.kamal/secrets"

fetch_local_secrets_from_1password

exec "$@"
