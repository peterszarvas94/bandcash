package event

import (
	"context"
	"log/slog"
	"sort"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type Events struct {
}

func (e *Events) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "time",
		DefaultDir:   "desc",
		AllowedSorts: []string{"time", "title", "amount", "description", "paid"},
	})
}

func New() *Events {
	return &Events{}
}

func (e *Events) ParticipantTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "name",
		DefaultDir:   "asc",
		AllowedSorts: []string{"name", "amount", "expense", "total", "paid"},
	})
}

func (e *Events) GetShowData(ctx context.Context, groupID, eventID string, query utils.TableQuery) (EventData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return EventData{}, err
	}

	event, err := db.Qry.GetEvent(ctx, db.GetEventParams{
		ID:      eventID,
		GroupID: groupID,
	})
	if err != nil {
		return EventData{}, err
	}

	allParticipants, err := db.Qry.ListParticipantsByEvent(ctx, db.ListParticipantsByEventParams{
		EventID: eventID,
		GroupID: groupID,
	})
	if err != nil {
		return EventData{}, err
	}

	members, err := db.Qry.ListMembers(ctx, groupID)
	if err != nil {
		return EventData{}, err
	}

	participantMemberIDs := make(map[string]bool, len(allParticipants))
	participantByMemberID := make(map[string]db.ListParticipantsByEventRow, len(allParticipants))
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
				MemberID:   member.ID,
				MemberName: member.Name,
				Included:   true,
				Amount:     participant.ParticipantAmount,
				Expense:    participant.ParticipantExpense,
			})
		}

		if participantMemberIDs[member.ID] {
			continue
		}
		filteredMembers = append(filteredMembers, member)
	}

	participants := allParticipants
	if query.Search != "" {
		search := strings.ToLower(strings.TrimSpace(query.Search))
		filtered := make([]db.ListParticipantsByEventRow, 0, len(allParticipants))
		for _, participant := range allParticipants {
			if strings.Contains(strings.ToLower(participant.Name), search) {
				filtered = append(filtered, participant)
			}
		}
		participants = filtered
	}

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

	totalItems := int64(len(participants))
	query = utils.ClampPage(query, totalItems)

	start := int(query.Offset())
	if start > len(participants) {
		start = len(participants)
	}
	end := start + query.PageSize
	if end > len(participants) {
		end = len(participants)
	}
	displayParticipants := participants[start:end]

	// Calculate paid/unpaid amounts and leftover
	// If event is unpaid: leftover = -totalPaid (we paid out but haven't received)
	// If event is paid: leftover = event.Amount - totalPaid (received minus paid out)
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

	leftover := event.Amount - totalPaid
	filteredLeftover := event.Amount - filteredPaid

	slog.Info("event.show.data", "event_id", eventID, "participants", len(participants), "members_total", len(members), "members_filtered", len(filteredMembers), "leftover", leftover)

	return EventData{
		Title:                "Bandcash - " + event.Title,
		Event:                &event,
		Participants:         displayParticipants,
		WizardRows:           wizardRows,
		Query:                query,
		Pager:                utils.BuildTablePagination(totalItems, query),
		Members:              filteredMembers,
		AllMembers:           members,
		WizardAddableMembers: filteredMembers,
		Leftover:             leftover,
		TotalPaid:            totalPaid,
		TotalUnpaid:          totalUnpaid,
		FilteredPaid:         filteredPaid,
		FilteredUnpaid:       filteredUnpaid,
		FilteredLeftover:     filteredLeftover,
		WizardEventAmount:    event.Amount,
		WizardError:          "",
		EditorMode:           "read",
		GroupID:              groupID,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "events.title"), Href: "/groups/" + groupID + "/events"},
			{Label: event.Title},
		},
		ParticipantsTable: utils.EventParticipantsTableLayout(),
	}, nil
}

