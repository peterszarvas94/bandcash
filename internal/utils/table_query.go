package utils

import (
	"net/url"
	"strconv"
	"strings"
	"time"

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

var StandardTablePageSizes = []int{100, 200, 500}

const (
	DefaultTablePageSize  = 100
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
	Summary  string `json:"summary"`
	DateMode string `json:"dateMode"`
	Year     string `json:"year"`
	From     string `json:"from"`
	To       string `json:"to"`
}

const (
	SummaryModeAll    = "all"
	SummaryModePaid   = "paid"
	SummaryModeUnpaid = "unpaid"
)

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
		Summary:  SummaryModeAll,
	}

	summary := strings.TrimSpace(c.QueryParam("summary"))
	if summary != "" {
		normalizedSummary := NormalizeSummaryMode(summary)
		if normalizedSummary != summary {
			rejected["summary"] = "must be all, paid, or unpaid"
		} else {
			query.Summary = normalizedSummary
		}
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

	year := strings.TrimSpace(c.QueryParam("year"))
	if year != "" {
		if isValidYear(year) {
			query.Year = year
		} else {
			rejected["year"] = "must be YYYY"
		}
	}

	dateMode := strings.TrimSpace(c.QueryParam("dateMode"))
	if dateMode == "custom" {
		query.DateMode = "custom"
	} else if dateMode != "" {
		rejected["dateMode"] = "must be custom"
	}

	from := strings.TrimSpace(c.QueryParam("from"))
	if from != "" {
		if isValidDateISO(from) {
			query.From = from
		} else {
			rejected["from"] = "must be YYYY-MM-DD"
		}
	}

	to := strings.TrimSpace(c.QueryParam("to"))
	if to != "" {
		if isValidDateISO(to) {
			query.To = to
		} else {
			rejected["to"] = "must be YYYY-MM-DD"
		}
	}

	query = normalizeDateFilterPriority(query)

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
		Summary:  SummaryModeAll,
		DateMode: strings.TrimSpace(query.DateMode),
		Year:     strings.TrimSpace(query.Year),
		From:     strings.TrimSpace(query.From),
		To:       strings.TrimSpace(query.To),
	}

	normalized.Summary = NormalizeSummaryMode(query.Summary)

	if normalized.DateMode != "custom" {
		normalized.DateMode = ""
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

	if !isValidYear(normalized.Year) {
		normalized.Year = ""
	}

	if !isValidDateISO(normalized.From) {
		normalized.From = ""
	}

	if !isValidDateISO(normalized.To) {
		normalized.To = ""
	}

	normalized = normalizeDateFilterPriority(normalized)

	return normalized
}

type TableQueryPatch struct {
	Search   *string
	Sort     *string
	Dir      *string
	Page     *int
	PageSize *int
	Summary  *string
	DateMode *string
	Year     *string
	From     *string
	To       *string
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

func BuildTableSummaryURL(basePath string, query TableQuery, summary string) string {
	normalizedSummary := NormalizeSummaryMode(summary)
	return BuildTableQueryURLWith(basePath, query, TableQueryPatch{Summary: &normalizedSummary})
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
		"summary":  query.Summary,
		"dateMode": query.DateMode,
		"year":     query.Year,
		"from":     query.From,
		"to":       query.To,
	}
}

type StandardTableQuerySpecParams struct {
	DefaultSort  string
	DefaultDir   string
	AllowedSorts []string
}

func StandardTableQuerySpec(params StandardTableQuerySpecParams) TableQuerySpec {
	sorts := make(map[string]struct{}, len(params.AllowedSorts))
	for _, sort := range params.AllowedSorts {
		sorts[sort] = struct{}{}
	}

	pageSizes := make(map[int]struct{}, len(StandardTablePageSizes))
	for _, size := range StandardTablePageSizes {
		pageSizes[size] = struct{}{}
	}

	return TableQuerySpec{
		DefaultSort:      params.DefaultSort,
		DefaultDir:       params.DefaultDir,
		AllowedSorts:     sorts,
		AllowedPageSizes: pageSizes,
		DefaultSize:      DefaultTablePageSize,
		MaxSearchLen:     DefaultTableMaxSearch,
	}
}

func BuildTableDateYearURL(basePath string, query TableQuery, year string) string {
	page := 1
	from := ""
	to := ""
	dateMode := ""
	trimmedYear := strings.TrimSpace(year)
	return BuildTableQueryURLWith(basePath, query, TableQueryPatch{
		Page:     &page,
		DateMode: &dateMode,
		Year:     &trimmedYear,
		From:     &from,
		To:       &to,
	})
}

func BuildTableDateClearURL(basePath string, query TableQuery) string {
	page := 1
	empty := ""
	return BuildTableQueryURLWith(basePath, query, TableQueryPatch{
		Page:     &page,
		DateMode: &empty,
		Year:     &empty,
		From:     &empty,
		To:       &empty,
	})
}

