## Security

- [ ] add auth abuse protection (cooldown/throttle by IP + email)
- [ ] evaluate 1Password-backed Kamal secrets (`op`/`kamal secrets`) instead of static `.kamal/secrets`, or use varlock

## Improvements

- [ ] support multiple admins per group (user->admin, admin->user, admin kick admin rules)
- [ ] add real-time collaborative editing notice ("Updated by another user [Refresh]")
- [ ] notify logged-in user when added/removed from group; on remove redirect to `/dashboard`
- [ ] add live update on viewer pages when admin changes viewer list (broadcast SSE)
- [ ] save filters
- [ ] cancel events

## Features

- [ ] recurring expenses/events (monthly subscriptions, rent, etc.)
- [ ] export/import CSV for members, events, and expenses
- [ ] audit log per group (who changed what and when)
- [ ] due-date reminders and digest notifications
- [ ] analytics dashboard (member spend, trends, monthly totals)
- [ ] quote maker
- [ ] save quotes as events
