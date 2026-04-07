package group

import (
	"time"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type NewGroupPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type EditGroupPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Group           db.Group
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type GroupWithRole struct {
	Group db.Group
	Role  string
}

type GroupsPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
	AllGroups       []GroupWithRole
	AdminGroups     []GroupSummary
	ReaderGroups    []GroupSummary
	Query           utils.TableQuery
	Pagination      utils.TablePagination
	GroupsTable     utils.TableLayout
}

type GroupSummary struct {
	Group       db.Group
	ViewerCount int
	AdminEmail  string
}

type UsersPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
	CurrentUserID   string
	Group           db.Group
	UserRows        []GroupUserRow
	IsAdmin         bool
	Query           utils.TableQuery
	Pager           utils.TablePagination
	GroupID         string
	UsersTable      utils.TableLayout
}

type UsersNewPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Group           db.Group
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type UserPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
	CurrentUserID   string
	GroupID         string
	Group           db.Group
	UserRow         GroupUserRow
	IsAdmin         bool
}

type UserEditPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Group           db.Group
	UserRow         GroupUserRow
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type UserInvitePageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Group           db.Group
	UserRow         GroupUserRow
	IsAdmin         bool
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type GroupUserRow struct {
	Kind      string
	Status    string
	Role      string
	Email     string
	UserID    string
	InviteID  string
	CreatedAt time.Time
}

type GroupPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
	Group           db.Group
	Admin           db.User
	Income          int64
	IncomePaid      int64
	IncomeUnpaid    int64
	Payouts         int64
	PayoutsPaid     int64
	PayoutsUnpaid   int64
	Expenses        int64
	ExpensesPaid    int64
	ExpensesUnpaid  int64
	Leftover        int64
	IsAdmin         bool
}

type GroupPaymentsPageData struct {
	Title               string
	Breadcrumbs         []utils.Crumb
	Signals             map[string]any
	IsAuthenticated     bool
	IsSuperAdmin        bool
	IsAdmin             bool
	GroupID             string
	Group               db.Group
	UnpaidEvents        []GroupPaymentEventRow
	UnpaidParticipants  []GroupPaymentParticipantRow
	UnpaidExpenses      []GroupPaymentExpenseRow
	EventsTable         utils.TableLayout
	ParticipantsTable   utils.TableLayout
	ExpensesTable       utils.TableLayout
}

type GroupPaymentEventRow struct {
	ID     string
	Title  string
	Amount int64
	PaidAt string
}

type GroupPaymentParticipantRow struct {
	MemberID      string
	MemberName    string
	EventID       string
	EventTitle    string
	PayoutAmount  int64
	PaidAt        string
}

type GroupPaymentExpenseRow struct {
	ID     string
	Title  string
	Amount int64
	PaidAt string
}
