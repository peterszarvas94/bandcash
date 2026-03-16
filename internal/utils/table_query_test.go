package utils

import (
	"net/http/httptest"
	"strings"
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
	t.Run("parses allowlisted query params", func(t *testing.T) {
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
	})
}

func TestParseTableQuery_InvalidValuesFallback(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/?q=abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz&sort=hack&dir=up&page=-1&pageSize=9999", nil)
	ctx := e.NewContext(req, httptest.NewRecorder())

	t.Run("falls back to defaults for invalid values", func(t *testing.T) {
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
	})

	t.Run("reports rejected fields", func(t *testing.T) {
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
	})
}

func TestClampPageAndPagination(t *testing.T) {
	query := TableQuery{Page: 99, PageSize: 10}

	t.Run("clamps page to total pages", func(t *testing.T) {
		query = ClampPage(query, 35)
		if query.Page != 4 {
			t.Fatalf("expected clamped page 4, got %d", query.Page)
		}
	})

	t.Run("builds pager metadata from clamped query", func(t *testing.T) {
		pager := BuildTablePagination(35, query)
		if pager.TotalPages != 4 || pager.HasNext || !pager.HasPrev {
			t.Fatalf("unexpected pager: %+v", pager)
		}
	})
}

func TestParseTableQuery_RejectsMaliciousOrIllegalInputs(t *testing.T) {
	t.Run("ignores malicious query values and falls back safely", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest("GET", "/?q=%20%20hello%20%20&sort=createdAt;DROP%20TABLE%20users&dir=desc&page=0&pageSize=2147483647", nil)
		ctx := e.NewContext(req, httptest.NewRecorder())

		query := ParseTableQuery(ctx, testQueryable{spec: StandardTableQuerySpec(StandardTableQuerySpecParams{
			DefaultSort:  "createdAt",
			DefaultDir:   "desc",
			AllowedSorts: []string{"name", "createdAt"},
		})})

		if query.Search != "hello" {
			t.Fatalf("expected trimmed search hello, got %q", query.Search)
		}
		if query.Sort != "createdAt" || query.Dir != "desc" {
			t.Fatalf("expected default sort/dir createdAt/desc, got %s/%s", query.Sort, query.Dir)
		}
		if query.SortSet {
			t.Fatal("expected SortSet=false for disallowed sort value")
		}
		if query.Page != 1 {
			t.Fatalf("expected fallback page 1, got %d", query.Page)
		}
		if query.PageSize != DefaultTablePageSize {
			t.Fatalf("expected fallback pageSize %d, got %d", DefaultTablePageSize, query.PageSize)
		}
	})
}

func TestParseTableQuery_DateFilters(t *testing.T) {
	t.Run("range overrides year when both provided", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest("GET", "/?year=2026&from=2026-01-01&to=2026-01-31", nil)
		ctx := e.NewContext(req, httptest.NewRecorder())

		query := ParseTableQuery(ctx, testQueryable{spec: StandardTableQuerySpec(StandardTableQuerySpecParams{
			DefaultSort:  "time",
			DefaultDir:   "asc",
			AllowedSorts: []string{"time"},
		})})

		if query.Year != "" {
			t.Fatalf("expected year to be cleared when range is set, got %q", query.Year)
		}
		if query.From != "2026-01-01" || query.To != "2026-01-31" {
			t.Fatalf("expected range preserved, got from=%q to=%q", query.From, query.To)
		}
	})

	t.Run("rejects invalid date inputs", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest("GET", "/?from=2026-44-01&to=bad", nil)
		ctx := e.NewContext(req, httptest.NewRecorder())

		result := ParseTableQueryWithResult(ctx, testQueryable{spec: StandardTableQuerySpec(StandardTableQuerySpecParams{
			DefaultSort:  "time",
			DefaultDir:   "asc",
			AllowedSorts: []string{"time"},
		})})
		if _, ok := result.Rejected["from"]; !ok {
			t.Fatalf("expected from to be rejected, got %+v", result.Rejected)
		}
		if _, ok := result.Rejected["to"]; !ok {
			t.Fatalf("expected to to be rejected, got %+v", result.Rejected)
		}
	})
}

