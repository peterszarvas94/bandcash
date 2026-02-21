# Production TODO

## Server Command Runbook

Run these on the new Ubuntu server (keep your current SSH session open while hardening SSH):

```bash
# 1) Base update
sudo apt-get update && sudo apt-get -y upgrade

# 2) Create deploy user
sudo adduser deploy
sudo usermod -aG sudo deploy

# 3) Copy current key to deploy user
sudo mkdir -p /home/deploy/.ssh
sudo cp ~/.ssh/authorized_keys /home/deploy/.ssh/authorized_keys
sudo chown -R deploy:deploy /home/deploy/.ssh
sudo chmod 700 /home/deploy/.ssh
sudo chmod 600 /home/deploy/.ssh/authorized_keys

# 4) SSH hardening (manual edit with vim)
sudo vim /etc/ssh/sshd_config
# Ensure these are present (and not contradicted later):
# PasswordAuthentication no
# KbdInteractiveAuthentication no
# ChallengeResponseAuthentication no
# PubkeyAuthentication yes
# PermitRootLogin no
sudo sshd -t
sudo systemctl restart ssh

# 5) Optional: passwordless sudo for deploy user
sudo visudo -f /etc/sudoers.d/99-deploy-nopasswd
# Add this line:
# deploy ALL=(ALL) NOPASSWD:ALL
sudo chmod 0440 /etc/sudoers.d/99-deploy-nopasswd

# 6) Firewall
sudo apt-get install -y ufw
sudo ufw allow OpenSSH
sudo ufw allow 80
sudo ufw allow 443
sudo ufw --force enable
sudo ufw status

# 7) Clone and bootstrap BandCash
git clone https://github.com/peterszarvas94/bandcash.git
cd bandcash
sudo DOMAIN=bandcash.app REPO_URL=https://github.com/peterszarvas94/bandcash.git bash deploy/bootstrap_server.sh

# 8) Configure SMTP and restart app
sudo nano /etc/systemd/system/bandcash.service
sudo systemctl daemon-reload
sudo systemctl restart bandcash
sudo systemctl status bandcash --no-pager
sudo systemctl status caddy --no-pager

# 9) Verify
curl -I https://bandcash.app
curl -i https://bandcash.app/health
```

## Goal

Deploy BandCash to a single VPS with a minimal, production-ready setup:

- GitHub Release based deploys (tag-driven, immutable)
- GitHub Actions -> SSH -> server deploy script
- systemd service for app runtime
- Caddy as HTTPS reverse proxy
- SQLite persisted on disk with pre-deploy backup

## 1) Server Provisioning

- [ ] Run one-shot bootstrap script:
  - [ ] `sudo DOMAIN=bandcash.app REPO_URL=https://github.com/peterszarvas94/bandcash.git bash deploy/bootstrap_server.sh`
- [ ] Create VPS (Ubuntu 24.04 LTS recommended, 2 vCPU / 2 GB RAM / 40 GB SSD)
- [ ] Point domain `A` record to server IP
- [ ] Open firewall ports: `22`, `80`, `443`
- [ ] Create non-root deploy user (sudo enabled)
- [ ] Configure SSH key auth and disable password login

## 2) Runtime Setup

- [ ] Install packages: `git`, `go`, `sqlite3`, `caddy`, `curl`
- [ ] Create directories:
  - [ ] `/opt/bandcash/app`
  - [ ] `/opt/bandcash/bin`
  - [ ] `/opt/bandcash/data`
  - [ ] `/opt/bandcash/backups`
  - [ ] `/opt/bandcash/logs`
- [ ] Clone repository to `/opt/bandcash/app`
- [ ] Copy `.env.example` to production env values in systemd service
- [ ] Ensure DB path points to persistent location (`/opt/bandcash/data/sqlite.db`)

## 3) Service + Proxy

- [ ] Install `deploy/bandcash.service` to `/etc/systemd/system/bandcash.service`
- [ ] Update service environment values (`URL`, SMTP, etc.)
- [ ] Run `sudo systemctl daemon-reload`
- [ ] Run `sudo systemctl enable bandcash`
- [ ] Install `deploy/Caddyfile` to `/etc/caddy/Caddyfile`
- [ ] Replace domain placeholder in Caddyfile
- [ ] Run `sudo systemctl reload caddy`

## 4) GitHub Deploy Automation

- [ ] Add repository secrets:
  - [ ] `PROD_SSH_PRIVATE_KEY`
  - [ ] `PROD_SSH_HOST`
  - [ ] `PROD_SSH_USER`
  - [ ] `PROD_SSH_PORT` (optional, default `22`)
  - [ ] `PROD_APP_DIR` (optional, default `/opt/bandcash/app`)
- [ ] Verify workflow file `.github/workflows/deploy-release.yml`
- [ ] Ensure `deploy/deploy.sh` exists on server in app repo
- [ ] Test with a prerelease tag first

## 5) Release Process

- [ ] Create and push tag (example: `v0.1.0`)
- [ ] Publish GitHub Release for that tag
- [ ] Confirm GitHub Action deploy job succeeds
- [ ] Verify health endpoint returns success (`/health`)
- [ ] Verify login/dashboard/group pages work in production

## 6) Backup + Recovery

- [ ] Confirm pre-deploy DB backup is created in `/opt/bandcash/backups`
- [ ] Add scheduled SQLite backups (daily cron/systemd timer)
- [ ] Test restore once on a copy

## Future Improvements

- Add staging environment and promote releases to prod
- Add automatic rollback when health check fails
- Add migration step in deploy script when schema changes are introduced
- Add uptime monitoring and error tracking
- Move to container-based deploy (for example Kamal) only when complexity requires it

## Good First Release

- Suggested first production tag: `v0.1.0`
