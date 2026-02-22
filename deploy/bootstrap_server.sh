#!/usr/bin/env bash
set -euo pipefail

# One-shot server bootstrap for BandCash.
# Run on a fresh Ubuntu server as root (or with sudo).
#
# Example:
#   sudo DOMAIN=bandcash.app REPO_URL=https://github.com/peterszarvas94/bandcash.git bash deploy/bootstrap_server.sh

DOMAIN="${DOMAIN:-}"
REPO_URL="${REPO_URL:-https://github.com/peterszarvas94/bandcash.git}"
DEPLOY_USER="${DEPLOY_USER:-deploy}"
APP_ROOT="${APP_ROOT:-/opt/bandcash}"
APP_PORT="${APP_PORT:-2222}"
PUBLIC_URL="${PUBLIC_URL:-}"

if [ -z "$DOMAIN" ]; then
  echo "DOMAIN is required (example: DOMAIN=bandcash.app)" >&2
  exit 1
fi

if [ -z "$PUBLIC_URL" ]; then
  PUBLIC_URL="https://${DOMAIN}"
fi

APP_DIR="${APP_ROOT}/app"
BIN_DIR="${APP_ROOT}/bin"
DATA_DIR="${APP_ROOT}/data"
BACKUP_DIR="${APP_ROOT}/backups"
LOG_DIR="${APP_ROOT}/logs"

echo "[bootstrap] Installing base packages"
apt-get update
apt-get install -y --no-install-recommends git curl sqlite3 ca-certificates caddy golang-go

if ! id "$DEPLOY_USER" >/dev/null 2>&1; then
  echo "[bootstrap] Creating deploy user: ${DEPLOY_USER}"
  adduser --disabled-password --gecos "" "$DEPLOY_USER"
  usermod -aG sudo "$DEPLOY_USER"
fi

echo "[bootstrap] Creating app directories"
mkdir -p "$APP_ROOT" "$BIN_DIR" "$DATA_DIR" "$BACKUP_DIR" "$LOG_DIR"
chown -R "$DEPLOY_USER:$DEPLOY_USER" "$APP_ROOT"

if [ ! -d "$APP_DIR/.git" ]; then
  echo "[bootstrap] Cloning repository"
  sudo -u "$DEPLOY_USER" git clone "$REPO_URL" "$APP_DIR"
else
  echo "[bootstrap] Repository exists, updating"
  sudo -u "$DEPLOY_USER" bash -c "cd '$APP_DIR' && git fetch --all --tags"
fi

echo "[bootstrap] Installing deploy script"
chmod 0755 "$APP_DIR/deploy/deploy.sh"
chown "$DEPLOY_USER:$DEPLOY_USER" "$APP_DIR/deploy/deploy.sh"

echo "[bootstrap] Installing systemd service"
sed \
  -e "s|^User=.*$|User=${DEPLOY_USER}|" \
  -e "s|^Group=.*$|Group=${DEPLOY_USER}|" \
  -e "s|^Environment=PORT=.*$|Environment=PORT=${APP_PORT}|" \
  -e "s|^Environment=URL=.*$|Environment=URL=${PUBLIC_URL}|" \
  -e "s|^Environment=DB_PATH=.*$|Environment=DB_PATH=${DATA_DIR}/sqlite.db|" \
  -e "s|^Environment=LOG_FOLDER=.*$|Environment=LOG_FOLDER=${LOG_DIR}|" \
  "$APP_DIR/deploy/bandcash.service" > /etc/systemd/system/bandcash.service

echo "[bootstrap] Configuring sudoers for deploy restarts"
cat >/etc/sudoers.d/bandcash-deploy <<EOF
${DEPLOY_USER} ALL=(root) NOPASSWD: /bin/systemctl restart bandcash, /bin/systemctl status bandcash
EOF
chmod 0440 /etc/sudoers.d/bandcash-deploy

echo "[bootstrap] Installing Caddy config"
sed \
  -e "s|bandcash.app|${DOMAIN}|g" \
  -e "s|127.0.0.1:2222|127.0.0.1:${APP_PORT}|g" \
  "$APP_DIR/deploy/Caddyfile" > /etc/caddy/Caddyfile

echo "[bootstrap] Enabling services"
systemctl daemon-reload
systemctl enable caddy
systemctl restart caddy
systemctl enable bandcash

echo "[bootstrap] Done"
echo "- Domain: ${DOMAIN}"
echo "- App dir: ${APP_DIR}"
echo "- Deploy user: ${DEPLOY_USER}"
echo "- Public URL: ${PUBLIC_URL}"
echo ""
echo "Next steps:"
echo "1) Edit /etc/systemd/system/bandcash.service and set SMTP + EMAIL_FROM values."
echo "2) systemctl daemon-reload"
echo "3) Add GitHub secrets (PROD_SSH_*)."
echo "4) Publish a GitHub release tag (for example v0.1.0)."