func BuildTableDateCustomURL(basePath string, query TableQuery) string {
	page := 1
	empty := ""
	dateMode := "custom"
	return BuildTableQueryURLWith(basePath, query, TableQueryPatch{
		Page:     &page,
		DateMode: &dateMode,
		Year:     &empty,
		From:     &empty,
		To:       &empty,
	})
}

func DateFilterAllButtonClass(query TableQuery) string {
	if DateFilterAllActive(query) {
		return "btn btn-xs btn-active"
	}
	return "btn btn-xs"
}

func DateFilterYearButtonClass(query TableQuery, year string) string {
	if DateFilterYearActive(query, year) {
		return "btn btn-xs btn-active"
	}
	return "btn btn-xs"
}

func DateFilterCustomButtonClass(query TableQuery) string {
	if DateFilterCustomActive(query) {
		return "btn btn-xs btn-active"
	}
	return "btn btn-xs"
}

func DateFilterAllActive(query TableQuery) bool {
	return query.DateMode != "custom" && query.Year == "" && !(query.From != "" && query.To != "")
}

func DateFilterYearActive(query TableQuery, year string) bool {
	return query.Year == year && !(query.From != "" && query.To != "")
}

func DateFilterCustomActive(query TableQuery) bool {
	return query.DateMode == "custom" || query.From != "" || query.To != ""
}

func PageSizeButtonClass(current, value int) string {
	if current == value {
		return "btn btn-xs btn-active"
	}
	return "btn btn-xs"
}

func BuildTableQueryURLWith(basePath string, query TableQuery, patch TableQueryPatch) string {
	resolved := query
	yearPatched := false

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

	if patch.Summary != nil {
		resolved.Summary = strings.TrimSpace(*patch.Summary)
	}

	if patch.DateMode != nil {
		resolved.DateMode = strings.TrimSpace(*patch.DateMode)
	}

	if patch.Year != nil {
		resolved.Year = strings.TrimSpace(*patch.Year)
		yearPatched = true
	}

	if patch.From != nil {
		resolved.From = strings.TrimSpace(*patch.From)
	}

	if patch.To != nil {
		resolved.To = strings.TrimSpace(*patch.To)
	}

	if resolved.Page < 1 {
		resolved.Page = 1
	}

	if resolved.PageSize < 1 {
		resolved.PageSize = DefaultTablePageSize
	}

	resolved.Summary = NormalizeSummaryMode(resolved.Summary)

	if !isValidYear(resolved.Year) {
		resolved.Year = ""
	}

	if !isValidDateISO(resolved.From) {
		resolved.From = ""
	}

	if !isValidDateISO(resolved.To) {
		resolved.To = ""
	}

	if resolved.DateMode != "custom" {
		resolved.DateMode = ""
	}

	if yearPatched && resolved.Year != "" {
		resolved.From = ""
		resolved.To = ""
	}

	resolved = normalizeDateFilterPriority(resolved)

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
		if resolved.Summary != SummaryModeAll {
			values.Set("summary", resolved.Summary)
		}
		if resolved.Year != "" {
			values.Set("year", resolved.Year)
		}
		if resolved.DateMode != "" {
			values.Set("dateMode", resolved.DateMode)
		}
		if resolved.From != "" {
			values.Set("from", resolved.From)
		}
		if resolved.To != "" {
			values.Set("to", resolved.To)
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

	if resolved.Summary != SummaryModeAll {
		values.Set("summary", resolved.Summary)
	} else {
		values.Del("summary")
	}

	if resolved.Year != "" {
		values.Set("year", resolved.Year)
	} else {
		values.Del("year")
	}

	if resolved.DateMode != "" {
		values.Set("dateMode", resolved.DateMode)
	} else {
		values.Del("dateMode")
	}

	if resolved.From != "" {
		values.Set("from", resolved.From)
	} else {
		values.Del("from")
	}

	if resolved.To != "" {
		values.Set("to", resolved.To)
	} else {
		values.Del("to")
	}

	u.RawQuery = values.Encode()
	return u.String()
}

func isValidYear(year string) bool {
	if year == "" || len(year) != 4 {
		return false
	}
	value, err := strconv.Atoi(year)
	if err != nil {
		return false
	}
	return value >= 1900 && value <= 3000
}

func isValidDateISO(value string) bool {
	if value == "" {
		return false
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func normalizeDateFilterPriority(query TableQuery) TableQuery {
	if query.From != "" && query.To != "" {
		query.Year = ""
		return query
	}

	if query.Year != "" {
		query.DateMode = ""
		query.From = ""
		query.To = ""
	}

	return query
}

func NormalizeSummaryMode(value string) string {
	switch strings.TrimSpace(value) {
	case SummaryModePaid:
		return SummaryModePaid
	case SummaryModeUnpaid:
		return SummaryModeUnpaid
	default:
		return SummaryModeAll
	}
}
