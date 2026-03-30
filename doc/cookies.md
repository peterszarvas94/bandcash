# Cookies

Cookie writes are centralized in `internal/utils/cookies.go`.

## Cookie Names

- `session` (`SessionCookieName`)
- `_csrf` (`CSRFCookieName`)
- `locale` (`internal/i18n.LocaleCookieName`)

## Utility Functions

Defined in `internal/utils/cookies.go`:

- `SetSessionCookie(c, token)`
- `ClearSessionCookie(c)`
- `SetCSRFCookie(c, token)`
- `SetLocaleCookie(c, locale)`

Use these helpers instead of ad-hoc `c.SetCookie(...)` in handlers/middleware.

## Locale Cookie Behavior

- Set in locale middleware when `?lang=` is provided.
- Set in auth middleware from resolved preferred language.
- Set in account language update handler after DB update.

All locale writes should use `utils.SetLocaleCookie(...)` to keep policy consistent.

Read path:

- `internal/i18n/i18n.go` (`LocaleFromRequest`) reads the cookie as second priority after query parameter.

## Security Defaults

`SetLocaleCookie` uses:

- `Path=/`
- `MaxAge=365 days`
- `HttpOnly=true`
- `SameSite=Lax`
- `Secure=true` in `production` and `staging`

## Operational Notes

- Locale cookie value is normalized through i18n matcher (`en`, `hu`, fallback to default).
- Session and CSRF cookies are also managed from the same utility file to keep environment/security flags consistent.
- If cookie policy changes are needed (expiry, SameSite, Secure), update utility helpers only.

## Short Summary

- One file owns cookie writes.
- `locale` cookie is set from query/auth/account flows.
- Keep policy changes in utility helpers, not per-handler code.
