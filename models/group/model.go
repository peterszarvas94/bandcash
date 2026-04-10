package group

import (
	"context"
	"log/slog"

	"bandcash/internal/db"
	"bandcash/internal/utils"
	groupstore "bandcash/models/group/data"
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
	total, err := groupstore.CountUserGroupsTable(ctx, userID, query.Search)
	if err != nil {
		slog.Error("group.model: failed to count groups", "err", err)
		return GroupsPageData{}, err
	}

	allGroups := make([]GroupWithRole, 0)
	if total > 0 {
		rows, err := groupstore.ListUserGroupsTable(ctx, userID, query.Search, int(total), 0)
		if err != nil {
			slog.Error("group.model: failed to list groups", "err", err)
			return GroupsPageData{}, err
		}
		allGroups = convertUserGroupRowsToGroupWithRole(rows)
		if err := m.enrichBalances(ctx, allGroups); err != nil {
			return GroupsPageData{}, err
		}
	}

	return GroupsPageData{
		AllGroups: allGroups,
		Query:     query,
	}, nil
}

func convertUserGroupRowsToGroupWithRole(rows []groupstore.UserGroupRow) []GroupWithRole {
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

		// Current balance mirrors the "All" value on the events balance card.
		groups[i].Balance = totals.Balance.All
	}
	return nil
}