func (e *Events) GetIndexData(ctx context.Context, groupID string, query utils.TableQuery) (EventsData, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	// Check cache first
	cacheKey := EventsFilterKey(groupID, query.Search, query.Year, query.From, query.To, query.Sort, query.Dir)
	if cached, ok := utils.CalcCacheInstance.Get(cacheKey); ok {
		if result, valid := cached.(eventCalcTotals); valid {
			return e.buildEventsData(ctx, groupID, group, query, result)
		}
	}

	// Get all events for the group to calculate in-memory
	allEvents, err := db.Qry.ListEvents(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	// Filter and calculate totals in-memory
	filteredEvents := make([]db.Event, 0, len(allEvents))
	var filteredTotal, filteredIncomePaid int64

	for _, event := range allEvents {
		// Apply filters
		if matchesFilters(event, query) {
			filteredEvents = append(filteredEvents, event)
			filteredTotal += event.Amount
			if event.Paid == 1 {
				filteredIncomePaid += event.Amount
			}
		}
	}

	participantTotals, err := db.Qry.SumParticipantTotalsByGroupFiltered(ctx, db.SumParticipantTotalsByGroupFilteredParams{
		GroupID: groupID,
		Search:  query.Search,
		Year:    query.Year,
		From:    query.From,
		To:      query.To,
	})
	if err != nil {
		return EventsData{}, err
	}

	expenseTotals, err := db.Qry.SumExpenseTotalsFiltered(ctx, db.SumExpenseTotalsFilteredParams{
		GroupID:    groupID,
		Search:     query.Search,
		YearFilter: query.Year,
		FromDate:   query.From,
		ToDate:     query.To,
	})
	if err != nil {
		return EventsData{}, err
	}

	// Sort filtered events
	sortEvents(filteredEvents, query.Sort, query.Dir)

	// Store in cache
	totals := eventCalcTotals{
		Filtered:      filteredEvents,
		Total:         filteredTotal,
		IncomePaid:    filteredIncomePaid,
		IncomeUnpaid:  filteredTotal - filteredIncomePaid,
		Paid:          participantTotals.TotalPaid,
		Unpaid:        participantTotals.TotalUnpaid,
		ExpensePaid:   expenseTotals.TotalPaid,
		ExpenseUnpaid: expenseTotals.TotalUnpaid,
	}
	utils.CalcCacheInstance.Set(cacheKey, totals)

	return e.buildEventsData(ctx, groupID, group, query, totals)
}

type eventCalcTotals struct {
	Filtered      []db.Event
	Total         int64
	IncomePaid    int64
	IncomeUnpaid  int64
	Paid          int64
	Unpaid        int64
	ExpensePaid   int64
	ExpenseUnpaid int64
}

func matchesFilters(event db.Event, query utils.TableQuery) bool {
	// Search filter
	if query.Search != "" {
		searchLower := strings.ToLower(query.Search)
		if !strings.Contains(strings.ToLower(event.Title), searchLower) &&
			!strings.Contains(strings.ToLower(event.Description), searchLower) {
			return false
		}
	}

	// Year filter
	if query.Year != "" && !strings.HasPrefix(event.Time, query.Year) {
		return false
	}

	// Date range filters
	if query.From != "" && event.Time < query.From {
		return false
	}
	if query.To != "" && event.Time > query.To {
		return false
	}

	return true
}

func sortEvents(events []db.Event, sortField, dir string) {
	less := func(i, j int) bool {
		switch sortField {
		case "title":
			if dir == "desc" {
				return events[i].Title > events[j].Title
			}
			return events[i].Title < events[j].Title
		case "amount":
			if dir == "desc" {
				return events[i].Amount > events[j].Amount
			}
			return events[i].Amount < events[j].Amount
		case "description":
			if dir == "desc" {
				return events[i].Description > events[j].Description
			}
			return events[i].Description < events[j].Description
		case "paid":
			if dir == "desc" {
				return events[i].Paid > events[j].Paid
			}
			return events[i].Paid < events[j].Paid
		default: // time
			if dir == "desc" {
				return events[i].Time > events[j].Time
			}
			return events[i].Time < events[j].Time
		}
	}
	sort.Slice(events, less)
}

func (e *Events) buildEventsData(ctx context.Context, groupID string, group db.Group, query utils.TableQuery, totals eventCalcTotals) (EventsData, error) {
	totalItems := int64(len(totals.Filtered))
	query = utils.ClampPage(query, totalItems)

	// Paginate
	start := query.Offset()
	end := start + int64(query.PageSize)
	if end > totalItems {
		end = totalItems
	}
	if start > totalItems {
		start = totalItems
	}

	var paginatedEvents []db.Event
	if start < totalItems {
		paginatedEvents = totals.Filtered[start:end]
	}

	// Calculate group totals for display
	groupTotals, err := utils.CalculateGroupTotals(ctx, groupID)
	if err != nil {
		return EventsData{}, err
	}

	return EventsData{
		Title:                  ctxi18n.T(ctx, "events.page_title"),
		Events:                 paginatedEvents,
		RecentYears:            utils.RecentYears(3),
		Query:                  query,
		Pager:                  utils.BuildTablePagination(totalItems, query),
		GroupID:                groupID,
		TotalEventAmount:       groupTotals.TotalEventAmount,
		TotalPaid:              groupTotals.EventPaid,
		TotalUnpaid:            groupTotals.EventUnpaid,
		FilteredTotal:          totals.Total,
		FilteredIncomePaid:     totals.IncomePaid,
		FilteredIncomeUnpaid:   totals.IncomeUnpaid,
		FilteredPayoutsPaid:    totals.Paid,
		FilteredPayoutsUnpaid:  totals.Unpaid,
		FilteredExpensesPaid:   totals.ExpensePaid,
		FilteredExpensesUnpaid: totals.ExpenseUnpaid,
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + groupID},
			{Label: ctxi18n.T(ctx, "events.title")},
		},
		EventsTable: utils.EventsIndexTableLayout(),
	}, nil
}
