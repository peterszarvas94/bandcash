#!/bin/bash
set -e

# Check if we're on the master branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "master" ]; then
  echo "Error: Production deployment must be from 'master' branch"
  echo "Current branch: $CURRENT_BRANCH"
  exit 1
fi

echo "Deploying to production from master branch..."

# Run kamal deploy
KAMAL_BIN=$(mise which kamal)
env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" deploy

# Boot BetterStack accessory after deployment
echo "Starting BetterStack logging accessory..."
if env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory details better-stack >/dev/null 2>&1; then
  env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory reboot better-stack
else
  env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory boot better-stack
fi

# Cleanup old Docker build cache
echo "Pruning Docker build cache..."
ssh peti@bandcash "docker buildx prune -af --filter 'until=24h'" || true

echo "Production deployment complete!"
