# Design

## General

- prefer drop shadows over borders for card-like elements (detail sections, tables, etc.)

## Color

- use blue for primary color
- find a good accent color (e.g. orange)

## Dashboard

- make groups as cards
- add button should be bigger and marked with primary color

## General for all pages

- improve details section
- make name as title not as detail entry
- move paid/unpaid/all radio into normal table filter, like search or date filters
- replace sidebar forms with native popups (dialog)

## UI

- check admin pages
- figure out tab styling (in both admin and group pages)

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
