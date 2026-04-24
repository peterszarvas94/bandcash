package event

import (
	"context"
	"log/slog"
	"sort"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
	authstore "bandcash/models/auth/data"
	eventstore "bandcash/models/event/data"
	expensestore "bandcash/models/expense/data"
	groupstore "bandcash/models/group/data"
	memberstore "bandcash/models/member/data"
)

func TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "date",
		DefaultDir:   "desc",
		AllowedSorts: []string{"date", "time", "title", "place", "amount", "description", "paid", "paid_at"},
	})
}

func ParticipantTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "name",
		DefaultDir:   "asc",
		AllowedSorts: []string{"name", "amount", "expense", "total", "paid", "paid_at"},
	})
}

func GetShowData(ctx context.Context, groupID, eventID string, query utils.TableQuery) (EventData, error) {
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return EventData{}, err
	}

	event, err := eventstore.GetEvent(ctx, eventstore.GetEventParams{
		ID:      eventID,
		GroupID: groupID,
	})
	if err != nil {
		return EventData{}, err
	}

	allParticipants, err := eventstore.ListParticipantsByEvent(ctx, eventstore.ListParticipantsByEventParams{
		EventID: eventID,
		GroupID: groupID,
	})
	if err != nil {
		return EventData{}, err
	}

	members, err := memberstore.ListMembers(ctx, groupID)
	if err != nil {
		return EventData{}, err
	}

	participantMemberIDs := make(map[string]bool, len(allParticipants))
	participantByMemberID := make(map[string]eventstore.ListParticipantsByEventRow, len(allParticipants))
	for _, participant := range allParticipants {
		participantMemberIDs[participant.ID] = true
		participantByMemberID[participant.ID] = participant
	}

	filteredMembers := make([]db.Member, 0, len(members))
	wizardRows := make([]ParticipantWizardRow, 0, len(allParticipants))
	for _, member := range members {
		participant, included := participantByMemberID[member.ID]
		if included {
			wizardRows = append(wizardRows, ParticipantWizardRow{
				RowID:      member.ID,
				MemberID:   member.ID,
				MemberName: member.Name,
				Included:   true,
				Amount:     participant.ParticipantAmount,
				Expense:    participant.ParticipantExpense,
				Note:       strings.TrimSpace(participant.ParticipantNote),
				Paid:       participant.ParticipantPaid == 1,
				PaidAt: func() string {
					if !participant.ParticipantPaidAt.Valid {
						return ""
					}
					return utils.FormatDateInput(participant.ParticipantPaidAt.String)
				}(),
			})
		}

		if participantMemberIDs[member.ID] {
			continue
		}
		filteredMembers = append(filteredMembers, member)
	}

	participants := allParticipants

	sort.SliceStable(participants, func(i, j int) bool {
		left := participants[i]
		right := participants[j]

		less := false
		equal := false

		switch query.Sort {
		case "total":
			leftTotal := left.ParticipantAmount + left.ParticipantExpense
			rightTotal := right.ParticipantAmount + right.ParticipantExpense
			if leftTotal == rightTotal {
				leftName := strings.ToLower(left.Name)
				rightName := strings.ToLower(right.Name)
				less = leftName < rightName
				equal = leftName == rightName
			} else {
				less = leftTotal < rightTotal
			}
		case "amount":
			if left.ParticipantAmount == right.ParticipantAmount {
				leftName := strings.ToLower(left.Name)
				rightName := strings.ToLower(right.Name)
				less = leftName < rightName
				equal = leftName == rightName
			} else {
				less = left.ParticipantAmount < right.ParticipantAmount
			}
		case "expense":
			if left.ParticipantExpense == right.ParticipantExpense {
				leftName := strings.ToLower(left.Name)
				rightName := strings.ToLower(right.Name)
				less = leftName < rightName
				equal = leftName == rightName
			} else {
				less = left.ParticipantExpense < right.ParticipantExpense
			}
		case "paid":
			if left.ParticipantPaid == right.ParticipantPaid {
				leftName := strings.ToLower(left.Name)
				rightName := strings.ToLower(right.Name)
				less = leftName < rightName
				equal = leftName == rightName
			} else {
				less = left.ParticipantPaid < right.ParticipantPaid
			}
		case "paid_at":
			if left.ParticipantPaidAt.Valid && right.ParticipantPaidAt.Valid {
				if left.ParticipantPaidAt.String == right.ParticipantPaidAt.String {
					leftName := strings.ToLower(left.Name)
					rightName := strings.ToLower(right.Name)
					less = leftName < rightName
					equal = leftName == rightName
				} else {
					less = left.ParticipantPaidAt.String < right.ParticipantPaidAt.String
				}
			} else if left.ParticipantPaidAt.Valid != right.ParticipantPaidAt.Valid {
				less = right.ParticipantPaidAt.Valid
			} else {
				leftName := strings.ToLower(left.Name)
				rightName := strings.ToLower(right.Name)
				less = leftName < rightName
				equal = leftName == rightName
			}
		default:
			leftName := strings.ToLower(left.Name)
			rightName := strings.ToLower(right.Name)
			less = leftName < rightName
			equal = leftName == rightName
		}

		if equal {
			return false
		}

		if query.Dir != "desc" {
			return less
		}
		return !less
	})

	// Calculate paid/unpaid amounts and balance
	// If event is unpaid: balance = -totalPaid (we paid out but haven't received)
	// If event is paid: balance = event.Amount - totalPaid (received minus paid out)
	var totalPaid, totalUnpaid int64
	for _, p := range allParticipants {
		amount := p.ParticipantAmount + p.ParticipantExpense
		if p.ParticipantPaid == 1 {
			totalPaid += amount
		} else {
			totalUnpaid += amount
		}
	}

	var filteredPaid, filteredUnpaid int64
	for _, p := range participants {
		amount := p.ParticipantAmount + p.ParticipantExpense
		if p.ParticipantPaid == 1 {
			filteredPaid += amount
		} else {
			filteredUnpaid += amount
		}
	}

	balance := event.Amount - totalPaid
	filteredBalance := event.Amount - filteredPaid

	slog.Info("event.show.data", "event_id", eventID, "participants", len(participants), "members_total", len(members), "members_filtered", len(filteredMembers), "balance", balance)

	return EventData{
		Title:             "bandcash - " + event.Title,
		Event:             &event,
		Participants:      participants,
		WizardRows:        wizardRows,
		Query:             query,
		Members:           filteredMembers,
		AllMembers:        members,
		Balance:           balance,
		TotalPaid:         totalPaid,
		TotalUnpaid:       totalUnpaid,
		FilteredPaid:      filteredPaid,
		FilteredUnpaid:    filteredUnpaid,
		FilteredBalance:   filteredBalance,
		WizardEventAmount: event.Amount,
		WizardError:       "",
		EditorMode:        "read",
		GroupID:           groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "events.title"), Href: "/groups/" + groupID + "/events"},
			{Label: event.Title},
		},
		ParticipantsTable: EventParticipantsTableLayout(),
	}, nil
}

func GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (EventsData, error) {
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	filters := eventstore.EventTableFilter{
		GroupID: groupID,
		Search:  query.Search,
		Year:    query.Year,
		From:    query.From,
		To:      query.To,
	}

	totalItems, err := eventstore.CountEventsTable(ctx, filters)
	if err != nil {
		return EventsData{}, err
	}
	query = utils.ClampPage(query, totalItems)

	events, err := eventstore.ListEventsTable(ctx, eventstore.EventTableListParams{
		EventTableFilter: filters,
		Sort:             query.Sort,
		Dir:              query.Dir,
		Limit:            query.PageSize,
		Offset:           int(query.Offset()),
	})
	if err != nil {
		return EventsData{}, err
	}

	incomeTotals, err := eventstore.SumEventIncomeTotalsTable(ctx, filters)
	if err != nil {
		return EventsData{}, err
	}

	participantTotals, err := eventstore.SumParticipantTotalsByGroupTable(ctx, filters)
	if err != nil {
		return EventsData{}, err
	}

	expenseTotals, err := expensestore.SumExpenseTotalsTable(ctx, expensestore.ExpenseTableFilter{
		GroupID: groupID,
		Search:  query.Search,
		Year:    query.Year,
		From:    query.From,
		To:      query.To,
	})
	if err != nil {
		return EventsData{}, err
	}

	totals := eventCalcTotals{
		TotalItems:    totalItems,
		Total:         incomeTotals.Total,
		IncomePaid:    incomeTotals.Paid,
		IncomeUnpaid:  incomeTotals.Total - incomeTotals.Paid,
		Paid:          participantTotals.TotalPaid,
		Unpaid:        participantTotals.TotalUnpaid,
		ExpensePaid:   expenseTotals.Paid,
		ExpenseUnpaid: expenseTotals.Total - expenseTotals.Paid,
	}

	return buildEventsData(ctx, groupID, group, query, events, totals)
}

type eventCalcTotals struct {
	TotalItems    int64
	Total         int64
	IncomePaid    int64
	IncomeUnpaid  int64
	Paid          int64
	Unpaid        int64
	ExpensePaid   int64
	ExpenseUnpaid int64
}

func buildEventsData(ctx context.Context, groupID string, group db.Group, query utils.TableQuery, events []db.Event, totals eventCalcTotals) (EventsData, error) {
	query = utils.ClampPage(query, totals.TotalItems)

	// Calculate group totals for display
	groupTotals, err := utils.CalculateGroupTotals(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	admin, err := authstore.GetUserByID(ctx, group.AdminUserID)
	if err != nil {
		return EventsData{}, err
	}

	groupCreatedAt := "-"
	if group.CreatedAt.Valid {
		groupCreatedAt = utils.FormatTimeLocalized(ctx, group.CreatedAt.Time)
	}

	return EventsData{
		Title:                  ctxi18n.T(ctx, "events.page_title"),
		GroupName:              group.Name,
		GroupAdminEmail:        admin.Email,
		GroupCreatedAt:         groupCreatedAt,
		Events:                 events,
		RecentYears:            utils.RecentYears(3),
		Query:                  query,
		Pager:                  utils.BuildTablePagination(totals.TotalItems, query),
		GroupID:                groupID,
		TotalEventAmount:       groupTotals.Income.All,
		TotalPaid:              groupTotals.Income.Paid,
		TotalUnpaid:            groupTotals.Income.Unpaid,
		FilteredTotal:          totals.Total,
		FilteredIncomePaid:     totals.IncomePaid,
		FilteredIncomeUnpaid:   totals.IncomeUnpaid,
		FilteredPayoutsPaid:    totals.Paid,
		FilteredPayoutsUnpaid:  totals.Unpaid,
		FilteredExpensesPaid:   totals.ExpensePaid,
		FilteredExpensesUnpaid: totals.ExpenseUnpaid,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "events.title")},
		},
		EventsTable: EventsIndexTableLayout(),
	}, nil
}
