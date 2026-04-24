package billing

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"bandcash/internal/db"
)

var ErrLemonAPIKeyMissing = errors.New("LEMON_API_KEY is required for webhook subscription sync")
var ErrSubscriptionItemIDMissing = errors.New("missing subscription item id")
var ErrCustomerPortalURLMissing = errors.New("missing customer portal url")
var ErrUpdatePaymentMethodURLMissing = errors.New("missing update payment method url")

var fetchCanonicalWebhookUpdate = fetchCanonicalWebhookUpdateFromAPI

var lemonAPIBaseURL = "https://api.lemonsqueezy.com/v1"
var lemonHTTPClient = &http.Client{Timeout: 15 * time.Second}

func lemonAPIRequest(ctx context.Context, method, path string, payload any) ([]byte, error) {
	apiKey := strings.TrimSpace(os.Getenv("LEMON_API_KEY"))
	if apiKey == "" {
		return nil, ErrLemonAPIKeyMissing
	}

	endpoint := strings.TrimRight(lemonAPIBaseURL, "/") + "/" + strings.TrimLeft(path, "/")
	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.api+json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/vnd.api+json")
	}

	resp, err := lemonHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lemon api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func fetchCanonicalWebhookUpdateFromAPI(ctx context.Context, webhookUpdate WebhookSubscriptionUpdate) (WebhookSubscriptionUpdate, error) {
	subscriptionID := strings.TrimSpace(webhookUpdate.SubscriptionID)
	if subscriptionID == "" {
		return WebhookSubscriptionUpdate{}, errors.New("missing subscription id")
	}

	body, err := lemonAPIRequest(ctx, http.MethodGet, "subscriptions/"+url.PathEscape(subscriptionID), nil)
	if err != nil {
		return WebhookSubscriptionUpdate{}, err
	}

	var responseRoot map[string]any
	if err := json.Unmarshal(body, &responseRoot); err != nil {
		return WebhookSubscriptionUpdate{}, err
	}
	data, ok := responseRoot["data"].(map[string]any)
	if !ok {
		return WebhookSubscriptionUpdate{}, errors.New("lemon api response missing data payload")
	}

	canonicalPayload := map[string]any{
		"meta": map[string]any{
			"event_name": webhookUpdate.EventType,
			"event_id":   webhookUpdate.EventID,
			"custom_data": map[string]any{
				"user_id": webhookUpdate.UserID,
			},
		},
		"data": data,
	}
	canonicalRaw, err := json.Marshal(canonicalPayload)
	if err != nil {
		return WebhookSubscriptionUpdate{}, err
	}

	canonicalUpdate, isSubscriptionEvent, err := ParseWebhookSubscription(canonicalRaw)
	if err != nil {
		return WebhookSubscriptionUpdate{}, err
	}
	if !isSubscriptionEvent {
		return WebhookSubscriptionUpdate{}, errors.New("lemon api response is not a subscription event")
	}

	if strings.TrimSpace(canonicalUpdate.CustomerID) == "" {
		canonicalUpdate.CustomerID = strings.TrimSpace(webhookUpdate.CustomerID)
	}
	if strings.TrimSpace(canonicalUpdate.CustomerEmail) == "" {
		canonicalUpdate.CustomerEmail = strings.TrimSpace(webhookUpdate.CustomerEmail)
	}
	if strings.TrimSpace(canonicalUpdate.UserID) == "" {
		canonicalUpdate.UserID = strings.TrimSpace(webhookUpdate.UserID)
	}
	if strings.TrimSpace(canonicalUpdate.SubscriptionItemID) == "" {
		canonicalUpdate.SubscriptionItemID = strings.TrimSpace(webhookUpdate.SubscriptionItemID)
	}
	if strings.TrimSpace(canonicalUpdate.PortalURL) == "" {
		canonicalUpdate.PortalURL = strings.TrimSpace(webhookUpdate.PortalURL)
	}
	if strings.TrimSpace(canonicalUpdate.UpdatePaymentURL) == "" {
		canonicalUpdate.UpdatePaymentURL = strings.TrimSpace(webhookUpdate.UpdatePaymentURL)
	}

	return canonicalUpdate, nil
}

