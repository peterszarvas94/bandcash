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
	return utils.TableQuerySpec{
		DefaultSort: "createdAt",
		DefaultDir:  "desc",
		AllowedSorts: map[string]struct{}{
			"name":      {},
			"createdAt": {},
		},
		AllowedPageSizes: map[int]struct{}{
			10:  {},
			50:  {},
			100: {},
			200: {},
		},
		DefaultSize:  50,
		MaxSearchLen: 100,
	}
}

func (m *GroupModel) GetGroupsPageData(ctx context.Context, userID string, query utils.TableQuery) (GroupsPageData, error) {
	// Get total count
	total, err := db.Qry.CountUserGroupsFiltered(ctx, db.CountUserGroupsFilteredParams{
		UserID: userID,
		Search: query.Search,
	})
	if err != nil {
		slog.Error("group.model: failed to count groups", "err", err)
		return GroupsPageData{}, err
	}

	// Adjust page if needed
	query = utils.ClampPage(query, total)

	// Fetch paginated groups based on sort
	limit := int64(query.PageSize)
	offset := query.Offset()

	var allGroups []GroupWithRole

	switch query.Sort {
	case "name":
		if query.Dir == "desc" {
			rows, err := db.Qry.ListUserGroupsByNameDescFiltered(ctx, db.ListUserGroupsByNameDescFilteredParams{
				UserID: userID,
				Search: query.Search,
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				slog.Error("group.model: failed to list groups by name desc", "err", err)
				return GroupsPageData{}, err
			}
			allGroups = convertNameDescRowsToGroupWithRole(ctx, rows)
		} else {
			rows, err := db.Qry.ListUserGroupsByNameAscFiltered(ctx, db.ListUserGroupsByNameAscFilteredParams{
				UserID: userID,
				Search: query.Search,
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				slog.Error("group.model: failed to list groups by name asc", "err", err)
				return GroupsPageData{}, err
			}
			allGroups = convertNameAscRowsToGroupWithRole(ctx, rows)
		}
	case "createdAt":
		if query.Dir == "asc" {
			rows, err := db.Qry.ListUserGroupsByCreatedAscFiltered(ctx, db.ListUserGroupsByCreatedAscFilteredParams{
				UserID: userID,
				Search: query.Search,
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				slog.Error("group.model: failed to list groups by created asc", "err", err)
				return GroupsPageData{}, err
			}
			allGroups = convertCreatedAscRowsToGroupWithRole(ctx, rows)
		} else {
			rows, err := db.Qry.ListUserGroupsByCreatedDescFiltered(ctx, db.ListUserGroupsByCreatedDescFilteredParams{
				UserID: userID,
				Search: query.Search,
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				slog.Error("group.model: failed to list groups by created desc", "err", err)
				return GroupsPageData{}, err
			}
			allGroups = convertCreatedDescRowsToGroupWithRole(ctx, rows)
		}
	default:
		// Default to createdAt desc
		rows, err := db.Qry.ListUserGroupsByCreatedDescFiltered(ctx, db.ListUserGroupsByCreatedDescFilteredParams{
			UserID: userID,
			Search: query.Search,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			slog.Error("group.model: failed to list groups default", "err", err)
			return GroupsPageData{}, err
		}
		allGroups = convertCreatedDescRowsToGroupWithRole(ctx, rows)
	}

	pagination := utils.BuildTablePagination(total, query)

	return GroupsPageData{
		AllGroups:  allGroups,
		Query:      query,
		Pagination: pagination,
	}, nil
}

func convertNameDescRowsToGroupWithRole(ctx context.Context, rows []db.ListUserGroupsByNameDescFilteredRow) []GroupWithRole {
	result := make([]GroupWithRole, len(rows))
	for i, r := range rows {
		admin, _ := db.Qry.GetUserByID(ctx, r.AdminUserID)
		viewers, _ := db.Qry.GetGroupReaders(ctx, r.ID)
		result[i] = GroupWithRole{
			Group: db.Group{
				ID:          r.ID,
				Name:        r.Name,
				AdminUserID: r.AdminUserID,
				CreatedAt:   r.CreatedAt,
			},
			Role:        r.Role,
			ViewerCount: len(viewers),
			AdminEmail:  admin.Email,
		}
	}
	return result
}

func convertNameAscRowsToGroupWithRole(ctx context.Context, rows []db.ListUserGroupsByNameAscFilteredRow) []GroupWithRole {
	result := make([]GroupWithRole, len(rows))
	for i, r := range rows {
		admin, _ := db.Qry.GetUserByID(ctx, r.AdminUserID)
		viewers, _ := db.Qry.GetGroupReaders(ctx, r.ID)
		result[i] = GroupWithRole{
			Group: db.Group{
				ID:          r.ID,
				Name:        r.Name,
				AdminUserID: r.AdminUserID,
				CreatedAt:   r.CreatedAt,
			},
			Role:        r.Role,
			ViewerCount: len(viewers),
			AdminEmail:  admin.Email,
		}
	}
	return result
}

func convertCreatedAscRowsToGroupWithRole(ctx context.Context, rows []db.ListUserGroupsByCreatedAscFilteredRow) []GroupWithRole {
	result := make([]GroupWithRole, len(rows))
	for i, r := range rows {
		admin, _ := db.Qry.GetUserByID(ctx, r.AdminUserID)
		viewers, _ := db.Qry.GetGroupReaders(ctx, r.ID)
		result[i] = GroupWithRole{
			Group: db.Group{
				ID:          r.ID,
				Name:        r.Name,
				AdminUserID: r.AdminUserID,
				CreatedAt:   r.CreatedAt,
			},
			Role:        r.Role,
			ViewerCount: len(viewers),
			AdminEmail:  admin.Email,
		}
	}
	return result
}

func convertCreatedDescRowsToGroupWithRole(ctx context.Context, rows []db.ListUserGroupsByCreatedDescFilteredRow) []GroupWithRole {
	result := make([]GroupWithRole, len(rows))
	for i, r := range rows {
		admin, _ := db.Qry.GetUserByID(ctx, r.AdminUserID)
		viewers, _ := db.Qry.GetGroupReaders(ctx, r.ID)
		result[i] = GroupWithRole{
			Group: db.Group{
				ID:          r.ID,
				Name:        r.Name,
				AdminUserID: r.AdminUserID,
				CreatedAt:   r.CreatedAt,
			},
			Role:        r.Role,
			ViewerCount: len(viewers),
			AdminEmail:  admin.Email,
		}
	}
	return result
}
