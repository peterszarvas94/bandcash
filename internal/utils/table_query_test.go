package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type testQueryable struct {
	spec TableQuerySpec
}

func (t testQueryable) TableQuerySpec() TableQuerySpec {
	return t.spec
}

func TestParseTableQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/?q=party&sort=amount&dir=desc&page=2&pageSize=10", nil)
	ctx := e.NewContext(req, httptest.NewRecorder())

	query := ParseTableQuery(ctx, testQueryable{spec: TableQuerySpec{
		DefaultSort: "time",
		DefaultDir:  "asc",
		AllowedSorts: map[string]struct{}{
			"time":   {},
			"title":  {},
			"amount": {},
		},
		AllowedPageSizes: map[int]struct{}{
			10: {},
			50: {},
		},
		DefaultSize:  50,
		MaxSearchLen: 100,
	}})

	if query.Page != 2 || query.PageSize != 10 {
		t.Fatalf("expected page/pageSize 2/10, got %d/%d", query.Page, query.PageSize)
	}
	if query.Search != "party" {
		t.Fatalf("expected search 'party', got %q", query.Search)
	}
	if query.Sort != "amount" || query.Dir != "desc" || !query.SortSet {
		t.Fatalf("unexpected parsed query: %+v", query)
	}
	if query.Offset() != 10 {
		t.Fatalf("expected offset 10, got %d", query.Offset())
	}
}

func TestParseTableQuery_InvalidValuesFallback(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/?q=abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz&sort=hack&dir=up&page=-1&pageSize=9999", nil)
	ctx := e.NewContext(req, httptest.NewRecorder())

	query := ParseTableQuery(ctx, testQueryable{spec: TableQuerySpec{
		DefaultSort: "time",
		DefaultDir:  "asc",
		AllowedSorts: map[string]struct{}{
			"time": {},
		},
		AllowedPageSizes: map[int]struct{}{
			10: {},
			50: {},
		},
		DefaultSize:  50,
		MaxSearchLen: 20,
	}})

	if query.Page != 1 {
		t.Fatalf("expected fallback page 1, got %d", query.Page)
	}
	if query.PageSize != 50 {
		t.Fatalf("expected fallback pageSize 50, got %d", query.PageSize)
	}
	if query.Sort != "time" || query.Dir != "asc" || query.SortSet {
		t.Fatalf("expected defaults for invalid values, got %+v", query)
	}
	if len(query.Search) != 20 {
		t.Fatalf("expected trimmed search length 20, got %d", len(query.Search))
	}

	result := ParseTableQueryWithResult(ctx, testQueryable{spec: TableQuerySpec{
		DefaultSort: "time",
		DefaultDir:  "asc",
		AllowedSorts: map[string]struct{}{
			"time": {},
		},
		AllowedPageSizes: map[int]struct{}{
			10: {},
			50: {},
		},
		DefaultSize:  50,
		MaxSearchLen: 20,
	}})
	if len(result.Rejected) == 0 {
		t.Fatal("expected rejected fields for invalid query")
	}
	if _, ok := result.Rejected["sort"]; !ok {
		t.Fatal("expected sort to be rejected")
	}
	if _, ok := result.Rejected["pageSize"]; !ok {
		t.Fatal("expected pageSize to be rejected")
	}
}

func TestClampPageAndPagination(t *testing.T) {
	query := TableQuery{Page: 99, PageSize: 10}
	query = ClampPage(query, 35)
	if query.Page != 4 {
		t.Fatalf("expected clamped page 4, got %d", query.Page)
	}

	pager := BuildTablePagination(35, query)
	if pager.TotalPages != 4 || pager.HasNext || !pager.HasPrev {
		t.Fatalf("unexpected pager: %+v", pager)
	}
}
