---
name: icons-generation
description: Generate only the needed Lucide templ icons for this repo.
---

## What I do
- Keep icon generation small and fast.
- Avoid generating the full Lucide set.
- Document the safe regeneration workflow for this repo.

## When to use me
Use this whenever you add/remove icons or touch `models/shared/icons/*`.

## Repo policy
- Do **not** generate all icons.
- Generate a reduced category set only.
- Keep icon package at `models/shared/icons`.

## Generator command
Use `lucide-gen` with constrained categories:

```bash
lucide-gen \
  -output ./models/shared/icons \
  -package icons \
  -categories "navigation,actions,communication,ui,data"
```

## After generation
Always run:

```bash
mise run format && mise run test
```

Only run `mise run templ` when explicitly requested.

## Practical workflow
1. Update template usages (`@icons.*`) first.
2. Regenerate reduced icon set with the command above.
3. If compile fails due missing icon names, swap to available icons from `models/shared/icons/registry.templ`.
4. Run format/tests.

## Notes
- This keeps icon files small and avoids slow parsing from 1000+ generated icons.
- Prefer semantic consistency, but prioritize a small generated icon set.
