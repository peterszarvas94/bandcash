#!/bin/bash
set -e

usage() {
  echo "Usage: $0 <staging|production>"
  echo ""
  echo "Environment is required:"
  echo "  staging    - Deploy to staging (requires development branch)"
  echo "  production - Deploy to production (requires master branch)"
}

prompt_version_bump() {
  local bump

  if [ ! -t 0 ]; then
    echo "Non-interactive shell detected; skipping version tag."
    VERSION_BUMP="skip"
    return
  fi

  while true; do
    echo ""
    echo "Choose release version bump after successful production deploy:"
    echo "  patch - Increment patch version (x.y.z+1)"
    echo "  minor - Increment minor version (x.y+1.0)"
    echo "  major - Increment major version (x+1.0.0)"
    echo "  skip  - Do not create a tag"
    read -r -p "Version bump [patch/minor/major/skip] (default: skip): " bump || true

    case "${bump:-skip}" in
      patch|minor|major|skip)
        VERSION_BUMP="${bump:-skip}"
        return
        ;;
      *)
        echo "Invalid choice: ${bump}"
        ;;
    esac
  done
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

VERSION_BUMP="skip"
if [ "$ENV" = "production" ]; then
  prompt_version_bump
fi

echo "Deploying to $ENV from $CURRENT_BRANCH branch..."

# Run kamal deploy
kamal deploy $CONFIG_FLAG

# Boot BetterStack accessory after deployment
echo "Starting BetterStack logging accessory..."
if kamal accessory details better-stack $CONFIG_FLAG >/dev/null 2>&1; then
  kamal accessory reboot better-stack $CONFIG_FLAG
else
  kamal accessory boot better-stack $CONFIG_FLAG
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

if [ "$ENV" = "production" ] && [ "$VERSION_BUMP" != "skip" ]; then
  echo "Creating production tag with '$VERSION_BUMP' bump..."
  ./scripts/tag.sh "$VERSION_BUMP"
fi

echo "$ENV deployment complete!"
