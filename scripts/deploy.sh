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
    CONFIG_FLAG=""
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
KAMAL_BIN=kamal
"$KAMAL_BIN" deploy $CONFIG_FLAG

# Boot BetterStack accessory after deployment
echo "Starting BetterStack logging accessory..."
if "$KAMAL_BIN" accessory details better-stack $CONFIG_FLAG >/dev/null 2>&1; then
  "$KAMAL_BIN" accessory reboot better-stack $CONFIG_FLAG
else
  "$KAMAL_BIN" accessory boot better-stack $CONFIG_FLAG
fi

# Cleanup Docker BuildKit cache after deployment
ssh peti@"$SERVER_SSH" "
BUILDKIT_CONTAINER=buildx_buildkit_kamal-remote-ssh---peti-${SERVER_SSH}0

if ! docker exec \"\${BUILDKIT_CONTAINER}\" true >/dev/null 2>&1; then
  exit 0
fi

BEFORE_MB=\$(docker exec \"\${BUILDKIT_CONTAINER}\" sh -lc 'du -sm /var/lib/buildkit | cut -f1')
docker exec \"\${BUILDKIT_CONTAINER}\" buildctl prune --all --keep-storage 6144 >/dev/null
AFTER_MB=\$(docker exec \"\${BUILDKIT_CONTAINER}\" sh -lc 'du -sm /var/lib/buildkit | cut -f1')
echo \"BuildKit cache: \${BEFORE_MB}MB -> \${AFTER_MB}MB (freed \$((BEFORE_MB - AFTER_MB))MB)\"
" || true

echo "$ENV deployment complete!"