func TestParseTableQueryWithResult_ReportsRejectedFields(t *testing.T) {
	t.Run("returns explicit rejected field map", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest("GET", "/?sort=not_allowed&dir=sideways&page=-3&pageSize=999", nil)
		ctx := e.NewContext(req, httptest.NewRecorder())

		result := ParseTableQueryWithResult(ctx, testQueryable{spec: StandardTableQuerySpec(StandardTableQuerySpecParams{
			DefaultSort:  "createdAt",
			DefaultDir:   "desc",
			AllowedSorts: []string{"name", "createdAt"},
		})})

		if result.Rejected == nil {
			t.Fatal("expected rejected fields map")
		}
		for _, key := range []string{"sort", "dir", "page", "pageSize"} {
			if _, ok := result.Rejected[key]; !ok {
				t.Fatalf("expected %s to be rejected, got %+v", key, result.Rejected)
			}
		}
	})
}

func TestNormalizeTableQuery_RejectsDisallowedPageSizeAndSort(t *testing.T) {
	t.Run("normalizes disallowed values to safe defaults", func(t *testing.T) {
		spec := StandardTableQuerySpec(StandardTableQuerySpecParams{
			DefaultSort:  "createdAt",
			DefaultDir:   "desc",
			AllowedSorts: []string{"name", "createdAt"},
		})

		query := NormalizeTableQuery(TableQuery{
			Page:     -5,
			PageSize: 1337,
			Search:   strings.Repeat("a", 200),
			Sort:     "email;delete",
			Dir:      "sideways",
		}, spec)

		if query.Page != 1 {
			t.Fatalf("expected normalized page 1, got %d", query.Page)
		}
		if query.PageSize != DefaultTablePageSize {
			t.Fatalf("expected normalized pageSize %d, got %d", DefaultTablePageSize, query.PageSize)
		}
		if query.Sort != "createdAt" || query.SortSet {
			t.Fatalf("expected default sort and SortSet=false, got %+v", query)
		}
		if query.Dir != "desc" {
			t.Fatalf("expected default dir desc, got %s", query.Dir)
		}
		if len(query.Search) != DefaultTableMaxSearch {
			t.Fatalf("expected max search len %d, got %d", DefaultTableMaxSearch, len(query.Search))
		}
	})
}

func TestBuildTableSortURL_CycleAndResetPage(t *testing.T) {
	t.Run("toggles asc to desc and resets page", func(t *testing.T) {
		query := TableQuery{Page: 4, PageSize: 50, Search: "party", Sort: "name", SortSet: true, Dir: "asc"}
		url := BuildTableSortURL("/dashboard", query, "name")
		if !strings.Contains(url, "sort=name") || !strings.Contains(url, "dir=desc") {
			t.Fatalf("expected desc sort for second click, got %s", url)
		}
		if strings.Contains(url, "page=4") {
			t.Fatalf("expected page reset in sort url, got %s", url)
		}
	})

	t.Run("clears sort on third click", func(t *testing.T) {
		third := BuildTableSortURL("/dashboard", TableQuery{Sort: "name", SortSet: true, Dir: "desc"}, "name")
		if strings.Contains(third, "sort=") || strings.Contains(third, "dir=") {
			t.Fatalf("expected sort cleared on third click, got %s", third)
		}
	})
}

func TestBuildTablePageURL_ClampsToRange(t *testing.T) {
	query := TableQuery{Page: 2, PageSize: 50, Search: "abc", Sort: "createdAt", SortSet: true, Dir: "desc"}

	t.Run("clamps negative page to minimum", func(t *testing.T) {
		min := BuildTablePageURL("/admin?tab=users", query, -2, 8)
		if strings.Contains(min, "page=-") || strings.Contains(min, "page=0") {
			t.Fatalf("expected page clamped to minimum, got %s", min)
		}
	})

	t.Run("clamps page above total pages to last page", func(t *testing.T) {
		max := BuildTablePageURL("/admin?tab=users", query, 99, 3)
		if !strings.Contains(max, "page=3") {
			t.Fatalf("expected page clamped to 3, got %s", max)
		}
	})
}

