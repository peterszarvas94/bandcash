CREATE TABLE IF NOT EXISTS billing_customers (
    user_id TEXT PRIMARY KEY,
    paddle_customer_id TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_billing_customers_paddle_customer_id ON billing_customers(paddle_customer_id);

CREATE TABLE IF NOT EXISTS billing_subscriptions (
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
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_status ON billing_subscriptions(status);

CREATE TABLE IF NOT EXISTS billing_webhook_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
