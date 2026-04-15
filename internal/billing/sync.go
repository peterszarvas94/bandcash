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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bandcash/internal/db"
	"bandcash/internal/utils"
	authstore "bandcash/models/auth/data"
)

type WebhookSubscriptionUpdate struct {
	EventID              string
	EventType            string
	PaddleSubscriptionID string
	PaddleCustomerID     string
	PaddlePriceID        string
	Status               string
	UserID               string
	CustomerEmail        string
	CurrentPeriodEndsAt  sql.NullTime
	CanceledAt           sql.NullTime
	GraceUntil           sql.NullTime
}

var ErrWebhookUserNotResolved = errors.New("webhook user could not be resolved")

func resolveUserIDBySubscriptionID(ctx context.Context, subscriptionID string) (string, error) {
	if strings.TrimSpace(subscriptionID) == "" {
		return "", nil
	}
	var userID string
	err := db.BunDB.QueryRowContext(ctx,
		"SELECT user_id FROM billing_subscriptions WHERE paddle_subscription_id = ? LIMIT 1",
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

func VerifyWebhookSignature(rawBody []byte, signatureHeader, endpointSecret string, tolerance time.Duration) bool {
	if len(rawBody) == 0 {
		return false
	}
	if strings.TrimSpace(signatureHeader) == "" || strings.TrimSpace(endpointSecret) == "" {
		return false
	}

	parts := strings.Split(signatureHeader, ";")
	values := map[string][]string{}
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" || value == "" {
			continue
		}
		values[key] = append(values[key], value)
	}

	tsRaw := ""
	if tsValues := values["ts"]; len(tsValues) > 0 {
		tsRaw = tsValues[0]
	}
	if tsRaw == "" {
		return false
	}

	timestamp, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		return false
	}

	if tolerance > 0 {
		now := time.Now().UTC()
		eventTime := time.Unix(timestamp, 0).UTC()
		delta := now.Sub(eventTime)
		if delta < 0 {
			delta = -delta
		}
		if delta > tolerance {
			return false
		}
	}

	signedPayload := fmt.Sprintf("%s:%s", tsRaw, string(rawBody))
	h := hmac.New(sha256.New, []byte(endpointSecret))
	_, _ = h.Write([]byte(signedPayload))
	expected := strings.ToLower(hex.EncodeToString(h.Sum(nil)))

	for _, candidate := range values["h1"] {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate == "" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(expected)) == 1 {
			return true
		}
	}

	return false
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
			"SELECT user_id FROM billing_customers WHERE paddle_customer_id = ? LIMIT 1",
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

	if customerID != "" {
		paddleEmail, err := FetchCustomerEmailFromPaddle(ctx, customerID)
		if err != nil {
			return "", err
		}
		if paddleEmail != "" {
			user, err := authstore.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(paddleEmail)))
			if err == nil {
				if upsertErr := UpsertCustomer(ctx, user.ID, customerID); upsertErr != nil {
					return "", upsertErr
				}
				return user.ID, nil
			}
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return "", err
			}
		}
	}

	return "", nil
}

