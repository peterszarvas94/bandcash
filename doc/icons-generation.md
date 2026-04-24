# icons-generation

## What I do
- Keep icon generation small and fast.
- Avoid generating the full Lucide set.
- Document the safe regeneration workflow for this repo.

## When to use me
Use this whenever you add/remove icons or touch `models/shared/icons/*`.

## Repo policy
- Do **not** generate all icons.
- Generate only icons used in the app.
- Keep `registry.templ` generation enabled (the app uses `icons.IconName` and `icons.Icon...` constants).
- Skip `categories.go` generation.
- Keep icon package at `models/shared/icons`.

## Generator command
Use `lucide-gen` with an explicit icon list:

```bash
lucide-gen \
  -output ./models/shared/icons \
  -package icons \
  -icons "<comma-separated-used-icons>" \
  -skip-categories
```

Example:

```bash
lucide-gen \
  -output ./models/shared/icons \
  -package icons \
  -icons "arrow-left,arrow-right,save,search,x" \
  -skip-categories
```

## After generation
Always run:

```bash
mise run format && mise run test
```

Then regenerate icon Go files:

```bash
templ generate "models/shared/icons/*.templ"
```

## Practical workflow
1. Update template usages (`@icons.*`) first.
2. Scan repo usages (`icons.Name(...)` and `icons.IconName` constants) and build an icon list.
3. Regenerate with `-icons "...list..." -skip-categories`.
4. Regenerate templ outputs for `models/shared/icons/*.templ`.
5. Run `go build ./cmd/server`; if symbols are missing, add the missing icon names and regenerate.
6. Run format/tests.

## Notes
- This keeps icon files small and avoids slow parsing from 1000+ generated icons.
- `-skip-registry` is currently **not** compatible with this app because shared components and handlers still use registry-backed icon types/constants.
- Prefer semantic consistency, but prioritize a small generated icon set.
