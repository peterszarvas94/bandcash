package billing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"bandcash/internal/db"
	authstore "bandcash/models/auth/data"
)

const (
	testUserID    = "usr_12345678901234567890"
	testUserEmail = "webhook.user@example.com"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("APP_ENV", "development")
	_ = os.Setenv("PADDLE_PRICE_ID", "pri_test_pro")
	_ = os.Setenv("PADDLE_API_BASE_URL", "https://sandbox-api.paddle.com")
	os.Exit(m.Run())
}

func setupTestDB(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "billing_test.sqlite")
	if err := db.Init(dbPath); err != nil {
		t.Fatalf("db.Init failed: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Migrate(); err != nil {
		t.Fatalf("db.Migrate failed: %v", err)
	}
}

func readFixture(t *testing.T) []byte {
	t.Helper()
	content, err := os.ReadFile("test_webhook_event.json")
	if err != nil {
		t.Fatalf("read fixture failed: %v", err)
	}
	return content
}

func subscriptionPayload(eventID, subscriptionID, customerID, userID, email, status, priceID string) []byte {
	return []byte(fmt.Sprintf(`{"event_id":"%s","event_type":"subscription.created","data":{"id":"%s","status":"%s","customer_id":"%s","customer":{"email":"%s"},"items":[{"price":{"id":"%s"},"quantity":1}],"current_billing_period":{"ends_at":"2026-05-14T18:47:00.844006Z"},"custom_data":{"user_id":"%s"}}}`,
		eventID,
		subscriptionID,
		status,
		customerID,
		email,
		priceID,
		userID,
	))
}

func TestParseWebhookSubscription_FromFixture(t *testing.T) {
	raw := readFixture(t)

	update, isSubscriptionEvent, err := ParseWebhookSubscription(raw)
	if err != nil {
		t.Fatalf("ParseWebhookSubscription returned error: %v", err)
	}
	if !isSubscriptionEvent {
		t.Fatal("expected subscription event")
	}
	if update.EventType != "subscription.created" {
		t.Fatalf("unexpected event type: %s", update.EventType)
	}
	if update.UserID != testUserID {
		t.Fatalf("unexpected user id: %s", update.UserID)
	}
	if update.PaddlePriceID != "pri_test_pro" {
		t.Fatalf("unexpected price id: %s", update.PaddlePriceID)
	}
	if update.PaddleCustomerID == "" {
		t.Fatal("expected paddle customer id to be parsed")
	}
	if !update.CurrentPeriodEndsAt.Valid {
		t.Fatal("expected current period end to be parsed")
	}
}

