package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUpdateSubscriptionItemQuantity_SendsInvoiceImmediatelyPatch(t *testing.T) {
	ctx := context.Background()
	if err := os.Setenv("EMAIL_PROVIDER", "resend"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("RESEND_API_KEY", "test_resend"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("LEMON_API_KEY", "test_api_key"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}

	var gotMethod string
	var gotPath string
	var gotAuth string
	var got map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"id":"si_123"}}`))
	}))
	t.Cleanup(server.Close)

	originalBaseURL := lemonAPIBaseURL
	originalClient := lemonHTTPClient
	lemonAPIBaseURL = server.URL
	lemonHTTPClient = server.Client()
	t.Cleanup(func() {
		lemonAPIBaseURL = originalBaseURL
		lemonHTTPClient = originalClient
	})

	err := UpdateSubscriptionItemQuantity(ctx, "si_123", 5)
	if err != nil {
		t.Fatalf("UpdateSubscriptionItemQuantity returned error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Fatalf("expected PATCH, got %s", gotMethod)
	}
	if gotPath != "/subscription-items/si_123" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotAuth != "Bearer test_api_key" {
		t.Fatalf("unexpected auth header: %s", gotAuth)
	}

	data, _ := got["data"].(map[string]any)
	if data == nil {
		t.Fatal("missing data object")
	}
	attrs, _ := data["attributes"].(map[string]any)
	if attrs == nil {
		t.Fatal("missing attributes object")
	}
	if attrs["quantity"] != float64(5) {
		t.Fatalf("expected quantity 5, got %#v", attrs["quantity"])
	}
	if attrs["invoice_immediately"] != true {
		t.Fatalf("expected invoice_immediately true, got %#v", attrs["invoice_immediately"])
	}
}

func TestUpdateSubscriptionItemQuantity_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	if err := os.Setenv("EMAIL_PROVIDER", "resend"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("RESEND_API_KEY", "test_resend"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("LEMON_API_KEY", ""); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}

	if err := UpdateSubscriptionItemQuantity(ctx, "", 2); err != ErrSubscriptionItemIDMissing {
		t.Fatalf("expected ErrSubscriptionItemIDMissing, got %v", err)
	}
	if err := UpdateSubscriptionItemQuantity(ctx, "si_123", 0); err == nil || !strings.Contains(err.Error(), "quantity") {
		t.Fatalf("expected quantity validation error, got %v", err)
	}
	if err := UpdateSubscriptionItemQuantity(ctx, "si_123", 2); err != ErrLemonAPIKeyMissing {
		t.Fatalf("expected ErrLemonAPIKeyMissing, got %v", err)
	}
}

func TestFetchCanonicalWebhookUpdateFromAPI_FallsBackToWebhookItemID(t *testing.T) {
	ctx := context.Background()
	if err := os.Setenv("EMAIL_PROVIDER", "resend"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("RESEND_API_KEY", "test_resend"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	if err := os.Setenv("LEMON_API_KEY", "test_api_key"); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		_, _ = w.Write([]byte(`{"data":{"id":"sub_123","attributes":{"status":"active","customer_id":"ctm_123","customer_email":"webhook.user@example.com","variant_id":"pri_test_pro","quantity":2}}}`))
	}))
	t.Cleanup(server.Close)

	originalBaseURL := lemonAPIBaseURL
	originalClient := lemonHTTPClient
	lemonAPIBaseURL = server.URL
	lemonHTTPClient = server.Client()
	t.Cleanup(func() {
		lemonAPIBaseURL = originalBaseURL
		lemonHTTPClient = originalClient
	})

	update, err := fetchCanonicalWebhookUpdateFromAPI(ctx, WebhookSubscriptionUpdate{
		EventID:            "evt_123",
		EventType:          "subscription_updated",
		SubscriptionID:     "sub_123",
		SubscriptionItemID: "si_from_webhook",
	})
	if err != nil {
		t.Fatalf("fetchCanonicalWebhookUpdateFromAPI returned error: %v", err)
	}
	if update.SubscriptionItemID != "si_from_webhook" {
		t.Fatalf("expected webhook fallback subscription item id, got %s", update.SubscriptionItemID)
	}
}

func TestGetSignedUpdatePaymentMethodURL_MissingUserID(t *testing.T) {
	ctx := context.Background()
	got, err := GetSignedUpdatePaymentMethodURL(ctx, "")
	if err != ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty url, got %s", got)
	}
}
