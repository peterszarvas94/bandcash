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

func stubCanonicalSyncFromWebhookPayload(t *testing.T) {
	t.Helper()
	original := fetchCanonicalWebhookUpdate
	fetchCanonicalWebhookUpdate = func(_ context.Context, update WebhookSubscriptionUpdate) (WebhookSubscriptionUpdate, error) {
		return update, nil
	}
	t.Cleanup(func() {
		fetchCanonicalWebhookUpdate = original
	})
}

func readFixture(t *testing.T) []byte {
	t.Helper()
	content, err := os.ReadFile("test_webhook_event.json")
	if err != nil {
		t.Fatalf("read fixture failed: %v", err)
	}
	return content
}

func subscriptionPayload(eventID, eventName, subscriptionID, customerID, userID, email, status, priceID string, quantity int) []byte {
	return []byte(fmt.Sprintf(`{"meta":{"event_name":"%s","event_id":"%s","custom_data":{"user_id":"%s"}},"data":{"id":"%s","attributes":{"status":"%s","customer_id":"%s","customer_email":"%s","variant_id":"%s","first_subscription_item":{"id":4567,"quantity":%d},"urls":{"customer_portal":"https://subscriptions.example.com/portal","update_payment_method":"https://subscriptions.example.com/update-payment"},"renews_at":"2026-05-14T18:47:00.844006Z"}}}`,
		eventName,
		eventID,
		userID,
		subscriptionID,
		status,
		customerID,
		email,
		priceID,
		quantity,
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
	if update.EventType != "subscription_created" {
		t.Fatalf("unexpected event type: %s", update.EventType)
	}
	if update.UserID != testUserID {
		t.Fatalf("unexpected user id: %s", update.UserID)
	}
	if update.VariantID != "pri_test_pro" {
		t.Fatalf("unexpected price id: %s", update.VariantID)
	}
	if update.CustomerID == "" {
		t.Fatal("expected customer id to be parsed")
	}
	if !update.CurrentPeriodEndsAt.Valid {
		t.Fatal("expected current period end to be parsed")
	}
	if update.SubscriptionItemID != "1234" {
		t.Fatalf("unexpected subscription item id: %s", update.SubscriptionItemID)
	}
	if update.SeatQuantity != 2 {
		t.Fatalf("expected seat quantity 2, got %d", update.SeatQuantity)
	}
}

func TestParseWebhookSubscription_SubscriptionItemFromRelationships(t *testing.T) {
	raw := []byte(`{"meta":{"event_name":"subscription_updated","event_id":"evt_rel"},"data":{"id":"sub_rel","attributes":{"status":"active","customer_id":"ctm_rel","customer_email":"webhook.user@example.com","variant_id":"pri_test_pro","quantity":3},"relationships":{"first_subscription_item":{"data":{"id":"si_rel_123"}}}}}`)

	update, isSubscriptionEvent, err := ParseWebhookSubscription(raw)
	if err != nil {
		t.Fatalf("ParseWebhookSubscription returned error: %v", err)
	}
	if !isSubscriptionEvent {
		t.Fatal("expected subscription event")
	}
	if update.SubscriptionItemID != "si_rel_123" {
		t.Fatalf("expected subscription item id from relationships, got %s", update.SubscriptionItemID)
	}
}

func TestParseWebhookSubscription_InvoicePayloadUsesAttributesSubscriptionID(t *testing.T) {
	raw := []byte(`{"meta":{"event_name":"subscription_payment_success","webhook_id":"wh_123"},"data":{"id":"6838947","type":"subscription-invoices","attributes":{"status":"paid","subscription_id":2083758,"customer_id":8393687,"user_email":"peterszarvas94@gmail.com"}}}`)

	update, isSubscriptionEvent, err := ParseWebhookSubscription(raw)
	if err != nil {
		t.Fatalf("ParseWebhookSubscription returned error: %v", err)
	}
	if !isSubscriptionEvent {
		t.Fatal("expected subscription event")
	}
	if update.SubscriptionID != "2083758" {
		t.Fatalf("expected subscription id 2083758, got %s", update.SubscriptionID)
	}
	if update.CustomerID != "8393687" {
		t.Fatalf("expected customer id 8393687, got %s", update.CustomerID)
	}
}

func TestProcessWebhook_SubscriptionCreated_PersistsBillingRows(t *testing.T) {
	setupTestDB(t)
	stubCanonicalSyncFromWebhookPayload(t)
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
	resolvedUserID, resolveErr := ResolveUserIDForWebhook(ctx, parsed.UserID, parsed.CustomerID, parsed.CustomerEmail)
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
	if sub.ProviderSubscriptionItemID == "" {
		t.Fatal("expected provider subscription item id")
	}
	if sub.SeatQuantity != 2 {
		t.Fatalf("expected seat quantity 2, got %d", sub.SeatQuantity)
	}

	state, err := CurrentAccessState(ctx, testUserID)
	if err != nil {
		t.Fatalf("CurrentAccessState failed: %v", err)
	}
	if state.SubscriptionCount != 2 {
		t.Fatalf("expected subscription count 2, got %d", state.SubscriptionCount)
	}
}

func TestProcessWebhook_UnresolvedUser_ReturnsErrorAndDoesNotMarkProcessed(t *testing.T) {
	setupTestDB(t)
	stubCanonicalSyncFromWebhookPayload(t)
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
		UserID:             testUserID,
		SubscriptionID:     "sub_direct_test",
		SubscriptionItemID: "4567",
		VariantID:          "pri_test_pro",
		SeatQuantity:       3,
		Status:             "active",
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

func TestProcessWebhook_SingleSubscription_UpdatesSeatQuantity(t *testing.T) {
	setupTestDB(t)
	stubCanonicalSyncFromWebhookPayload(t)
	ctx := context.Background()

	if _, err := authstore.CreateUser(ctx, authstore.CreateUserParams{ID: testUserID, Email: testUserEmail, PreferredLang: "en"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	first := subscriptionPayload("evt_aaaaaaaaaaaaaaaaaaaa", "subscription_created", "sub_slot_single", "ctm_slot", testUserID, testUserEmail, "active", "pri_test_pro", 1)
	second := subscriptionPayload("evt_bbbbbbbbbbbbbbbbbbbb", "subscription_updated", "sub_slot_single", "ctm_slot", testUserID, testUserEmail, "active", "pri_test_pro", 4)

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
	if state.SubscriptionCount != 4 {
		t.Fatalf("expected 4 subscription slots, got %d", state.SubscriptionCount)
	}
	if state.RemainingSlots != 4 {
		t.Fatalf("expected 4 remaining slots with zero groups, got %d", state.RemainingSlots)
	}

	rows, err := db.BunDB.NewSelect().TableExpr("billing_subscriptions").Count(ctx)
	if err != nil {
		t.Fatalf("count billing_subscriptions failed: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected single subscription row, got %d", rows)
	}
}

func TestProcessWebhook_UsesCanonicalSubscriptionState(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	if _, err := authstore.CreateUser(ctx, authstore.CreateUserParams{ID: testUserID, Email: testUserEmail, PreferredLang: "en"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	original := fetchCanonicalWebhookUpdate
	fetchCanonicalWebhookUpdate = func(_ context.Context, update WebhookSubscriptionUpdate) (WebhookSubscriptionUpdate, error) {
		update.VariantID = "pri_test_pro"
		update.SeatQuantity = 5
		update.Status = "active"
		return update, nil
	}
	t.Cleanup(func() {
		fetchCanonicalWebhookUpdate = original
	})

	raw := subscriptionPayload("evt_canonical_aaaaaaaaaa", "subscription_created", "sub_canonical", "ctm_canonical", testUserID, testUserEmail, "active", "pri_test_pro", 1)
	processed, err := ProcessWebhook(ctx, raw)
	if err != nil {
		t.Fatalf("ProcessWebhook returned error: %v", err)
	}
	if !processed {
		t.Fatal("expected webhook to be processed")
	}

	state, err := CurrentAccessState(ctx, testUserID)
	if err != nil {
		t.Fatalf("CurrentAccessState failed: %v", err)
	}
	if state.SubscriptionCount != 5 {
		t.Fatalf("expected canonical subscription count 5, got %d", state.SubscriptionCount)
	}
}
