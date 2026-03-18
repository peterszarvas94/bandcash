#!/bin/bash
set -e

# Get the latest tag, default to v0.0.0 if none exists
latest_tag=$(git tag --list 'v*' --sort=-v:refname | head -1)

if [ -z "$latest_tag" ]; then
  new_tag="v0.0.1"
else
  # Parse version components (assumes format vX.Y.Z)
  version=${latest_tag#v}
  major=$(echo "$version" | cut -d. -f1)
  minor=$(echo "$version" | cut -d. -f2)
  patch=$(echo "$version" | cut -d. -f3)
  
  # Increment patch version
  new_patch=$((patch + 1))
  new_tag="v${major}.${minor}.${new_patch}"
fi

echo "Creating tag: $new_tag"
git tag "$new_tag"

echo "Pushing tag to origin..."
git push origin "$new_tag"

# Run kamal deploy
exec env -u GEM_HOME -u GEM_PATH $(mise which kamal) deploy
