# modals

## Why this exists
- This app uses SSE + Datastar morphing, so modal lifecycle can break if DOM nodes are replaced while a modal is open.
- We use native `<dialog>` for accessibility/focus trap, with one shared wrapper and consistent signal flow.

## Current standard
- Use native `<dialog>` (not Popover API) for app modals.
- Use `@shared.DialogShell(...)` as the only wrapper.
- Keep modal state in page signals (`open`, `fetching`, values/labels).
- Use route pair: `GET` opens modal state, `POST` saves.

## Shared wrapper contract
- Wrapper: `models/shared/component_dialog.templ` (`DialogShell`).
- `DialogShell` sets `data-ignore-morph` on the `<dialog>` root.
  - This is critical: it prevents Datastar morph from replacing an open dialog node and leaving a stuck backdrop.
- `DialogProps` provides common expressions:
  - `EffectExpr` -> show/close behavior
  - `OverlayCloseExpr` -> click outside to close
  - `CancelExpr` -> prevent native cancel while fetching
  - `CloseExpr` -> normalize signal reset on close
  - `FetchDoneExpr` -> reset `fetching/open` after request

## How to build a modal (feature-level)
1. Add dialog state to `<feature>/page_data.go`.
2. Expose it in `<feature>/signals.go`.
3. Create a dialog template using `@shared.DialogShell(...)`.
4. Add open button(s) in page table/card.
5. Add routes:
   - `GET ...` open modal (set state + patch html/signals)
   - `POST ...` save (mutate + patch html/signals + notify)
6. Keep submit button in modal using Datastar action (`@post(...)`) and `fetching` guard.

## Route shape used in events
- Event-level paid at:
  - `GET /groups/:groupId/events/:id/paid_at`
  - `POST /groups/:groupId/events/:id/paid_at`
- Member-scoped fields inside event:
  - `GET /groups/:groupId/events/:id/members/:memberId/note`
  - `POST /groups/:groupId/events/:id/members/:memberId/note`
  - `GET /groups/:groupId/events/:id/members/:memberId/paid_at`
  - `POST /groups/:groupId/events/:id/members/:memberId/paid_at`
  - `POST /groups/:groupId/events/:id/members/:memberId/paid`

## Dialog vs Popover in this app
- Prefer `<dialog>` when you need native focus trap and keyboard behavior.
- Popover API is simpler to toggle, but lacks modal focus-trap semantics.
- Because this app morphs aggressively, `<dialog>` must be paired with `data-ignore-morph` to stay stable.

## Gotchas
- Do not mix popup implementations per page (dialog + popover) unless truly needed.
- Do not bind submit/loading to global page flags; use modal-specific `fetching`.
- Keep close/reset logic in one place (wrapper expressions), not scattered in buttons.
- If a modal appears once then blocks clicks, first verify dialog root was not morphed (check `data-ignore-morph`).

## Troubleshooting checklist
1. Check latest log file first (`doc/logs.md`).
2. Verify `GET open` route returns 200 and sets dialog signal `open=true`.
3. Verify `POST save` returns 200 and resets `fetching/open`.
4. Confirm modal root uses `DialogShell` and has `data-ignore-morph`.
5. Confirm page patch does not remove/recreate dialog node.
