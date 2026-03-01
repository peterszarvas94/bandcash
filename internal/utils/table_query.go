package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type TableQuerySpec struct {
	DefaultSort      string
	DefaultDir       string
	AllowedSorts     map[string]struct{}
	AllowedPageSizes map[int]struct{}
	DefaultSize      int
	MaxSearchLen     int
}

var StandardTablePageSizes = []int{10, 50, 100, 200}

const (
	DefaultTablePageSize  = 50
	DefaultTableMaxSearch = 100
)

type Queryable interface {
	TableQuerySpec() TableQuerySpec
}

type TableQuery struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
	Sort     string `json:"sort"`
	SortSet  bool   `json:"sortSet"`
	Dir      string `json:"dir"`
}

type TableQueryParseResult struct {
	Query    TableQuery
	Rejected map[string]string
}

type TablePagination struct {
	Page       int
	PageSize   int
	TotalItems int
	TotalPages int
	HasPrev    bool
	HasNext    bool
}

func ParseTableQuery(c echo.Context, queryable Queryable) TableQuery {
	return ParseTableQueryWithResult(c, queryable).Query
}

func ParseTableQueryWithResult(c echo.Context, queryable Queryable) TableQueryParseResult {
	spec := queryable.TableQuerySpec()
	rejected := make(map[string]string)

	query := TableQuery{
		Page:     1,
		PageSize: intDefault(spec.DefaultSize, 20),
		Sort:     spec.DefaultSort,
		Dir:      defaultDirection(spec.DefaultDir),
	}

	if rawPage := c.QueryParam("page"); rawPage != "" {
		value, err := strconv.Atoi(rawPage)
		if err != nil || value <= 0 {
			rejected["page"] = "must be a positive integer"
		} else {
			query.Page = value
		}
	}

	if rawPageSize := c.QueryParam("pageSize"); rawPageSize != "" {
		value, err := strconv.Atoi(rawPageSize)
		if err != nil || value <= 0 {
			rejected["pageSize"] = "must be a positive integer"
		} else {
			query.PageSize = value
		}
	}

	if len(spec.AllowedPageSizes) > 0 {
		if _, ok := spec.AllowedPageSizes[query.PageSize]; !ok {
			rejected["pageSize"] = "value not allowlisted"
			query.PageSize = intDefault(spec.DefaultSize, 20)
		}
	}

	search := strings.TrimSpace(c.QueryParam("q"))
	maxSearchLen := intDefault(spec.MaxSearchLen, 100)
	if len(search) > maxSearchLen {
		rejected["q"] = "trimmed to max length"
		search = search[:maxSearchLen]
	}
	query.Search = search

	rawSort := c.QueryParam("sort")
	if rawSort != "" {
		if _, ok := spec.AllowedSorts[rawSort]; ok {
			query.Sort = rawSort
			query.SortSet = true
		} else {
			rejected["sort"] = "value not allowlisted"
		}
	}

	dir := c.QueryParam("dir")
	if dir != "" {
		if !query.SortSet {
			rejected["dir"] = "requires a valid sort"
		} else if dir == "asc" || dir == "desc" {
			query.Dir = dir
		} else {
			rejected["dir"] = "must be asc or desc"
		}
	}

	if !query.SortSet {
		query.Dir = defaultDirection(spec.DefaultDir)
	}

	if len(rejected) == 0 {
		rejected = nil
	}

	return TableQueryParseResult{Query: query, Rejected: rejected}
}

func (q TableQuery) Offset() int64 {
	return int64((q.Page - 1) * q.PageSize)
}

func BuildTablePagination(totalItems int64, query TableQuery) TablePagination {
	total := int(totalItems)
	totalPages := 0
	if total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}

	if totalPages == 0 {
		totalPages = 1
	}

	return TablePagination{
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
		HasPrev:    query.Page > 1,
		HasNext:    query.Page < totalPages,
	}
}

func ClampPage(query TableQuery, totalItems int64) TableQuery {
	totalPages := 0
	if totalItems > 0 {
		totalPages = int((totalItems + int64(query.PageSize) - 1) / int64(query.PageSize))
	}

	if totalPages == 0 {
		totalPages = 1
	}

	if query.Page > totalPages {
		query.Page = totalPages
	}

	if query.Page < 1 {
		query.Page = 1
	}

	return query
}