func TestProcessWebhook_SubscriptionCreated_PersistsBillingRows(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	if _, err := authstore.CreateUser(ctx, authstore.CreateUserParams{
		ID:            testUserID,
		Email:         testUserEmail,
		PreferredLang: "en",
	}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	raw := readFixture(t)
	parsed, isSub, parseErr := ParseWebhookSubscription(raw)
	if parseErr != nil {
		t.Fatalf("ParseWebhookSubscription failed: %v", parseErr)
	}
	if !isSub {
		t.Fatal("expected subscription event in fixture")
	}
	resolvedUserID, resolveErr := ResolveUserIDForWebhook(ctx, parsed.UserID, parsed.PaddleCustomerID, parsed.CustomerEmail)
	if resolveErr != nil {
		t.Fatalf("ResolveUserIDForWebhook failed: %v", resolveErr)
	}
	if resolvedUserID != testUserID {
		t.Fatalf("expected resolved user id %s, got %s", testUserID, resolvedUserID)
	}

	processed, err := ProcessWebhook(ctx, raw)
	if err != nil {
		t.Fatalf("ProcessWebhook returned error: %v", err)
	}
	if !processed {
		t.Fatal("expected webhook to be processed")
	}

	customersCount, err := db.BunDB.NewSelect().TableExpr("billing_customers").Count(ctx)
	if err != nil {
		t.Fatalf("count billing_customers failed: %v", err)
	}
	subscriptionsCount, err := db.BunDB.NewSelect().TableExpr("billing_subscriptions").Count(ctx)
	if err != nil {
		t.Fatalf("count billing_subscriptions failed: %v", err)
	}
	processedCount, err := db.BunDB.NewSelect().TableExpr("billing_webhook_events").Count(ctx)
	if err != nil {
		t.Fatalf("count billing_webhook_events failed: %v", err)
	}

	if subscriptionsCount != 1 {
		t.Fatalf("expected 1 billing subscription row, got %d (customers=%d processed=%d)", subscriptionsCount, customersCount, processedCount)
	}
	if customersCount != 1 {
		t.Fatalf("expected 1 billing customer row, got %d (subscriptions=%d processed=%d)", customersCount, subscriptionsCount, processedCount)
	}

	var sub db.BillingSubscription
	if err := db.BunDB.NewSelect().Model(&sub).Where("user_id = ?", testUserID).Scan(ctx); err != nil {
		t.Fatalf("select billing_subscriptions failed: %v", err)
	}
	if sub.Tier != TierPro {
		t.Fatalf("expected tier %s, got %s", TierPro, sub.Tier)
	}
	if sub.Status != "active" {
		t.Fatalf("expected status active, got %s", sub.Status)
	}

	state, err := CurrentAccessState(ctx, testUserID)
	if err != nil {
		t.Fatalf("CurrentAccessState failed: %v", err)
	}
	if state.SubscriptionCount != 1 {
		t.Fatalf("expected subscription count 1, got %d", state.SubscriptionCount)
	}
}

func TestProcessWebhook_UnresolvedUser_ReturnsErrorAndDoesNotMarkProcessed(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	raw := readFixture(t)
	processed, err := ProcessWebhook(ctx, raw)
	if err == nil {
		t.Fatal("expected error for unresolved user")
	}
	if err != ErrWebhookUserNotResolved {
		t.Fatalf("unexpected error: %v", err)
	}
	if processed {
		t.Fatal("expected webhook not processed")
	}

	count, countErr := db.BunDB.NewSelect().TableExpr("billing_webhook_events").Count(ctx)
	if countErr != nil {
		t.Fatalf("count billing_webhook_events failed: %v", countErr)
	}
	if count != 0 {
		t.Fatalf("expected 0 processed webhook rows, got %d", count)
	}
}

func TestUpsertSubscription_DirectPersist(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	if _, err := authstore.CreateUser(ctx, authstore.CreateUserParams{
		ID:            testUserID,
		Email:         testUserEmail,
		PreferredLang: "en",
	}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := UpsertCustomer(ctx, testUserID, "ctm_direct_test"); err != nil {
		t.Fatalf("UpsertCustomer failed: %v", err)
	}
	if err := UpsertSubscription(ctx, WebhookSubscriptionUpdate{
		UserID:               testUserID,
		PaddleSubscriptionID: "sub_direct_test",
		PaddlePriceID:        "pri_test_pro",
		Status:               "active",
	}); err != nil {
		t.Fatalf("UpsertSubscription failed: %v", err)
	}

	customersCount, err := db.BunDB.NewSelect().TableExpr("billing_customers").Count(ctx)
	if err != nil {
		t.Fatalf("count billing_customers failed: %v", err)
	}
	subscriptionsCount, err := db.BunDB.NewSelect().TableExpr("billing_subscriptions").Count(ctx)
	if err != nil {
		t.Fatalf("count billing_subscriptions failed: %v", err)
	}
	if customersCount != 1 || subscriptionsCount != 1 {
		t.Fatalf("expected billing rows to persist (customers=%d subscriptions=%d)", customersCount, subscriptionsCount)
	}
}

func TestProcessWebhook_MultipleActiveSubscriptions_AccumulatesSlots(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	if _, err := authstore.CreateUser(ctx, authstore.CreateUserParams{ID: testUserID, Email: testUserEmail, PreferredLang: "en"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	first := subscriptionPayload("evt_aaaaaaaaaaaaaaaaaaaa", "sub_slot_one", "ctm_slot", testUserID, testUserEmail, "active", "pri_test_pro")
	second := subscriptionPayload("evt_bbbbbbbbbbbbbbbbbbbb", "sub_slot_two", "ctm_slot", testUserID, testUserEmail, "active", "pri_test_pro")

	if processed, err := ProcessWebhook(ctx, first); err != nil || !processed {
		t.Fatalf("first webhook failed processed=%v err=%v", processed, err)
	}
	if processed, err := ProcessWebhook(ctx, second); err != nil || !processed {
		t.Fatalf("second webhook failed processed=%v err=%v", processed, err)
	}

	state, err := CurrentAccessState(ctx, testUserID)
	if err != nil {
		t.Fatalf("CurrentAccessState failed: %v", err)
	}
	if state.SubscriptionCount != 2 {
		t.Fatalf("expected 2 subscription slots, got %d", state.SubscriptionCount)
	}
	if state.RemainingSlots != 2 {
		t.Fatalf("expected 2 remaining slots with zero groups, got %d", state.RemainingSlots)
	}
}
