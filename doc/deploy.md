# Deploy

Use this for production/staging deployment workflow only.

## Secrets and Environment

- Deployment uses Kamal with `.kamal/secrets` (gitignored).
- Safe template is `.kamal/secrets.example`.
- Prefer `kamal secrets fetch/extract` helpers in `.kamal/secrets` for 1Password-backed values.

## Commands

- `mise run deploy-check` - render Kamal config to validate secret/config resolution.
- `mise run deploy` - run Kamal production deploy, then optionally create and push a `vX.Y.Z` tag (`patch|minor|major|skip` prompt).
- `mise run deploy-staging` - run staging deploy (no version tag prompt).
- `mise run tag <patch|minor|major>` - manual tag creation/push remains available.

## Notes

- Avoid committing real secrets.
- Always run `deploy-check` before deploy.
