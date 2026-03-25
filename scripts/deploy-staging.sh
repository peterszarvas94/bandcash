#!/bin/bash
set -e

# Check if we're on the development branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "development" ]; then
  echo "Error: Staging deployment must be from 'development' branch"
  echo "Current branch: $CURRENT_BRANCH"
  exit 1
fi

echo "Deploying to staging from development branch..."

# Run kamal deploy with staging config (skip hooks to avoid lock conflict)
KAMAL_BIN=$(mise which kamal)
env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" deploy --config-file config/deploy.staging.yml --skip-hooks

# Boot BetterStack accessory after deployment
echo "Starting BetterStack logging accessory..."
if env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory details better-stack --config-file config/deploy.staging.yml >/dev/null 2>&1; then
  env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory reboot better-stack --config-file config/deploy.staging.yml
else
  env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory boot better-stack --config-file config/deploy.staging.yml
fi

# After successful deployment, run seed command on staging server
echo "Seeding staging database..."
env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" app exec --config-file config/deploy.staging.yml '/app/seed --db /storage/sqlite.db'

echo "Staging deployment complete!"
