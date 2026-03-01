# Skill: ui-feedback

## Goal
Apply one strict feedback model across the app.

## Rules
- Inline errors are only for form validation errors (user-correctable input, usually 422).
- Notifications are for everything else: success, info, and non-validation errors.
- Do not show the same failure in both inline and notification channels.
- Do not use page-level `notice/error` blocks or query-param flash messages (`?msg`, `?error`).

## Implementation pattern
1. Validation failure:
   - Keep/patch form signals like `errors`, `authError`.
   - Return `422`.
2. Non-validation failure:
   - `utils.Notify(c, "error", ctxi18n.T(...))`.
   - Return/redirect with normal flow.
3. Success:
   - `utils.Notify(c, "success"|"info", ctxi18n.T(...))`.
   - Continue with patch/redirect.

## Rendering requirements
- Every top-level page template should include `@shared.Notifications()`.
- For handlers rendering full pages, use `utils.RenderPage(...)`.
- For server-side patch rendering, use `utils.RenderHTMLForRequest(...)` so notification queues are drained.

## i18n
- Notification text must use locale keys from `internal/i18n/locales/en/en.yaml` and `internal/i18n/locales/hu/hu.yaml`.
- Do not hardcode notification strings in handlers.
