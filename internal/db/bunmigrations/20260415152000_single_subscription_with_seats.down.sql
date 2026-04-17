CREATE TABLE IF NOT EXISTS billing_subscriptions_old (
    provider_subscription_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    provider_variant_id TEXT NOT NULL,
    tier TEXT NOT NULL CHECK (tier IN ('free', 'pro', 'max')),
    status TEXT NOT NULL,
    current_period_ends_at DATETIME,
    grace_until DATETIME,
    canceled_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO billing_subscriptions_old (
    provider_subscription_id,
    user_id,
    provider_variant_id,
    tier,
    status,
    current_period_ends_at,
    grace_until,
    canceled_at,
    created_at,
    updated_at
)
SELECT
    provider_subscription_id,
    user_id,
    provider_variant_id,
    tier,
    status,
    current_period_ends_at,
    grace_until,
    canceled_at,
    created_at,
    updated_at
FROM billing_subscriptions;

DROP TABLE billing_subscriptions;
ALTER TABLE billing_subscriptions_old RENAME TO billing_subscriptions;

DROP INDEX IF EXISTS idx_billing_subscriptions_provider_subscription_item_id;
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_user_id ON billing_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_status ON billing_subscriptions(status);
