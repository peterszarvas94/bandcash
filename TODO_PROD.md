# Production (Kamal)

This repo now uses Kamal for deployment. Legacy GitHub SSH/systemd deploy files were removed.

## 0) If server is partially configured from old setup

Run from local machine:

```bash
SSH_HOST=bandcash FORCE=1 ./deploy/reset_partial_server.sh
```

If you also want to replace Caddy config with a simple placeholder while resetting:

```bash
SSH_HOST=bandcash FORCE=1 RESET_CADDY=1 ./deploy/reset_partial_server.sh
```

## 1) Local prerequisites

- Install Kamal (`gem install kamal`).
- Ensure Docker can build locally.
- Ensure SSH works: `ssh deploy@bandcash`.

## 2) Server prerequisites (no app deploy yet)

On server (`bandcash`):

```bash
sudo apt-get update
sudo apt-get -y upgrade
sudo apt-get install -y docker.io curl git
sudo usermod -aG docker deploy
```

Then re-login so docker group is active for `deploy`.

## 3) Configure Kamal files in repo

- Main config: `config/deploy.yml`
- Secrets template: `.kamal/secrets-common`

Fill `.kamal/secrets-common` locally:

- `KAMAL_REGISTRY_USERNAME`
- `KAMAL_REGISTRY_PASSWORD`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`

Notes:

- App data is persisted via volume: `bandcash_data:/storage`
- SQLite path in container: `/storage/sqlite.db`
- Healthcheck path: `/health`

## 4) First-time Kamal setup

From repo root (local machine):

```bash
kamal setup
```

This installs/boots kamal-proxy and prepares first app release.

## 5) Subsequent deploys

```bash
kamal deploy
```

## 6) Verify

```bash
curl -i https://bandcash.app/health
kamal app details
kamal app logs
```

## 7) Rollback / cleanup helpers

- Rollback: `kamal rollback`
- Remove old containers/images: `kamal prune all`
