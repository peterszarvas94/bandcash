package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func setupQueryTestDB(t *testing.T) (*Queries, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "query-test.db")
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	goose.SetBaseFS(migrationsFS)
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	cleanup := func() {
		_ = sqlDB.Close()
	}

	return New(sqlDB), cleanup
}

func seedBaseGroup(t *testing.T, q *Queries) string {
	t.Helper()

	ctx := context.Background()
	if _, err := q.CreateUser(ctx, CreateUserParams{ID: "usr_admin", Email: "admin@example.com"}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	group, err := q.CreateGroup(ctx, CreateGroupParams{ID: "grp_main", Name: "Main Group", AdminUserID: "usr_admin"})
	if err != nil {
		t.Fatalf("create group: %v", err)
	}

	if _, err := q.CreateGroup(ctx, CreateGroupParams{ID: "grp_other", Name: "Other Group", AdminUserID: "usr_admin"}); err != nil {
		t.Fatalf("create other group: %v", err)
	}

	return group.ID
}

func TestEventsFilteredQueries(t *testing.T) {
	q, cleanup := setupQueryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	groupID := seedBaseGroup(t, q)

	rows := []CreateEventParams{
		{ID: "evt_alpha", GroupID: groupID, Title: "Alpha", Time: "2026-01-01T10:00", Description: "", Amount: 100},
		{ID: "evt_beta", GroupID: groupID, Title: "Beta Party", Time: "2026-01-02T10:00", Description: "Has notes", Amount: 700},
		{ID: "evt_gamma", GroupID: groupID, Title: "Gamma", Time: "2026-01-03T10:00", Description: "contains test", Amount: 300},
		{ID: "evt_other", GroupID: "grp_other", Title: "Other test", Time: "2026-01-04T10:00", Description: "other", Amount: 999},
	}
	for _, row := range rows {
		if _, err := q.CreateEvent(ctx, row); err != nil {
			t.Fatalf("create event %s: %v", row.ID, err)
		}
	}

	count, err := q.CountEventsFiltered(ctx, CountEventsFilteredParams{GroupID: groupID, Search: "test"})
	if err != nil {
		t.Fatalf("count events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	list, err := q.ListEventsByAmountDescFiltered(ctx, ListEventsByAmountDescFilteredParams{GroupID: groupID, Search: "", Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("list events by amount desc: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 events, got %d", len(list))
	}
	if list[0].ID != "evt_beta" || list[1].ID != "evt_gamma" {
		t.Fatalf("unexpected order: %s, %s", list[0].ID, list[1].ID)
	}
}

func TestMembersFilteredQueries(t *testing.T) {
	q, cleanup := setupQueryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	groupID := seedBaseGroup(t, q)

	rows := []CreateMemberParams{
		{ID: "mem_alice", GroupID: groupID, Name: "Alice", Description: ""},
		{ID: "mem_bob", GroupID: groupID, Name: "Bob", Description: "drummer"},
		{ID: "mem_carol", GroupID: groupID, Name: "Carol", Description: "test singer"},
		{ID: "mem_other", GroupID: "grp_other", Name: "Other", Description: "test"},
	}
	for _, row := range rows {
		if _, err := q.CreateMember(ctx, row); err != nil {
			t.Fatalf("create member %s: %v", row.ID, err)
		}
	}

	count, err := q.CountMembersFiltered(ctx, CountMembersFilteredParams{GroupID: groupID, Search: ""})
	if err != nil {
		t.Fatalf("count members: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}

	list, err := q.ListMembersByNameAscFiltered(ctx, ListMembersByNameAscFilteredParams{GroupID: groupID, Search: "", Limit: 2, Offset: 1})
	if err != nil {
		t.Fatalf("list members by name asc: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 members, got %d", len(list))
	}
	if list[0].Name != "Bob" || list[1].Name != "Carol" {
		t.Fatalf("unexpected order: %s, %s", list[0].Name, list[1].Name)
	}
}

func TestExpensesFilteredQueries(t *testing.T) {
	q, cleanup := setupQueryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	groupID := seedBaseGroup(t, q)

	rows := []CreateExpenseParams{
		{ID: "exp_rent", GroupID: groupID, Title: "Rent", Description: "", Amount: 1000, Date: "2026-01-01"},
		{ID: "exp_food", GroupID: groupID, Title: "Food", Description: "team lunch", Amount: 200, Date: "2026-01-03"},
		{ID: "exp_travel", GroupID: groupID, Title: "Travel", Description: "test fuel", Amount: 500, Date: "2026-01-02"},
		{ID: "exp_other", GroupID: "grp_other", Title: "Other", Description: "test", Amount: 999, Date: "2026-01-04"},
	}
	for _, row := range rows {
		if _, err := q.CreateExpense(ctx, row); err != nil {
			t.Fatalf("create expense %s: %v", row.ID, err)
		}
	}

	count, err := q.CountExpensesFiltered(ctx, CountExpensesFilteredParams{GroupID: groupID, Search: "test"})
	if err != nil {
		t.Fatalf("count expenses: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	list, err := q.ListExpensesByDateDescFiltered(ctx, ListExpensesByDateDescFilteredParams{GroupID: groupID, Search: "", Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("list expenses by date desc: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(list))
	}
	if list[0].ID != "exp_food" || list[1].ID != "exp_travel" {
		t.Fatalf("unexpected order: %s, %s", list[0].ID, list[1].ID)
	}
}
