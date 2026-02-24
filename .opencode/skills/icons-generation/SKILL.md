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
Use the installed `lucide-gen` binary with constrained categories:

```bash
/Users/szarvaspeter/.local/share/mise/installs/go/1.26.0/bin/lucide-gen \
  -output ./models/shared/icons \
  -package icons \
  -categories "navigation,actions,communication,ui,data"
```

## After generation
Always run:

```bash
mise run templ && mise run format && mise run test
```

## Practical workflow
1. Update template usages (`@icons.*`) first.
2. Regenerate reduced icon set with the command above.
3. If compile fails due missing icon names, swap to available icons from `models/shared/icons/registry.templ`.
4. Run templ/format/tests.

## Notes
- This keeps icon files small and avoids slow parsing from 1000+ generated icons.
- Prefer semantic consistency, but prioritize a small generated icon set.
