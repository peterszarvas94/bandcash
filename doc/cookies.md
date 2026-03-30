# Cookies

This project centralizes cookie helpers in `internal/utils/cookies.go`.

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

## Locale Cookie Behavior

- Set in locale middleware when `?lang=` is provided.
- Set in auth middleware from resolved preferred language.
- Set in account language update handler after DB update.

All locale writes should use `utils.SetLocaleCookie(...)` to keep policy consistent.

## Security Defaults

`SetLocaleCookie` uses:

- `Path=/`
- `MaxAge=365 days`
- `HttpOnly=true`
- `SameSite=Lax`
- `Secure=true` in `production` and `staging`
