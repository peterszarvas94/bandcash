# Production (Kamal) — Remaining Steps

## 1) Prepare server runtime

If you use `./scripts/secure_server_bootstrap.sh`, Docker/runtime packages and `mise` are installed there.

Then logout/login so docker group is active for the admin user.

## 2) Confirm local prerequisites

```bash
mise install
ssh peti@bandcash
```

## 3) Fill Kamal secrets

Set these in `.kamal/secrets` (local only):

- `KAMAL_REGISTRY_USERNAME`
- `KAMAL_REGISTRY_PASSWORD`
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `EMAIL_FROM`

## 4) First deploy

```bash
mise run kamal setup
```

Notes:

- Builds are configured to run on the remote server builder (`ssh://peti@bandcash`) to avoid local macOS cross-arch builds.
- DB migrations run automatically on app startup (every deploy) from embedded migration files.

## 5) Verify

```bash
curl -i https://bandcash.app/health
mise run kamal app details
mise run kamal app logs
```

## 6) Ongoing commands

- Deploy updates: `mise run kamal deploy`
- Rollback: `mise run kamal rollback`
- Prune old artifacts: `mise run kamal prune all`
