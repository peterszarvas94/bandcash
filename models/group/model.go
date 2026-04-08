package group

import (
	"context"
	"log/slog"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type GroupModel struct {
}

func NewModel() *GroupModel {
	return &GroupModel{}
}

func (m *GroupModel) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "name",
		DefaultDir:   "asc",
		AllowedSorts: []string{"name"},
	})
}

func (m *GroupModel) GetGroupsPageData(ctx context.Context, userID string, query utils.TableQuery) (GroupsPageData, error) {
	total, err := db.Qry.CountUserGroupsFiltered(ctx, db.CountUserGroupsFilteredParams{
		UserID: userID,
		Search: query.Search,
	})
	if err != nil {
		slog.Error("group.model: failed to count groups", "err", err)
		return GroupsPageData{}, err
	}

	allGroups := make([]GroupWithRole, 0)
	if total > 0 {
		rows, err := db.Qry.ListUserGroupsByNameAscFiltered(ctx, db.ListUserGroupsByNameAscFilteredParams{
			UserID: userID,
			Search: query.Search,
			Limit:  total,
			Offset: 0,
		})
		if err != nil {
			slog.Error("group.model: failed to list groups", "err", err)
			return GroupsPageData{}, err
		}
		allGroups = convertNameAscRowsToGroupWithRole(rows)
		if err := m.enrichBalances(ctx, allGroups); err != nil {
			return GroupsPageData{}, err
		}
	}

	return GroupsPageData{
		AllGroups: allGroups,
		Query:     query,
	}, nil
}

func convertNameAscRowsToGroupWithRole(rows []db.ListUserGroupsByNameAscFilteredRow) []GroupWithRole {
	result := make([]GroupWithRole, len(rows))
	for i, r := range rows {
		result[i] = GroupWithRole{
			Group: db.Group{
				ID:          r.ID,
				Name:        r.Name,
				AdminUserID: r.AdminUserID,
				CreatedAt:   r.CreatedAt,
			},
			Role: r.Role,
		}
	}
	return result
}

func (m *GroupModel) enrichBalances(ctx context.Context, groups []GroupWithRole) error {
	for i := range groups {
		totals, err := utils.CalculateGroupTotals(ctx, groups[i].Group.ID)
		if err != nil {
			slog.Error("group.model: failed to calculate balance", "group_id", groups[i].Group.ID, "err", err)
			return err
		}

		// Paid balance mirrors the first number on the balance summary card.
		groups[i].Balance = totals.EventPaid - totals.PayoutPaid - totals.ExpensePaid
	}
	return nil
}
