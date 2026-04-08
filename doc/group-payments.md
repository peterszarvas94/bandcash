# group-payments

## What I do
- Document the split group payment pages and their responsibilities.
- Map each page to routes, templates, and query sources.
- Explain how payment toggle/paid-at actions patch the active page.

## When to use me
Use this when changing group payment tabs, payment table behavior, or paid/unpaid payment flows.

## Pages and routes
- `Pending Payouts`: `GET /groups/:groupId/pending-payouts`
- `Pending Incomes`: `GET /groups/:groupId/pending-incomes`
- `Recent Payouts`: `GET /groups/:groupId/recent-payouts`
- `Recent Incomes`: `GET /groups/:groupId/recent-incomes`

## Route and handler entry points
- Route registration: `models/group/routes.go`
- Page handlers: `ToPayPage`, `ToReceivePage`, `RecentOutgoingPage`, `RecentIncomePage` in `models/group/handlers.go`
- Active page patch selection: `detectPaymentsPageFromReferer` and `renderCurrentPaymentsPageHTML` in `models/group/handlers.go`

## Templates
- `models/group/page_to_pay.templ`
- `models/group/page_to_receive.templ`
- `models/group/page_recent_outgoing.templ`
- `models/group/page_recent_income.templ`
- Main sections:
  - `models/group/component_to_pay_main.templ`
  - `models/group/component_to_receive_main.templ`
  - `models/group/component_recent_outgoing_main.templ`
  - `models/group/component_recent_income_main.templ`

## Data/query sources
- Outgoing (to pay / recent outgoing): `ListUnpaidOutgoingPaymentsByGroup`, `ListPaidOutgoingPaymentsByGroup`
- Event income (to receive / recent income): `ListUnpaidEventsByGroup`, `ListPaidEventsByGroup`
- Query definitions: `internal/db/queries/participants.sql`, `internal/db/queries/events.sql`, `internal/db/queries/expenses.sql`
- Supporting SQL view migration: `internal/db/migrations/033_create_group_outgoing_payments_view.sql`

## Shared UI and table behavior
- Sidebar tabs are defined in `models/shared/component_group_tabs.templ`.
- All pages use table query signals (`mode: "table"` and `tableQuery`) for sort/search/pagination continuity.
- Paid-at dialog state is centralized via `newPaymentsDialogSignals`.

## Mutation endpoints
- Event payment toggles:
  - `PUT /groups/:groupId/payments/events/:id/toggle-paid`
  - `POST /groups/:groupId/payments/events/:id/paid_at`
- Participant payment toggles:
  - `PUT /groups/:groupId/payments/participants/:eventId/:memberId/toggle-paid`
  - `POST /groups/:groupId/payments/participants/:eventId/:memberId/paid_at`
- Expense payment toggles:
  - `PUT /groups/:groupId/payments/expenses/:id/toggle-paid`
  - `POST /groups/:groupId/payments/expenses/:id/paid_at`

## Notes
- Keep payment pages as separate route-based views; do not merge them back into a single payments route.
- For payment mutations, preserve referer query state so the current tab and table state are patched correctly.
- Do not hand-edit generated files (`*_templ.go`, `internal/db/*.sql.go`).