func UpdateSubscriptionItemQuantity(ctx context.Context, subscriptionItemID string, quantity int) error {
	subscriptionItemID = strings.TrimSpace(subscriptionItemID)
	if subscriptionItemID == "" {
		return ErrSubscriptionItemIDMissing
	}
	if quantity < 1 {
		return errors.New("quantity must be positive")
	}

	payload := map[string]any{
		"data": map[string]any{
			"type": "subscription-items",
			"id":   subscriptionItemID,
			"attributes": map[string]any{
				"quantity":            quantity,
				"invoice_immediately": true,
			},
		},
	}

	_, err := lemonAPIRequest(ctx, http.MethodPatch, "subscription-items/"+url.PathEscape(subscriptionItemID), payload)
	return err
}

func SyncSubscriptionFromProvider(ctx context.Context, userID string) (db.BillingSubscription, bool, error) {
	sub, exists, err := GetUserSubscription(ctx, userID)
	if err != nil || !exists {
		return sub, exists, err
	}
	if strings.TrimSpace(sub.ProviderSubscriptionID) == "" {
		return sub, true, nil
	}

	update, err := fetchCanonicalWebhookUpdateFromAPI(ctx, WebhookSubscriptionUpdate{
		EventType:      "subscription_updated",
		SubscriptionID: sub.ProviderSubscriptionID,
		UserID:         userID,
	})
	if err != nil {
		return db.BillingSubscription{}, false, err
	}
	update.UserID = userID
	if err := UpsertCustomer(ctx, userID, update.CustomerID); err != nil {
		return db.BillingSubscription{}, false, err
	}
	if err := UpsertSubscription(ctx, update); err != nil {
		return db.BillingSubscription{}, false, err
	}

	return GetUserSubscription(ctx, userID)
}

func GetSignedCustomerPortalURL(ctx context.Context, userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", ErrInvalidUserID
	}

	var customerID string
	err := db.BunDB.QueryRowContext(ctx,
		"SELECT provider_customer_id FROM billing_customers WHERE user_id = ? LIMIT 1",
		userID,
	).Scan(&customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrCustomerPortalURLMissing
		}
		return "", err
	}
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return "", ErrCustomerPortalURLMissing
	}

	body, err := lemonAPIRequest(ctx, http.MethodGet, "customers/"+url.PathEscape(customerID), nil)
	if err != nil {
		return "", err
	}

	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return "", err
	}
	portalURL := firstString(root,
		"data.attributes.urls.customer_portal",
		"data.attributes.urls.customer_portal_url",
	)
	if strings.TrimSpace(portalURL) == "" {
		return "", ErrCustomerPortalURLMissing
	}
	return strings.TrimSpace(portalURL), nil
}

func GetSignedUpdatePaymentMethodURL(ctx context.Context, userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", ErrInvalidUserID
	}

	subscription, exists, err := GetUserSubscription(ctx, userID)
	if err != nil {
		return "", err
	}
	if !exists || strings.TrimSpace(subscription.ProviderSubscriptionID) == "" {
		return "", ErrUpdatePaymentMethodURLMissing
	}

	body, err := lemonAPIRequest(ctx, http.MethodGet, "subscriptions/"+url.PathEscape(strings.TrimSpace(subscription.ProviderSubscriptionID)), nil)
	if err != nil {
		return "", err
	}

	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return "", err
	}
	updatePaymentURL := firstString(root,
		"data.attributes.urls.update_payment_method",
		"data.attributes.urls.update_payment_method_url",
	)
	if strings.TrimSpace(updatePaymentURL) == "" {
		return "", ErrUpdatePaymentMethodURLMissing
	}
	return strings.TrimSpace(updatePaymentURL), nil
}