func intDefault(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func defaultDirection(dir string) string {
	if dir == "desc" {
		return "desc"
	}
	return "asc"
}

func NormalizeTableQuery(query TableQuery, spec TableQuerySpec) TableQuery {
	normalized := TableQuery{
		Page:     1,
		PageSize: intDefault(spec.DefaultSize, 20),
		Sort:     spec.DefaultSort,
		Dir:      defaultDirection(spec.DefaultDir),
	}

	if query.Page > 0 {
		normalized.Page = query.Page
	}

	if query.PageSize > 0 {
		normalized.PageSize = query.PageSize
	}

	if len(spec.AllowedPageSizes) > 0 {
		if _, ok := spec.AllowedPageSizes[normalized.PageSize]; !ok {
			normalized.PageSize = intDefault(spec.DefaultSize, 20)
		}
	}

	search := strings.TrimSpace(query.Search)
	maxSearchLen := intDefault(spec.MaxSearchLen, 100)
	if len(search) > maxSearchLen {
		search = search[:maxSearchLen]
	}
	normalized.Search = search

	if query.Sort != "" {
		if _, ok := spec.AllowedSorts[query.Sort]; ok {
			normalized.Sort = query.Sort
			normalized.SortSet = true
		}
	}

	if normalized.SortSet {
		if query.Dir == "asc" || query.Dir == "desc" {
			normalized.Dir = query.Dir
		}
	}

	return normalized
}

type TableQueryPatch struct {
	Search   *string
	Sort     *string
	Dir      *string
	Page     *int
	PageSize *int
}

type SortCycle struct {
	Sort string
	Dir  string
}

func NextSortCycle(query TableQuery, column string) SortCycle {
	if !query.SortSet || query.Sort != column {
		return SortCycle{Sort: column, Dir: "asc"}
	}

	if query.Dir == "asc" {
		return SortCycle{Sort: column, Dir: "desc"}
	}

	return SortCycle{}
}

func BuildTableQueryURL(basePath string, query TableQuery) string {
	return BuildTableQueryURLWith(basePath, query, TableQueryPatch{})
}

func BuildTableSortURL(basePath string, query TableQuery, column string) string {
	next := NextSortCycle(query, column)
	page := 1
	return BuildTableQueryURLWith(basePath, query, TableQueryPatch{
		Sort: &next.Sort,
		Dir:  &next.Dir,
		Page: &page,
	})
}

func BuildTablePageURL(basePath string, query TableQuery, page, totalPages int) string {
	if page < 1 {
		page = 1
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	return BuildTableQueryURLWith(basePath, query, TableQueryPatch{Page: &page})
}

func BuildTablePageSizeURL(basePath string, query TableQuery, pageSize int) string {
	page := 1
	return BuildTableQueryURLWith(basePath, query, TableQueryPatch{
		Page:     &page,
		PageSize: &pageSize,
	})
}

func TableQuerySignals(query TableQuery) map[string]any {
	sort := ""
	dir := ""
	if query.SortSet {
		sort = query.Sort
		dir = query.Dir
	}

	return map[string]any{
		"search":   query.Search,
		"sort":     sort,
		"sortSet":  query.SortSet,
		"dir":      dir,
		"page":     query.Page,
		"pageSize": query.PageSize,
	}
}

func StandardTableQuerySpec(defaultSort, defaultDir string, allowedSorts ...string) TableQuerySpec {
	sorts := make(map[string]struct{}, len(allowedSorts))
	for _, sort := range allowedSorts {
		sorts[sort] = struct{}{}
	}

	pageSizes := make(map[int]struct{}, len(StandardTablePageSizes))
	for _, size := range StandardTablePageSizes {
		pageSizes[size] = struct{}{}
	}

	return TableQuerySpec{
		DefaultSort:      defaultSort,
		DefaultDir:       defaultDir,
		AllowedSorts:     sorts,
		AllowedPageSizes: pageSizes,
		DefaultSize:      DefaultTablePageSize,
		MaxSearchLen:     DefaultTableMaxSearch,
	}
}

func BuildTableSearchDatastarAction(basePath string, defaultPageSize int) string {
	if defaultPageSize <= 0 {
		defaultPageSize = DefaultTablePageSize
	}
	return fmt.Sprintf("const url = globalThis.tableSearchAction('%s', $tableQuery, %d); @get(url)", basePath, defaultPageSize)
}

func BuildTableQueryDatastarAction(url string) string {
	return fmt.Sprintf("history.pushState(null, '', '%s'); @get('%s')", url, url)
}

func PageSizeButtonClass(current, value int) string {
	if current == value {
		return "btn btn-sm btn-active"
	}
	return "btn btn-sm"
}

func BuildTableQueryURLWith(basePath string, query TableQuery, patch TableQueryPatch) string {
	resolved := query

	if patch.Search != nil {
		resolved.Search = strings.TrimSpace(*patch.Search)
	}

	if patch.Sort != nil {
		resolved.Sort = *patch.Sort
		resolved.SortSet = resolved.Sort != ""
	}

	if patch.Dir != nil {
		resolved.Dir = *patch.Dir
	}

	if !resolved.SortSet {
		resolved.Dir = ""
	}

	if patch.Page != nil {
		resolved.Page = *patch.Page
	}

	if patch.PageSize != nil {
		resolved.PageSize = *patch.PageSize
	}

	if resolved.Page < 1 {
		resolved.Page = 1
	}

	if resolved.PageSize < 1 {
		resolved.PageSize = DefaultTablePageSize
	}

	// Parse the basePath to handle existing query parameters properly
	u, err := url.Parse(basePath)
	if err != nil {
		// If parsing fails, fall back to simple concatenation
		values := url.Values{}
		if resolved.Search != "" {
			values.Set("q", resolved.Search)
		}
		if resolved.SortSet {
			values.Set("sort", resolved.Sort)
			values.Set("dir", resolved.Dir)
		}
		if resolved.Page > 1 {
			values.Set("page", strconv.Itoa(resolved.Page))
		}
		if resolved.PageSize != DefaultTablePageSize {
			values.Set("pageSize", strconv.Itoa(resolved.PageSize))
		}
		encoded := values.Encode()
		if encoded == "" {
			return basePath
		}
		if strings.Contains(basePath, "?") {
			return basePath + "&" + encoded
		}
		return basePath + "?" + encoded
	}

	// Merge existing query parameters with new ones
	values := u.Query()
	if resolved.Search != "" {
		values.Set("q", resolved.Search)
	} else {
		values.Del("q")
	}
	if resolved.SortSet {
		values.Set("sort", resolved.Sort)
		values.Set("dir", resolved.Dir)
	} else {
		values.Del("sort")
		values.Del("dir")
	}
	if resolved.Page > 1 {
		values.Set("page", strconv.Itoa(resolved.Page))
	} else {
		values.Del("page")
	}
	if resolved.PageSize != DefaultTablePageSize {
		values.Set("pageSize", strconv.Itoa(resolved.PageSize))
	} else {
		values.Del("pageSize")
	}

	u.RawQuery = values.Encode()
	return u.String()
}
