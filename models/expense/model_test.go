package expense

import (
	"database/sql"
	"testing"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func TestExpenseTableSpec(t *testing.T) {
	e := New()
	spec := e.TableQuerySpec()

	if spec.DefaultSort != "date" || spec.DefaultDir != "desc" {
		t.Fatalf("unexpected expense spec defaults: %+v", spec)
	}
	for _, key := range []string{"date", "title", "amount", "paid", "paid_at"} {
		if _, ok := spec.AllowedSorts[key]; !ok {
			t.Fatalf("expected allowed sort %q", key)
		}
	}
}

func TestMatchesExpenseFilters(t *testing.T) {
	expense := db.Expense{Title: "Gas", Description: "Tour fuel", Date: "2026-03-10"}

	if !matchesExpenseFilters(expense, utils.TableQuery{}) {
		t.Fatal("expected empty query to match")
	}
	if matchesExpenseFilters(expense, utils.TableQuery{Search: "hotel"}) {
		t.Fatal("expected missing search to fail")
	}
	if !matchesExpenseFilters(expense, utils.TableQuery{Search: "fuel"}) {
		t.Fatal("expected description search to match")
	}
	if matchesExpenseFilters(expense, utils.TableQuery{Year: "2025"}) {
		t.Fatal("expected wrong year to fail")
	}
	if matchesExpenseFilters(expense, utils.TableQuery{From: "2026-03-11"}) {
		t.Fatal("expected from-date after expense to fail")
	}
	if matchesExpenseFilters(expense, utils.TableQuery{To: "2026-03-09"}) {
		t.Fatal("expected to-date before expense to fail")
	}
}

func TestSortExpensesByPaidAt(t *testing.T) {
	expenses := []db.Expense{
		{ID: "exp_1", Date: "2026-01-01", PaidAt: sql.NullString{String: "2026-02-02", Valid: true}},
		{ID: "exp_2", Date: "2026-01-02", PaidAt: sql.NullString{}},
		{ID: "exp_3", Date: "2026-01-03", PaidAt: sql.NullString{String: "2026-01-01", Valid: true}},
	}

	sortExpenses(expenses, "paid_at", "asc")
	if expenses[0].ID != "exp_3" || expenses[1].ID != "exp_1" || expenses[2].ID != "exp_2" {
		t.Fatalf("unexpected asc paid_at order: %s, %s, %s", expenses[0].ID, expenses[1].ID, expenses[2].ID)
	}

	sortExpenses(expenses, "paid_at", "desc")
	if expenses[0].ID != "exp_2" || expenses[1].ID != "exp_1" || expenses[2].ID != "exp_3" {
		t.Fatalf("unexpected desc paid_at order: %s, %s, %s", expenses[0].ID, expenses[1].ID, expenses[2].ID)
	}
}
