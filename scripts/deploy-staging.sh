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

# Run kamal deploy with staging config
KAMAL_BIN=$(mise which kamal)
env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" deploy --config-file config/deploy.staging.yml

# Boot BetterStack accessory after deployment
echo "Starting BetterStack logging accessory..."
if env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory details better-stack --config-file config/deploy.staging.yml >/dev/null 2>&1; then
  env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory reboot better-stack --config-file config/deploy.staging.yml
else
  env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory boot better-stack --config-file config/deploy.staging.yml
fi

# Cleanup old Docker build cache
echo "Pruning Docker build cache..."
ssh peti@bandcash_staging "docker buildx prune -af --filter 'until=24h'" || true

echo "Staging deployment complete!"
