#!/bin/bash
set -e

usage() {
  echo "Usage: $0 [major|minor|patch] [--skip-tag]"
  echo "Run without bump type to be prompted"
}

# Parse arguments
BUMP_TYPE=""
SKIP_TAG=false

for arg in "$@"; do
  case "$arg" in
    major|minor|patch)
      if [ -n "$BUMP_TYPE" ]; then
        echo "Only one bump type can be provided"
        usage
        exit 1
      fi
      BUMP_TYPE="$arg"
      ;;
    --skip-tag|--no-tag)
      SKIP_TAG=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Invalid argument: $arg"
      usage
      exit 1
      ;;
  esac
done

if [ "$SKIP_TAG" = false ]; then
  # If no bump type provided, prompt user
  if [ -z "$BUMP_TYPE" ]; then
    echo "Select version bump type:"
    echo "  1) major (x+1.0.0)"
    echo "  2) minor (x.x+1.0)"
    echo "  3) patch (x.x.x+1)"
    read -p "Enter choice (1-3): " choice

    case "$choice" in
      1) BUMP_TYPE="major" ;;
      2) BUMP_TYPE="minor" ;;
      3) BUMP_TYPE="patch" ;;
      *)
        echo "Invalid choice: $choice"
        exit 1
        ;;
    esac
  fi

  # Get the latest tag, default to v0.0.0 if none exists
  latest_tag=$(git tag --list 'v*' --sort=-v:refname | head -1)

  if [ -z "$latest_tag" ]; then
    major=0
    minor=0
    patch=0
  else
    # Parse version components (assumes format vX.Y.Z)
    version=${latest_tag#v}
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

  new_tag="v${major}.${minor}.${patch}"

  echo "Bumping $BUMP_TYPE version..."
  echo "Creating tag: $new_tag"
  git tag "$new_tag"

  echo "Pushing tag to origin..."
  git push origin "$new_tag"
else
  echo "Skipping tag creation"
fi

# Run kamal deploy
exec env -u GEM_HOME -u GEM_PATH $(mise which kamal) deploy
