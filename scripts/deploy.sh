#!/bin/bash
set -e

usage() {
  echo "Usage: $0 <staging|production>"
  echo ""
  echo "Environment is required:"
  echo "  staging    - Deploy to staging (requires development branch)"
  echo "  production - Deploy to production (requires master branch)"
}

# Require exactly one argument
if [ $# -ne 1 ]; then
  echo "Error: Environment is required"
  usage
  exit 1
fi

ENV="$1"

# Validate environment and set configuration
case "$ENV" in
  staging)
    REQUIRED_BRANCH="development"
    SERVER_SSH="bandcash_staging"
    CONFIG_FLAG="--destination staging"
    ;;
  production)
    REQUIRED_BRANCH="master"
    SERVER_SSH="bandcash"
    CONFIG_FLAG="--destination production"
    ;;
  -h|--help)
    usage
    exit 0
    ;;
  *)
    echo "Error: Invalid environment: $ENV"
    usage
    exit 1
    ;;
esac

# Check if we're on the required branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "$REQUIRED_BRANCH" ]; then
  echo "Error: $ENV deployment must be from '$REQUIRED_BRANCH' branch"
  echo "Current branch: $CURRENT_BRANCH"
  exit 1
fi

echo "Deploying to $ENV from $CURRENT_BRANCH branch..."

# Run kamal deploy
KAMAL_BIN=$(mise which kamal)
env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" deploy $CONFIG_FLAG

# Boot BetterStack accessory after deployment
echo "Starting BetterStack logging accessory..."
if env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory details better-stack $CONFIG_FLAG >/dev/null 2>&1; then
  env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory reboot better-stack $CONFIG_FLAG
else
  env -u GEM_HOME -u GEM_PATH "$KAMAL_BIN" accessory boot better-stack $CONFIG_FLAG
fi

# Cleanup old Docker build cache
echo "Pruning Docker build cache..."
ssh peti@"$SERVER_SSH" "docker buildx prune -af --filter 'until=24h'" || true

echo "$ENV deployment complete!"
