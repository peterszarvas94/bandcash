#!/usr/bin/env bash
set -euo pipefail

# One-shot server hardening bootstrap.
# Run from your local machine. It connects as root (password prompt via SSH),
# creates an admin user using the existing root authorized_keys, then hardens
# SSH + firewall.

HOST="${HOST:-}"
SSH_PORT="${SSH_PORT:-22}"
ADMIN_USER="${ADMIN_USER:-}"

if [ -z "$HOST" ]; then
  read -r -p "Server host (IP or DNS): " HOST
fi

if [ -z "$ADMIN_USER" ]; then
  read -r -p "New admin username: " ADMIN_USER
fi

if [ -z "$HOST" ]; then
  echo "Host is required." >&2
  exit 1
fi

if [ -z "$ADMIN_USER" ]; then
  echo "Admin username is required." >&2
  exit 1
fi

echo "[1/4] Connecting as root on $HOST:$SSH_PORT (you may be asked for root password)..."

ssh -p "$SSH_PORT" root@"$HOST" \
  ADMIN_USER="$ADMIN_USER" \
  SSH_PORT="$SSH_PORT" \
  'bash -s' <<'EOF'
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get -y upgrade
apt-get install -y sudo ufw fail2ban docker.io curl git
systemctl enable --now docker

if ! id -u "$ADMIN_USER" >/dev/null 2>&1; then
  useradd -m -s /bin/bash "$ADMIN_USER"
fi

usermod -aG sudo "$ADMIN_USER"
usermod -aG docker "$ADMIN_USER"

install -d -m 700 -o "$ADMIN_USER" -g "$ADMIN_USER" "/home/$ADMIN_USER/.ssh"
touch "/home/$ADMIN_USER/.ssh/authorized_keys"
chown "$ADMIN_USER:$ADMIN_USER" "/home/$ADMIN_USER/.ssh/authorized_keys"
chmod 600 "/home/$ADMIN_USER/.ssh/authorized_keys"

if [ -f /root/.ssh/authorized_keys ]; then
  cat /root/.ssh/authorized_keys >> "/home/$ADMIN_USER/.ssh/authorized_keys"
  sort -u "/home/$ADMIN_USER/.ssh/authorized_keys" -o "/home/$ADMIN_USER/.ssh/authorized_keys"
else
  echo "Missing /root/.ssh/authorized_keys on target host" >&2
  exit 1
fi

install -d -m 755 /etc/ssh/sshd_config.d
cat >/etc/ssh/sshd_config.d/99-hardening.conf <<SSHCONF
PermitRootLogin no
PasswordAuthentication no
KbdInteractiveAuthentication no
PubkeyAuthentication yes
ChallengeResponseAuthentication no
UsePAM yes
SSHCONF

sshd -t
systemctl reload ssh

ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow "$SSH_PORT"/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

systemctl enable --now fail2ban
EOF

echo "[2/4] Verifying SSH access as '$ADMIN_USER'..."
if ! ssh -o BatchMode=yes -o ConnectTimeout=10 -p "$SSH_PORT" "$ADMIN_USER@$HOST" "whoami" >/dev/null 2>&1; then
  echo "Failed to verify SSH login for $ADMIN_USER. Check key/agent and try: ssh -p $SSH_PORT $ADMIN_USER@$HOST" >&2
  exit 1
fi

echo "[3/4] Verifying root SSH is blocked..."
if ssh -o BatchMode=yes -o ConnectTimeout=8 -p "$SSH_PORT" root@"$HOST" "whoami" >/dev/null 2>&1; then
  echo "WARNING: root SSH still appears enabled. Check /etc/ssh/sshd_config.d/99-hardening.conf" >&2
else
  echo "Root SSH is blocked (expected)."
fi

echo "[4/4] Done."
echo "You can now use: ssh -p $SSH_PORT $ADMIN_USER@$HOST"
