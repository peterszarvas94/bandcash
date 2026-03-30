# Language Features

This document describes locale resolution and sync behavior across public and authenticated pages.

## Resolution Order

Locale is resolved in this order:

1. `?lang=<code>` query param (if present and valid)
2. `locale` cookie
3. `Accept-Language` header
4. default locale (`hu`)

Source: `internal/i18n/i18n.go` (`LocaleFromRequest`).

## Persistence Rules

- **Guest users**
  - Public language picker submits `GET` with `lang`.
  - Locale middleware stores that value in the `locale` cookie.
  - Next pages can omit `lang` in links and still keep locale.

- **Authenticated users**
  - Auth middleware resolves locale from `preferred_lang` with optional query override.
  - Resolved locale is written to `locale` cookie on each authenticated request.
  - Account language update writes both DB (`preferred_lang`) and `locale` cookie.

- **Authenticated users on public pages**
  - Home/auth handlers sync `?lang` into `preferred_lang` so changes on `/`, `/login`, and legal pages persist into account/group pages.

## Main Integration Points

- Locale resolution: `internal/i18n/i18n.go`
- Locale cookie write on query override: `internal/middleware/locale.go`
- Auth locale binding + cookie sync: `internal/middleware/auth.go`
- Account language save: `models/account/handlers.go` (`UpdateLanguage`)
- Public language picker UI: `models/shared/component_page_language_picker.templ`
- Public-page DB sync helpers:
  - `models/home/handlers.go`
  - `models/auth/handlers.go`

## UX Notes

- Shared short language labels are `ENG` / `HUN`.
- Public and account selectors use `input-xs` sizing.
- Header/footer links no longer need explicit `?lang` propagation for normal navigation.

## Short Summary

- Guest persistence: query -> locale cookie.
- Auth persistence: query/user pref -> locale cookie + DB pref.

## Troubleshooting

- If locale reverts between pages, verify the `locale` cookie is present and path is `/`.
- If locale changes are not reflected on authenticated group/account pages, verify `preferred_lang` update happened (`users.preferred_lang`).
- If a deep link should force locale, use explicit `?lang=...` once; cookie and DB sync will follow.
