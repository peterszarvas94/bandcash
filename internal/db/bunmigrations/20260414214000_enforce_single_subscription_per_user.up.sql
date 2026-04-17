CREATE TABLE IF NOT EXISTS billing_subscriptions_single (
    user_id TEXT PRIMARY KEY,
    paddle_subscription_id TEXT NOT NULL UNIQUE,
    paddle_price_id TEXT NOT NULL,
    tier TEXT NOT NULL CHECK (tier IN ('free', 'pro', 'max')),
    status TEXT NOT NULL,
    current_period_ends_at DATETIME,
    grace_until DATETIME,
    canceled_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO billing_subscriptions_single (
    user_id,
    paddle_subscription_id,
    paddle_price_id,
    tier,
    status,
    current_period_ends_at,
    grace_until,
    canceled_at,
    created_at,
    updated_at
)
SELECT
    s.user_id,
    s.paddle_subscription_id,
    s.paddle_price_id,
    s.tier,
    s.status,
    s.current_period_ends_at,
    s.grace_until,
    s.canceled_at,
    s.created_at,
    s.updated_at
FROM billing_subscriptions s
WHERE s.rowid = (
    SELECT s2.rowid
    FROM billing_subscriptions s2
    WHERE s2.user_id = s.user_id
    ORDER BY
      CASE lower(trim(s2.status))
        WHEN 'active' THEN 5
        WHEN 'trialing' THEN 4
        WHEN 'past_due' THEN 3
        WHEN 'paused' THEN 2
        ELSE 1
      END DESC,
      CASE lower(trim(s2.tier))
        WHEN 'max' THEN 3
        WHEN 'pro' THEN 2
        ELSE 1
      END DESC,
      s2.updated_at DESC,
      s2.created_at DESC
    LIMIT 1
);

DROP TABLE billing_subscriptions;
ALTER TABLE billing_subscriptions_single RENAME TO billing_subscriptions;

CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_status ON billing_subscriptions(status);
