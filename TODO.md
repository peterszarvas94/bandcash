# Design

## General

- remove open/close info, its deprecated

## UI

- check admin pages

# Features

- recurring expenses/events (monthly subscriptions, rent, etc.)
- export/import CSV for members, events, and expenses
- audit log per group (who changed what and when)
- due-date reminders and digest notifications
- analytics dashboard (member spend, trends, monthly totals)
- quote maker
- mark quotes as "accepted", "invoiced", etc.
- save quotes as events
- attach files to quotes / events

# Improvements

- add real-time collaborative editing notice ("Updated by another user [Refresh]")
- notify logged-in user when added/removed from group; on remove redirect to `/dashboard`
- add live update on viewer pages when admin changes viewer list (broadcast SSE)
- save filters
- cancel events

# Billing (Paddle) - New Model

- pricing model: `10 EUR / month / group` (single paid price)
- allow multiple active subscriptions for the same user (owner account)
- env changes:
  - use single `PADDLE_PRICE_ID`
  - keep `PADDLE_CLIENT_TOKEN`, `PADDLE_API_KEY`, `PADDLE_API_BASE_URL`, `PADDLE_WEBHOOK_SECRET`
- webhook scope for v1:
  - process `subscription.created`, `subscription.updated`, `subscription.canceled`
  - count active subscriptions per owner user
- entitlement rule:
  - `allowed_owned_groups = active_subscription_count`
  - group create/ownership transfer blocked when owned groups would exceed purchased count
- account page updates:
  - show `purchased` (active subscriptions)
  - show `used` (owned groups)
  - show `remaining = purchased - used`
  - show CTA to buy another subscription slot
- pricing page updates:
  - remove Free/Pro/Max cards
  - single checkout CTA using `data-paddle-price-id={PADDLE_PRICE_ID}`
- transfer ownership rules:
  - receiver must have free slot (`used < purchased`) before accepting transfer
  - re-check at acceptance time
- cleanup old tier system:
  - remove tier-based middleware/messages/i18n
  - replace with slot/quota wording (`subscription slots`)