func TestBuildTablePageSizeURL_ResetsPageAndRejectsIllegalPageSize(t *testing.T) {
	spec := StandardTableQuerySpec(StandardTableQuerySpecParams{
		DefaultSort:  "createdAt",
		DefaultDir:   "desc",
		AllowedSorts: []string{"name", "createdAt"},
	})
	query := TableQuery{Page: 6, PageSize: 50, Search: "abc", Sort: "createdAt", SortSet: true, Dir: "desc"}

	t.Run("resets page to first when page-size changes", func(t *testing.T) {
		url := BuildTablePageSizeURL("/dashboard", query, 9999)
		if strings.Contains(url, "page=6") {
			t.Fatalf("expected page reset in page-size url, got %s", url)
		}
	})

	t.Run("fallbacks disallowed page-size when parsed", func(t *testing.T) {
		url := BuildTablePageSizeURL("/dashboard", query, 9999)
		e := echo.New()
		req := httptest.NewRequest("GET", url, nil)
		ctx := e.NewContext(req, httptest.NewRecorder())
		parsed := ParseTableQuery(ctx, testQueryable{spec: spec})
		if parsed.PageSize != DefaultTablePageSize {
			t.Fatalf("expected disallowed pageSize fallback to %d, got %d", DefaultTablePageSize, parsed.PageSize)
		}
	})
}

func TestBuildTableQueryURLWith_MergesExistingBaseQuery(t *testing.T) {
	t.Run("preserves existing base query params while patching", func(t *testing.T) {
		query := TableQuery{Page: 2, PageSize: 50, Search: "abc", Sort: "createdAt", SortSet: true, Dir: "desc"}
		url := BuildTableQueryURLWith("/admin?tab=users&foo=bar", query, TableQueryPatch{Search: func() *string { s := ""; return &s }()})
		if !strings.Contains(url, "tab=users") || !strings.Contains(url, "foo=bar") {
			t.Fatalf("expected existing base query params preserved, got %s", url)
		}
		if strings.Contains(url, "q=") {
			t.Fatalf("expected empty search to remove q, got %s", url)
		}
	})
}

func TestBuildTableQueryURLWith_DateFilters(t *testing.T) {
	t.Run("clears year when complete range is provided", func(t *testing.T) {
		query := TableQuery{Page: 1, PageSize: 10, Year: "2026"}
		from := "2026-01-01"
		to := "2026-12-31"
		url := BuildTableQueryURLWith("/groups/g1/events", query, TableQueryPatch{From: &from, To: &to})

		if strings.Contains(url, "year=") {
			t.Fatalf("expected year to be removed when range set, got %s", url)
		}
		if !strings.Contains(url, "from=2026-01-01") || !strings.Contains(url, "to=2026-12-31") {
			t.Fatalf("expected range params in url, got %s", url)
		}
	})

	t.Run("year selection clears existing range", func(t *testing.T) {
		query := TableQuery{Page: 1, PageSize: 10, From: "2026-01-01", To: "2026-12-31"}
		year := "2025"
		url := BuildTableQueryURLWith("/groups/g1/events", query, TableQueryPatch{Year: &year})

		if !strings.Contains(url, "year=2025") {
			t.Fatalf("expected year in url, got %s", url)
		}
		if strings.Contains(url, "from=") || strings.Contains(url, "to=") {
			t.Fatalf("expected range params to be removed, got %s", url)
		}
	})
}

func TestTableQuerySignals_HidesSortWhenNotExplicitlySet(t *testing.T) {
	t.Run("omits sort and dir when sort is implicit default", func(t *testing.T) {
		signals := TableQuerySignals(TableQuery{Page: 1, PageSize: 50, Search: "abc", Sort: "createdAt", SortSet: false, Dir: "desc"})

		if got, _ := signals["sort"].(string); got != "" {
			t.Fatalf("expected hidden sort, got %q", got)
		}
		if got, _ := signals["dir"].(string); got != "" {
			t.Fatalf("expected hidden dir, got %q", got)
		}
	})
}
