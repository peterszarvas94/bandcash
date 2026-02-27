# P1

- [x] loading indicator for forms and buttons
- [x] remove unnecessary sidebars (for example read-only users)
- [ ] pagination for tables
- [ ] search for tables
- [ ] filter for tables
- [ ] order for tables
- [ ] improve mobile friendliness
- [x] add monitoring + alerting for `/health` (defer with centralized logs/observability)
- [ ] add expenses table (title, description, amount, date)

# P2

- [ ] notify logged-in user when added/removed from group; on remove redirect to `/dashboard`
- [ ] add live update on viewer pages when admin changes viewer list (broadcast SSE)
- [ ] add auth abuse protection (cooldown/throttle by IP + email)
- [ ] evaluate 1Password-backed Kamal secrets (`op`/`kamal secrets`) instead of static `.kamal/secrets`
- [ ] add real-time collaborative editing notice ("Updated by another user [Refresh]")
- [ ] support multiple admins per group (user->admin, admin->user, admin kick admin rules)
