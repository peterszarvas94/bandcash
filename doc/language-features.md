# Language Features

This document describes how language selection works across public and authenticated pages.

## Resolution Order

Locale is resolved in this order:

1. `?lang=<code>` query param (if present and valid)
2. `locale` cookie
3. `Accept-Language` header
4. default locale (`hu`)

Implementation: `internal/i18n/i18n.go` (`LocaleFromRequest`).

## Public Pages

- Public pages use a per-page language picker (`GET` form with `lang` query).
- When `lang` is present, locale middleware persists it into the `locale` cookie.
- Navigation can stay clean (no `lang` propagation required on every link), because cookie fallback keeps the selected locale.

Implementation:

- `internal/middleware/locale.go`
- `models/shared/component_page_language_picker.templ`

## Authenticated Users

- Auth middleware resolves locale from user preference (`preferred_lang`) with optional query override.
- Resolved locale is written into the `locale` cookie for consistency across routes.
- Account language update persists to DB and updates the cookie immediately.

Implementation:

- `internal/middleware/auth.go`
- `models/account/handlers.go` (`UpdateLanguage`)

## UI Notes

- Shared short language labels use `ENG` / `HUN`.
- Account language selector and public page picker use compact input sizing (`input-xs`).
