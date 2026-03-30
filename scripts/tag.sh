#!/bin/bash
set -e

usage() {
  echo "Usage: $0 <major|minor|patch>"
  echo ""
  echo "Bump type is required:"
  echo "  major - Increment major version (x+1.0.0)"
  echo "  minor - Increment minor version (x.y+1.0)"
  echo "  patch - Increment patch version (x.y.z+1)"
  echo ""
  echo "Tag prefix is selected by branch:"
  echo "  development -> staging-vX.Y.Z"
  echo "  others      -> vX.Y.Z"
}

# Require exactly one argument
if [ $# -ne 1 ]; then
  echo "Error: Bump type is required"
  usage
  exit 1
fi

BUMP_TYPE="$1"

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
TAG_PREFIX="v"
if [ "$CURRENT_BRANCH" = "development" ]; then
  TAG_PREFIX="staging-v"
fi

# Validate bump type
case "$BUMP_TYPE" in
  major|minor|patch)
    # Valid bump type
    ;;
  -h|--help)
    usage
    exit 0
    ;;
  *)
    echo "Error: Invalid bump type: $BUMP_TYPE"
    usage
    exit 1
    ;;
esac

# Get latest tag by branch prefix, default to 0.0.0 if none exists
latest_tag=$(git tag --list "${TAG_PREFIX}*" --sort=-v:refname | sed -n '1p')

if [ -z "$latest_tag" ]; then
  major=0
  minor=0
  patch=0
else
  # Parse version components (assumes format <prefix>X.Y.Z)
  version=${latest_tag#${TAG_PREFIX}}
  major=$(echo "$version" | cut -d. -f1)
  minor=$(echo "$version" | cut -d. -f2)
  patch=$(echo "$version" | cut -d. -f3)
fi

# Bump the appropriate version component
case "$BUMP_TYPE" in
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  patch)
    patch=$((patch + 1))
    ;;
esac

new_tag="${TAG_PREFIX}${major}.${minor}.${patch}"

echo "Current branch: $CURRENT_BRANCH"
echo "Current tag: ${latest_tag:-none}"
echo "Bumping $BUMP_TYPE version..."
echo "Creating tag: $new_tag"

git tag "$new_tag"

echo "Pushing tag to origin..."
git push origin "$new_tag"

echo "Tag $new_tag created and pushed successfully!"