func FetchCustomerEmailFromPaddle(ctx context.Context, customerID string) (string, error) {
	apiKey := strings.TrimSpace(utils.Env().PaddleAPIKey)
	if apiKey == "" {
		return "", nil
	}
	apiBaseURL := strings.TrimSpace(utils.Env().PaddleAPIBaseURL)
	if apiBaseURL == "" {
		return "", errors.New("PADDLE_API_BASE_URL is required for Paddle API calls")
	}
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return "", nil
	}

	endpoint := strings.TrimRight(apiBaseURL, "/") + "/customers/" + url.PathEscape(customerID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("paddle customer fetch failed status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Data struct {
			Email string `json:"email"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(payload.Data.Email)), nil
}

func UpsertCustomer(ctx context.Context, userID, customerID string) error {
	userID = strings.TrimSpace(userID)
	customerID = strings.TrimSpace(customerID)
	if userID == "" || customerID == "" {
		return nil
	}
	result, err := db.BunDB.ExecContext(ctx,
		"UPDATE billing_customers SET paddle_customer_id = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?",
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
		"INSERT INTO billing_customers(user_id, paddle_customer_id) VALUES (?, ?)",
		userID,
		customerID,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			_, updateErr := db.BunDB.ExecContext(ctx,
				"UPDATE billing_customers SET paddle_customer_id = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?",
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
	if strings.TrimSpace(update.PaddlePriceID) == "" {
		var existingPriceID string
		err := db.BunDB.QueryRowContext(ctx,
			"SELECT paddle_price_id FROM billing_subscriptions WHERE paddle_subscription_id = ? LIMIT 1",
			strings.TrimSpace(update.PaddleSubscriptionID),
		).Scan(&existingPriceID)
		if err == nil {
			update.PaddlePriceID = strings.TrimSpace(existingPriceID)
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	tier := TierFromPriceID(update.PaddlePriceID)
	graceUntil := update.GraceUntil
	if status == "past_due" && !graceUntil.Valid {
		graceUntil = sql.NullTime{Time: time.Now().UTC().Add(PastDueGracePeriod), Valid: true}
	}

	row := db.BillingSubscription{
		UserID:               strings.TrimSpace(update.UserID),
		PaddleSubscriptionID: strings.TrimSpace(update.PaddleSubscriptionID),
		PaddlePriceID:        strings.TrimSpace(update.PaddlePriceID),
		Tier:                 tier,
		Status:               status,
		CurrentPeriodEndsAt:  update.CurrentPeriodEndsAt,
		GraceUntil:           graceUntil,
		CanceledAt:           update.CanceledAt,
	}

	result, err := db.BunDB.ExecContext(ctx,
		"UPDATE billing_subscriptions SET user_id = ?, paddle_price_id = ?, tier = ?, status = ?, current_period_ends_at = ?, grace_until = ?, canceled_at = ?, updated_at = CURRENT_TIMESTAMP WHERE paddle_subscription_id = ?",
		row.UserID,
		row.PaddlePriceID,
		row.Tier,
		row.Status,
		row.CurrentPeriodEndsAt,
		row.GraceUntil,
		row.CanceledAt,
		row.PaddleSubscriptionID,
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
		"INSERT INTO billing_subscriptions(paddle_subscription_id, user_id, paddle_price_id, tier, status, current_period_ends_at, grace_until, canceled_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		row.PaddleSubscriptionID,
		row.UserID,
		row.PaddlePriceID,
		row.Tier,
		row.Status,
		row.CurrentPeriodEndsAt,
		row.GraceUntil,
		row.CanceledAt,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			_, updateErr := db.BunDB.ExecContext(ctx,
				"UPDATE billing_subscriptions SET user_id = ?, paddle_price_id = ?, tier = ?, status = ?, current_period_ends_at = ?, grace_until = ?, canceled_at = ?, updated_at = CURRENT_TIMESTAMP WHERE paddle_subscription_id = ?",
				row.UserID,
				row.PaddlePriceID,
				row.Tier,
				row.Status,
				row.CurrentPeriodEndsAt,
				row.GraceUntil,
				row.CanceledAt,
				row.PaddleSubscriptionID,
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

func ParseWebhookSubscription(rawBody []byte) (WebhookSubscriptionUpdate, bool, error) {
	var root map[string]any
	if err := json.Unmarshal(rawBody, &root); err != nil {
		return WebhookSubscriptionUpdate{}, false, err
	}

	eventType := firstString(root, "event_type", "type")
	eventID := firstString(root, "event_id", "id", "notification_id")
	if eventType == "" {
		return WebhookSubscriptionUpdate{}, false, errors.New("missing event type")
	}
	if eventID == "" {
		return WebhookSubscriptionUpdate{}, false, errors.New("missing event id")
	}

	if !strings.HasPrefix(strings.ToLower(eventType), "subscription.") {
		return WebhookSubscriptionUpdate{EventID: eventID, EventType: eventType}, false, nil
	}

	dataAny := root["data"]
	data, ok := dataAny.(map[string]any)
	if !ok {
		return WebhookSubscriptionUpdate{}, false, errors.New("missing data payload")
	}

	priceID := firstString(data,
		"items[0].price.id",
		"items[0].price_id",
		"price.id",
		"price_id",
	)

	update := WebhookSubscriptionUpdate{
		EventID:              eventID,
		EventType:            eventType,
		PaddleSubscriptionID: firstString(data, "id", "subscription_id"),
		PaddleCustomerID:     firstString(data, "customer_id", "customer.id"),
		PaddlePriceID:        priceID,
		Status:               strings.ToLower(firstString(data, "status")),
		UserID:               firstString(data, "custom_data.user_id", "metadata.user_id"),
		CustomerEmail:        strings.ToLower(firstString(data, "customer.email", "customer_email", "email")),
		CurrentPeriodEndsAt:  parseTime(firstString(data, "current_billing_period.ends_at", "next_billed_at", "scheduled_change.effective_at")),
		CanceledAt:           parseTime(firstString(data, "canceled_at")),
		GraceUntil:           parseTime(firstString(data, "custom_data.grace_until")),
	}

	if update.PaddleSubscriptionID == "" {
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

	userID, err := ResolveUserIDForWebhook(ctx, update.UserID, update.PaddleCustomerID, update.CustomerEmail)
	if err != nil {
		return false, err
	}
	if userID == "" {
		userID, err = resolveUserIDBySubscriptionID(ctx, update.PaddleSubscriptionID)
		if err != nil {
			return false, err
		}
	}
	if userID == "" {
		return false, ErrWebhookUserNotResolved
	}

	update.UserID = userID
	if err := UpsertCustomer(ctx, update.UserID, update.PaddleCustomerID); err != nil {
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
