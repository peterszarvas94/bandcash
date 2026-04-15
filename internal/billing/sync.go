package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"bandcash/internal/db"
	authstore "bandcash/models/auth/data"
)

type WebhookSubscriptionUpdate struct {
	EventID             string
	EventType           string
	SubscriptionID      string
	CustomerID          string
	VariantID           string
	Status              string
	UserID              string
	CustomerEmail       string
	CurrentPeriodEndsAt sql.NullTime
	CanceledAt          sql.NullTime
	GraceUntil          sql.NullTime
}

var ErrWebhookUserNotResolved = errors.New("webhook user could not be resolved")

func resolveUserIDBySubscriptionID(ctx context.Context, subscriptionID string) (string, error) {
	if strings.TrimSpace(subscriptionID) == "" {
		return "", nil
	}
	var userID string
	err := db.BunDB.QueryRowContext(ctx,
		"SELECT user_id FROM billing_subscriptions WHERE provider_subscription_id = ? LIMIT 1",
		strings.TrimSpace(subscriptionID),
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(userID), nil
}

func VerifyWebhookSignature(rawBody []byte, signatureHeader, endpointSecret string) bool {
	if len(rawBody) == 0 {
		return false
	}
	if strings.TrimSpace(signatureHeader) == "" || strings.TrimSpace(endpointSecret) == "" {
		return false
	}

	h := hmac.New(sha256.New, []byte(endpointSecret))
	_, _ = h.Write(rawBody)
	expected := strings.ToLower(hex.EncodeToString(h.Sum(nil)))
	candidate := strings.ToLower(strings.TrimSpace(signatureHeader))
	return subtle.ConstantTimeCompare([]byte(candidate), []byte(expected)) == 1
}

func markWebhookProcessed(ctx context.Context, eventID, eventType string) (bool, error) {
	if strings.TrimSpace(eventID) == "" {
		return false, errors.New("missing event id")
	}
	_, err := db.BunDB.ExecContext(ctx,
		"INSERT INTO billing_webhook_events(event_id, event_type) VALUES (?, ?)",
		strings.TrimSpace(eventID),
		strings.TrimSpace(eventType),
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ResolveUserIDForWebhook(ctx context.Context, userID, customerID, customerEmail string) (string, error) {
	userID = strings.TrimSpace(userID)
	customerID = strings.TrimSpace(customerID)
	customerEmail = strings.ToLower(strings.TrimSpace(customerEmail))

	if userID != "" {
		if _, err := authstore.GetUserByID(ctx, userID); err == nil {
			if customerID != "" {
				if upsertErr := UpsertCustomer(ctx, userID, customerID); upsertErr != nil {
					return "", upsertErr
				}
			}
			return userID, nil
		}
	}

	if customerID != "" {
		var mappedUserID string
		err := db.BunDB.QueryRowContext(ctx,
			"SELECT user_id FROM billing_customers WHERE provider_customer_id = ? LIMIT 1",
			customerID,
		).Scan(&mappedUserID)
		if err == nil && strings.TrimSpace(mappedUserID) != "" {
			return strings.TrimSpace(mappedUserID), nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}

	if customerEmail != "" {
		user, err := authstore.GetUserByEmail(ctx, customerEmail)
		if err == nil {
			if customerID != "" {
				if upsertErr := UpsertCustomer(ctx, user.ID, customerID); upsertErr != nil {
					return "", upsertErr
				}
			}
			return user.ID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}

	return "", nil
}

func UpsertCustomer(ctx context.Context, userID, customerID string) error {
	userID = strings.TrimSpace(userID)
	customerID = strings.TrimSpace(customerID)
	if userID == "" || customerID == "" {
		return nil
	}
	result, err := db.BunDB.ExecContext(ctx,
		"UPDATE billing_customers SET provider_customer_id = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?",
		customerID,
		userID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	_, err = db.BunDB.ExecContext(ctx,
		"INSERT INTO billing_customers(user_id, provider_customer_id) VALUES (?, ?)",
		userID,
		customerID,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			_, updateErr := db.BunDB.ExecContext(ctx,
				"UPDATE billing_customers SET provider_customer_id = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?",
				customerID,
				userID,
			)
			return updateErr
		}
		return err
	}
	return nil
}

func UpsertSubscription(ctx context.Context, update WebhookSubscriptionUpdate) error {
	status := strings.ToLower(strings.TrimSpace(update.Status))
	if strings.TrimSpace(update.VariantID) == "" {
		var existingPriceID string
		err := db.BunDB.QueryRowContext(ctx,
			"SELECT provider_variant_id FROM billing_subscriptions WHERE provider_subscription_id = ? LIMIT 1",
			strings.TrimSpace(update.SubscriptionID),
		).Scan(&existingPriceID)
		if err == nil {
			update.VariantID = strings.TrimSpace(existingPriceID)
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	tier := TierFromPriceID(update.VariantID)
	graceUntil := update.GraceUntil
	if status == "past_due" && !graceUntil.Valid {
		graceUntil = sql.NullTime{Time: time.Now().UTC().Add(PastDueGracePeriod), Valid: true}
	}

	row := db.BillingSubscription{
		UserID:                 strings.TrimSpace(update.UserID),
		ProviderSubscriptionID: strings.TrimSpace(update.SubscriptionID),
		ProviderVariantID:      strings.TrimSpace(update.VariantID),
		Tier:                   tier,
		Status:                 status,
		CurrentPeriodEndsAt:    update.CurrentPeriodEndsAt,
		GraceUntil:             graceUntil,
		CanceledAt:             update.CanceledAt,
	}

	result, err := db.BunDB.ExecContext(ctx,
		"UPDATE billing_subscriptions SET user_id = ?, provider_variant_id = ?, tier = ?, status = ?, current_period_ends_at = ?, grace_until = ?, canceled_at = ?, updated_at = CURRENT_TIMESTAMP WHERE provider_subscription_id = ?",
		row.UserID,
		row.ProviderVariantID,
		row.Tier,
		row.Status,
		row.CurrentPeriodEndsAt,
		row.GraceUntil,
		row.CanceledAt,
		row.ProviderSubscriptionID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	_, err = db.BunDB.ExecContext(ctx,
		"INSERT INTO billing_subscriptions(provider_subscription_id, user_id, provider_variant_id, tier, status, current_period_ends_at, grace_until, canceled_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		row.ProviderSubscriptionID,
		row.UserID,
		row.ProviderVariantID,
		row.Tier,
		row.Status,
		row.CurrentPeriodEndsAt,
		row.GraceUntil,
		row.CanceledAt,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			_, updateErr := db.BunDB.ExecContext(ctx,
				"UPDATE billing_subscriptions SET user_id = ?, provider_variant_id = ?, tier = ?, status = ?, current_period_ends_at = ?, grace_until = ?, canceled_at = ?, updated_at = CURRENT_TIMESTAMP WHERE provider_subscription_id = ?",
				row.UserID,
				row.ProviderVariantID,
				row.Tier,
				row.Status,
				row.CurrentPeriodEndsAt,
				row.GraceUntil,
				row.CanceledAt,
				row.ProviderSubscriptionID,
			)
			return updateErr
		}
		return err
	}
	return nil
}

func isUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}

func parseTime(value string) sql.NullTime {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullTime{}
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return sql.NullTime{Time: t.UTC(), Valid: true}
		}
	}
	return sql.NullTime{}
}

func getValueByPath(data map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = data
	for _, part := range parts {
		name := part
		idx := -1
		if left := strings.Index(part, "["); left > 0 && strings.HasSuffix(part, "]") {
			name = part[:left]
			rawIdx := strings.TrimSuffix(part[left+1:], "]")
			parsedIdx, err := strconv.Atoi(rawIdx)
			if err == nil {
				idx = parsedIdx
			}
		}

		obj, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current, ok = obj[name]
		if !ok {
			return nil
		}

		if idx >= 0 {
			arr, ok := current.([]any)
			if !ok || idx >= len(arr) || idx < 0 {
				return nil
			}
			current = arr[idx]
		}
	}
	return current
}

func firstString(data map[string]any, paths ...string) string {
	for _, path := range paths {
		value := getValueByPath(data, path)
		s, ok := value.(string)
		if ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func statusFromEventType(eventType string) string {
	switch strings.TrimSpace(strings.ToLower(eventType)) {
	case "subscription_created", "subscription_resumed", "subscription_unpaused":
		return "active"
	case "subscription_cancelled", "subscription_expired":
		return "canceled"
	case "subscription_paused":
		return "paused"
	case "subscription_payment_failed":
		return "past_due"
	default:
		return ""
	}
}

func ParseWebhookSubscription(rawBody []byte) (WebhookSubscriptionUpdate, bool, error) {
	var root map[string]any
	if err := json.Unmarshal(rawBody, &root); err != nil {
		return WebhookSubscriptionUpdate{}, false, err
	}

	eventType := strings.ToLower(firstString(root, "meta.event_name", "event_name", "event_type", "type"))
	eventID := firstString(root, "meta.event_id", "meta.webhook_id", "event_id", "id", "notification_id")
	if eventID == "" {
		hash := sha256.Sum256(rawBody)
		eventID = hex.EncodeToString(hash[:])
	}
	if eventType == "" {
		return WebhookSubscriptionUpdate{}, false, errors.New("missing event type")
	}

	if !strings.HasPrefix(eventType, "subscription_") {
		return WebhookSubscriptionUpdate{EventID: eventID, EventType: eventType}, false, nil
	}

	dataAny := root["data"]
	data, ok := dataAny.(map[string]any)
	if !ok {
		return WebhookSubscriptionUpdate{}, false, errors.New("missing data payload")
	}

	priceID := firstString(data,
		"attributes.variant_id",
		"attributes.first_subscription_item.variant_id",
		"items[0].price.id",
		"items[0].price_id",
		"price.id",
		"price_id",
	)

	update := WebhookSubscriptionUpdate{
		EventID:             eventID,
		EventType:           eventType,
		SubscriptionID:      firstString(data, "id", "attributes.id", "subscription_id"),
		CustomerID:          firstString(data, "attributes.customer_id", "customer_id", "customer.id"),
		VariantID:           priceID,
		Status:              strings.ToLower(firstString(data, "attributes.status", "status")),
		UserID:              firstString(root, "meta.custom_data.user_id", "meta.custom_data.userID", "data.attributes.custom_data.user_id", "custom_data.user_id", "metadata.user_id"),
		CustomerEmail:       strings.ToLower(firstString(data, "attributes.user_email", "attributes.customer_email", "customer.email", "customer_email", "email")),
		CurrentPeriodEndsAt: parseTime(firstString(data, "attributes.renews_at", "attributes.ends_at", "current_billing_period.ends_at", "next_billed_at", "scheduled_change.effective_at")),
		CanceledAt:          parseTime(firstString(data, "attributes.cancelled", "attributes.ends_at", "canceled_at")),
		GraceUntil:          parseTime(firstString(root, "meta.custom_data.grace_until", "data.attributes.custom_data.grace_until", "custom_data.grace_until")),
	}
	if update.Status == "" {
		update.Status = statusFromEventType(eventType)
	}

	if update.SubscriptionID == "" {
		return WebhookSubscriptionUpdate{}, false, errors.New("missing subscription id")
	}

	return update, true, nil
}

func ProcessWebhook(ctx context.Context, rawBody []byte) (bool, error) {
	update, isSubscriptionEvent, err := ParseWebhookSubscription(rawBody)
	if err != nil {
		return false, err
	}

	if !isSubscriptionEvent {
		inserted, err := markWebhookProcessed(ctx, update.EventID, update.EventType)
		if err != nil {
			return false, err
		}
		if !inserted {
			return false, nil
		}
		return true, nil
	}

	userID, err := ResolveUserIDForWebhook(ctx, update.UserID, update.CustomerID, update.CustomerEmail)
	if err != nil {
		return false, err
	}
	if userID == "" {
		userID, err = resolveUserIDBySubscriptionID(ctx, update.SubscriptionID)
		if err != nil {
			return false, err
		}
	}
	if userID == "" {
		return false, ErrWebhookUserNotResolved
	}

	update.UserID = userID
	if err := UpsertCustomer(ctx, update.UserID, update.CustomerID); err != nil {
		return false, err
	}
	if err := UpsertSubscription(ctx, update); err != nil {
		return false, err
	}

	inserted, err := markWebhookProcessed(ctx, update.EventID, update.EventType)
	if err != nil {
		return false, err
	}
	if !inserted {
		return false, nil
	}
	return true, nil
}
