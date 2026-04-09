package admin

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	adminstore "bandcash/models/admin/data"
)

func TestParseIntParamAndAdminQueries(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/admin/users?page=3&pageSize=20&q=%20alice%20&sort=email&dir=desc", nil)
	c := e.NewContext(req, httptest.NewRecorder())

	if got := parseIntParam(c, "page", 1); got != 3 {
		t.Fatalf("expected page 3, got %d", got)
	}
	if got := parseIntParam(c, "missing", 50); got != 50 {
		t.Fatalf("expected missing param to use default 50, got %d", got)
	}

	users := parseAdminUsersQuery(c)
	if users.Page != 3 || users.PageSize != 20 {
		t.Fatalf("unexpected users pagination: %+v", users)
	}
	if users.Search != "alice" || users.Sort != "email" || users.Dir != "desc" || !users.SortSet {
		t.Fatalf("unexpected users query parse: %+v", users)
	}

	groups := parseAdminGroupsQuery(c)
	if groups.Page != 3 || groups.PageSize != 20 || groups.Search != "alice" {
		t.Fatalf("unexpected groups query parse: %+v", groups)
	}
}

func TestAdminTabLabelFallback(t *testing.T) {
	flags := adminTabLabel(t.Context(), "flags")
	if flags == "" {
		t.Fatal("expected flags label")
	}

	unknown := adminTabLabel(t.Context(), "unknown")
	if unknown != flags {
		t.Fatalf("expected unknown tab to fallback to flags label, got %q want %q", unknown, flags)
	}
}

func TestMapUserRowsSetsBanFlag(t *testing.T) {
	rows := []adminstore.AdminUserTableRow{
		{ID: "usr_1", Email: "a@example.com", IsBanned: 1},
		{ID: "usr_2", Email: "b@example.com", IsBanned: 0},
	}
	mapped := mapUsers(rows)
	if len(mapped) != 2 || !mapped[0].IsBanned {
		t.Fatalf("expected banned user in desc mapping, got %+v", mapped)
	}
	if mapped[1].IsBanned {
		t.Fatalf("expected second user to be unbanned, got %+v", mapped)
	}
}
