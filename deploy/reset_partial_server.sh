#!/usr/bin/env bash
set -euo pipefail

# Resets leftovers from the legacy systemd/Caddy deployment attempt.
# Run from local machine.
#
# Example:
#   SSH_HOST=bandcash ./deploy/reset_partial_server.sh

SSH_HOST="${SSH_HOST:-}"
SSH_USER="${SSH_USER:-deploy}"
SSH_PORT="${SSH_PORT:-22}"
APP_ROOT="${APP_ROOT:-/opt/bandcash}"
RESET_CADDY="${RESET_CADDY:-0}"
FORCE="${FORCE:-0}"

if [ -z "$SSH_HOST" ]; then
  echo "SSH_HOST is required" >&2
  exit 1
fi

if [ "$FORCE" != "1" ]; then
  echo "This will remove legacy deploy artifacts on ${SSH_USER}@${SSH_HOST}."
  echo "Re-run with FORCE=1 to continue."
  exit 1
fi

ssh -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" "
set -euo pipefail

if sudo systemctl list-unit-files | grep -q '^bandcash.service'; then
  sudo systemctl disable --now bandcash || true
  sudo rm -f /etc/systemd/system/bandcash.service
  sudo systemctl daemon-reload
fi

sudo rm -f /etc/sudoers.d/bandcash-deploy || true
sudo rm -rf '$APP_ROOT' || true

if [ '$RESET_CADDY' = '1' ]; then
  sudo tee /etc/caddy/Caddyfile >/dev/null <<'EOF'
:80 {
  respond \"Caddy is running\" 200
}
EOF
  sudo systemctl restart caddy || true
fi
"

echo "Reset complete on ${SSH_HOST}."
echo "Next: use Kamal (kamal setup) from local repo."
