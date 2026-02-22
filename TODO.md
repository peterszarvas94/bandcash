# P0

- [x] add rate limiting (global + stricter auth endpoints)
- [ ] enforce request body size limits
- [ ] normalize + strictly validate all inputs server-side
- [ ] add auth abuse protection (cooldown/throttle by IP + email)
- [ ] enable security checks in CI (`govulncheck`, dependency updates)
- [ ] add monitoring + alerting for `/health`

# P1

- [ ] remove unnecessary sidebars (for example read-only users)
- [ ] notify logged-in user when added/removed from group; on remove redirect to `/dashboard`
- [ ] add live update on viewer pages when admin changes viewer list (broadcast SSE)
- [ ] improve mobile friendliness

# P2

- [ ] add real-time collaborative editing notice ("Updated by another user [Refresh]")
- [ ] add expenses table (title, description, amount, date)
- [ ] support multiple admins per group (user->admin, admin->user, admin kick admin rules)
