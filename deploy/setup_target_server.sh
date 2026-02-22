#!/usr/bin/env bash
set -euo pipefail

# One-shot remote bootstrap wrapper.
# Run this from your local machine to setup a target Ubuntu server.
#
# Example:
#   SSH_HOST=203.0.113.10 DOMAIN=bandcash.app ./deploy/setup_target_server.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

SSH_HOST="${SSH_HOST:-}"
SSH_USER="${SSH_USER:-root}"
SSH_PORT="${SSH_PORT:-22}"

DOMAIN="${DOMAIN:-}"
REPO_URL="${REPO_URL:-https://github.com/peterszarvas94/bandcash.git}"
DEPLOY_USER="${DEPLOY_USER:-deploy}"
APP_ROOT="${APP_ROOT:-/opt/bandcash}"
APP_PORT="${APP_PORT:-2222}"
PUBLIC_URL="${PUBLIC_URL:-}"

if [ -z "$SSH_HOST" ]; then
  echo "SSH_HOST is required (example: SSH_HOST=203.0.113.10)" >&2
  exit 1
fi

if [ -z "$DOMAIN" ]; then
  echo "DOMAIN is required (example: DOMAIN=bandcash.app)" >&2
  exit 1
fi

REMOTE_BOOTSTRAP="/tmp/bandcash-bootstrap-server.sh"

echo "[setup] Copying bootstrap script to ${SSH_USER}@${SSH_HOST}:${REMOTE_BOOTSTRAP}"
scp -P "$SSH_PORT" "$SCRIPT_DIR/bootstrap_server.sh" "${SSH_USER}@${SSH_HOST}:${REMOTE_BOOTSTRAP}"

echo "[setup] Running remote bootstrap"
ssh -p "$SSH_PORT" "${SSH_USER}@${SSH_HOST}" \
  "chmod +x '$REMOTE_BOOTSTRAP' && sudo DOMAIN='$DOMAIN' REPO_URL='$REPO_URL' DEPLOY_USER='$DEPLOY_USER' APP_ROOT='$APP_ROOT' APP_PORT='$APP_PORT' PUBLIC_URL='$PUBLIC_URL' bash '$REMOTE_BOOTSTRAP'"

echo "[setup] Cleaning remote temp file"
ssh -p "$SSH_PORT" "${SSH_USER}@${SSH_HOST}" "rm -f '$REMOTE_BOOTSTRAP'"

echo "[setup] Done"
echo "Next steps on server:"
echo "1) Edit /etc/systemd/system/bandcash.service and set SMTP_* + EMAIL_FROM"
echo "2) sudo systemctl daemon-reload && sudo systemctl restart bandcash"
echo "3) Verify: curl -i https://${DOMAIN}/health"
