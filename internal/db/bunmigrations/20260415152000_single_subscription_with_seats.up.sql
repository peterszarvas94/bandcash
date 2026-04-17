ALTER TABLE billing_subscriptions ADD COLUMN provider_subscription_item_id TEXT;
ALTER TABLE billing_subscriptions ADD COLUMN seat_quantity INTEGER NOT NULL DEFAULT 1;
ALTER TABLE billing_subscriptions ADD COLUMN provider_portal_url TEXT;
ALTER TABLE billing_subscriptions ADD COLUMN provider_update_payment_url TEXT;

UPDATE billing_subscriptions
SET seat_quantity = 1
WHERE seat_quantity IS NULL OR seat_quantity < 1;

DROP INDEX IF EXISTS idx_billing_subscriptions_user_id;
CREATE UNIQUE INDEX IF NOT EXISTS idx_billing_subscriptions_user_id ON billing_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_provider_subscription_item_id ON billing_subscriptions(provider_subscription_item_id);
