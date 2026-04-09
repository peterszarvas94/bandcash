package store

import (
	"context"

	"bandcash/internal/db"
)

func ListMembers(ctx context.Context, groupID string) ([]db.Member, error) {
	rows := make([]db.Member, 0)
	err := db.BunDB.NewSelect().Model(&rows).Where("group_id = ?", groupID).OrderExpr("created_at DESC").Scan(ctx)
	return rows, err
}

func GetMember(ctx context.Context, arg GetMemberParams) (db.Member, error) {
	var row db.Member
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Scan(ctx)
	return row, err
}

func GetMemberByID(ctx context.Context, id string) (db.Member, error) {
	var row db.Member
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	return row, err
}

func CreateMember(ctx context.Context, arg CreateMemberParams) (db.Member, error) {
	member := db.Member{ID: arg.ID, GroupID: arg.GroupID, Name: arg.Name, Description: arg.Description}
	if _, err := db.BunDB.NewInsert().Model(&member).Exec(ctx); err != nil {
		return db.Member{}, err
	}
	return GetMember(ctx, GetMemberParams{ID: arg.ID, GroupID: arg.GroupID})
}

func UpdateMember(ctx context.Context, arg UpdateMemberParams) (db.Member, error) {
	_, err := db.BunDB.NewUpdate().Model((*db.Member)(nil)).
		Set("name = ?", arg.Name).
		Set("description = ?", arg.Description).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return db.Member{}, err
	}
	return GetMember(ctx, GetMemberParams{ID: arg.ID, GroupID: arg.GroupID})
}

func DeleteMember(ctx context.Context, arg DeleteMemberParams) error {
	_, err := db.BunDB.NewDelete().Model((*db.Member)(nil)).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Exec(ctx)
	return err
}

func DeleteMemberByID(ctx context.Context, id string) error {
	_, err := db.BunDB.NewDelete().Model((*db.Member)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
