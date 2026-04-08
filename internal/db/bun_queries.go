package db

import (
	"context"
	"database/sql"
)

// Database query helpers used by application layers.

func q() *Queries { return New(DB) }

func UpdateUserPreferredLang(ctx context.Context, arg UpdateUserPreferredLangParams) error {
	return q().UpdateUserPreferredLang(ctx, arg)
}

func UpsertUserDetailCardState(ctx context.Context, arg UpsertUserDetailCardStateParams) error {
	return q().UpsertUserDetailCardState(ctx, arg)
}

func ListUserSessions(ctx context.Context, userID string) ([]UserSession, error) {
	return q().ListUserSessions(ctx, userID)
}

func GetUserSessionByToken(ctx context.Context, token string) (UserSession, error) {
	return q().GetUserSessionByToken(ctx, token)
}

func DeleteUserSession(ctx context.Context, arg DeleteUserSessionParams) error {
	return q().DeleteUserSession(ctx, arg)
}

func DeleteAllUserSessions(ctx context.Context, userID string) error {
	return q().DeleteAllUserSessions(ctx, userID)
}

func ListUserDetailCardStates(ctx context.Context, userID string) ([]ListUserDetailCardStatesRow, error) {
	return q().ListUserDetailCardStates(ctx, userID)
}

func IsUserBanned(ctx context.Context, userID string) (int64, error) {
	return q().IsUserBanned(ctx, userID)
}

func GetGroupAccessRole(ctx context.Context, arg GetGroupAccessRoleParams) (string, error) {
	return q().GetGroupAccessRole(ctx, arg)
}

func ListEvents(ctx context.Context, groupID string) ([]Event, error) {
	return q().ListEvents(ctx, groupID)
}

func ListExpenses(ctx context.Context, groupID string) ([]Expense, error) {
	return q().ListExpenses(ctx, groupID)
}

func SumParticipantPaidAmountsByGroup(ctx context.Context, groupID string) (SumParticipantPaidAmountsByGroupRow, error) {
	return q().SumParticipantPaidAmountsByGroup(ctx, groupID)
}

func GetAppFlagBool(ctx context.Context, key string) (int64, error) {
	return q().GetAppFlagBool(ctx, key)
}

func UpsertAppFlagBool(ctx context.Context, arg UpsertAppFlagBoolParams) error {
	return q().UpsertAppFlagBool(ctx, arg)
}

func CountUsersFiltered(ctx context.Context, search string) (int64, error) {
	return q().CountUsersFiltered(ctx, search)
}

func ListUsersByEmailDescFiltered(ctx context.Context, arg ListUsersByEmailDescFilteredParams) ([]ListUsersByEmailDescFilteredRow, error) {
	return q().ListUsersByEmailDescFiltered(ctx, arg)
}

func ListUsersByEmailAscFiltered(ctx context.Context, arg ListUsersByEmailAscFilteredParams) ([]ListUsersByEmailAscFilteredRow, error) {
	return q().ListUsersByEmailAscFiltered(ctx, arg)
}

func ListUsersByCreatedAscFiltered(ctx context.Context, arg ListUsersByCreatedAscFilteredParams) ([]ListUsersByCreatedAscFilteredRow, error) {
	return q().ListUsersByCreatedAscFiltered(ctx, arg)
}

func ListUsersByCreatedDescFiltered(ctx context.Context, arg ListUsersByCreatedDescFilteredParams) ([]ListUsersByCreatedDescFilteredRow, error) {
	return q().ListUsersByCreatedDescFiltered(ctx, arg)
}

func CountGroupsFiltered(ctx context.Context, search string) (int64, error) {
	return q().CountGroupsFiltered(ctx, search)
}

func ListGroupsByNameDescFiltered(ctx context.Context, arg ListGroupsByNameDescFilteredParams) ([]Group, error) {
	return q().ListGroupsByNameDescFiltered(ctx, arg)
}

func ListGroupsByNameAscFiltered(ctx context.Context, arg ListGroupsByNameAscFilteredParams) ([]Group, error) {
	return q().ListGroupsByNameAscFiltered(ctx, arg)
}

func ListGroupsByCreatedAscFiltered(ctx context.Context, arg ListGroupsByCreatedAscFilteredParams) ([]Group, error) {
	return q().ListGroupsByCreatedAscFiltered(ctx, arg)
}

func ListGroupsByCreatedDescFiltered(ctx context.Context, arg ListGroupsByCreatedDescFilteredParams) ([]Group, error) {
	return q().ListGroupsByCreatedDescFiltered(ctx, arg)
}

func CountSessionsFiltered(ctx context.Context, search string) (int64, error) {
	return q().CountSessionsFiltered(ctx, search)
}

func ListSessionsByEmailDescFiltered(ctx context.Context, arg ListSessionsByEmailDescFilteredParams) ([]ListSessionsByEmailDescFilteredRow, error) {
	return q().ListSessionsByEmailDescFiltered(ctx, arg)
}

func ListSessionsByEmailAscFiltered(ctx context.Context, arg ListSessionsByEmailAscFilteredParams) ([]ListSessionsByEmailAscFilteredRow, error) {
	return q().ListSessionsByEmailAscFiltered(ctx, arg)
}

func ListSessionsByCreatedAscFiltered(ctx context.Context, arg ListSessionsByCreatedAscFilteredParams) ([]ListSessionsByCreatedAscFilteredRow, error) {
	return q().ListSessionsByCreatedAscFiltered(ctx, arg)
}

func ListSessionsByCreatedDescFiltered(ctx context.Context, arg ListSessionsByCreatedDescFilteredParams) ([]ListSessionsByCreatedDescFilteredRow, error) {
	return q().ListSessionsByCreatedDescFiltered(ctx, arg)
}

func BanUser(ctx context.Context, arg BanUserParams) error {
	return q().BanUser(ctx, arg)
}

func UnbanUser(ctx context.Context, userID string) error {
	return q().UnbanUser(ctx, userID)
}

func GetUserByEmail(ctx context.Context, email string) (User, error) {
	return q().GetUserByEmail(ctx, email)
}

func CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	return q().CreateUser(ctx, arg)
}

func CreateMagicLink(ctx context.Context, arg CreateMagicLinkParams) (MagicLink, error) {
	return q().CreateMagicLink(ctx, arg)
}

func GetMagicLinkByToken(ctx context.Context, token string) (MagicLink, error) {
	return q().GetMagicLinkByToken(ctx, token)
}

func UseMagicLink(ctx context.Context, id string) error {
	return q().UseMagicLink(ctx, id)
}

func CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (UserSession, error) {
	return q().CreateUserSession(ctx, arg)
}

func RemoveGroupReader(ctx context.Context, arg RemoveGroupReaderParams) error {
	return q().RemoveGroupReader(ctx, arg)
}

func CreateGroupAdmin(ctx context.Context, arg CreateGroupAdminParams) (CreateGroupAdminRow, error) {
	return q().CreateGroupAdmin(ctx, arg)
}

func RemoveGroupAdmin(ctx context.Context, arg RemoveGroupAdminParams) error {
	return q().RemoveGroupAdmin(ctx, arg)
}

func CreateGroupReader(ctx context.Context, arg CreateGroupReaderParams) (CreateGroupReaderRow, error) {
	return q().CreateGroupReader(ctx, arg)
}

func ListGroupsByAdmin(ctx context.Context, userID string) ([]Group, error) {
	return q().ListGroupsByAdmin(ctx, userID)
}

func ListGroupsByReader(ctx context.Context, userID string) ([]Group, error) {
	return q().ListGroupsByReader(ctx, userID)
}

func CountParticipantsByMemberFiltered(ctx context.Context, arg CountParticipantsByMemberFilteredParams) (int64, error) {
	return q().CountParticipantsByMemberFiltered(ctx, arg)
}

func SumParticipantTotalsByMemberFiltered(ctx context.Context, arg SumParticipantTotalsByMemberFilteredParams) (SumParticipantTotalsByMemberFilteredRow, error) {
	return q().SumParticipantTotalsByMemberFiltered(ctx, arg)
}

func ListParticipantsByMemberByTitleDescFiltered(ctx context.Context, arg ListParticipantsByMemberByTitleDescFilteredParams) ([]ListParticipantsByMemberByTitleDescFilteredRow, error) {
	return q().ListParticipantsByMemberByTitleDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByTitleAscFiltered(ctx context.Context, arg ListParticipantsByMemberByTitleAscFilteredParams) ([]ListParticipantsByMemberByTitleAscFilteredRow, error) {
	return q().ListParticipantsByMemberByTitleAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByCutDescFiltered(ctx context.Context, arg ListParticipantsByMemberByCutDescFilteredParams) ([]ListParticipantsByMemberByCutDescFilteredRow, error) {
	return q().ListParticipantsByMemberByCutDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByCutAscFiltered(ctx context.Context, arg ListParticipantsByMemberByCutAscFilteredParams) ([]ListParticipantsByMemberByCutAscFilteredRow, error) {
	return q().ListParticipantsByMemberByCutAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByExpenseDescFiltered(ctx context.Context, arg ListParticipantsByMemberByExpenseDescFilteredParams) ([]ListParticipantsByMemberByExpenseDescFilteredRow, error) {
	return q().ListParticipantsByMemberByExpenseDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByExpenseAscFiltered(ctx context.Context, arg ListParticipantsByMemberByExpenseAscFilteredParams) ([]ListParticipantsByMemberByExpenseAscFilteredRow, error) {
	return q().ListParticipantsByMemberByExpenseAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByPaidDescFiltered(ctx context.Context, arg ListParticipantsByMemberByPaidDescFilteredParams) ([]ListParticipantsByMemberByPaidDescFilteredRow, error) {
	return q().ListParticipantsByMemberByPaidDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByPaidAscFiltered(ctx context.Context, arg ListParticipantsByMemberByPaidAscFilteredParams) ([]ListParticipantsByMemberByPaidAscFilteredRow, error) {
	return q().ListParticipantsByMemberByPaidAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByPaidAtDescFiltered(ctx context.Context, arg ListParticipantsByMemberByPaidAtDescFilteredParams) ([]ListParticipantsByMemberByPaidAtDescFilteredRow, error) {
	return q().ListParticipantsByMemberByPaidAtDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByPaidAtAscFiltered(ctx context.Context, arg ListParticipantsByMemberByPaidAtAscFilteredParams) ([]ListParticipantsByMemberByPaidAtAscFilteredRow, error) {
	return q().ListParticipantsByMemberByPaidAtAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByTimeAscFiltered(ctx context.Context, arg ListParticipantsByMemberByTimeAscFilteredParams) ([]ListParticipantsByMemberByTimeAscFilteredRow, error) {
	return q().ListParticipantsByMemberByTimeAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByTimeDescFiltered(ctx context.Context, arg ListParticipantsByMemberByTimeDescFilteredParams) ([]ListParticipantsByMemberByTimeDescFilteredRow, error) {
	return q().ListParticipantsByMemberByTimeDescFiltered(ctx, arg)
}

func ListParticipantsByEvent(ctx context.Context, arg ListParticipantsByEventParams) ([]ListParticipantsByEventRow, error) {
	return q().ListParticipantsByEvent(ctx, arg)
}

func SumParticipantTotalsByGroupFiltered(ctx context.Context, arg SumParticipantTotalsByGroupFilteredParams) (SumParticipantTotalsByGroupFilteredRow, error) {
	return q().SumParticipantTotalsByGroupFiltered(ctx, arg)
}

func SumExpenseTotalsFiltered(ctx context.Context, arg SumExpenseTotalsFilteredParams) (SumExpenseTotalsFilteredRow, error) {
	return q().SumExpenseTotalsFiltered(ctx, arg)
}

func CreateGroup(ctx context.Context, arg CreateGroupParams) (Group, error) {
	return q().CreateGroup(ctx, arg)
}

func UpdateGroupName(ctx context.Context, arg UpdateGroupNameParams) (Group, error) {
	return q().UpdateGroupName(ctx, arg)
}

func DeleteGroup(ctx context.Context, id string) error {
	return q().DeleteGroup(ctx, id)
}

func ListGroupPendingInvites(ctx context.Context, groupID sql.NullString) ([]MagicLink, error) {
	return q().ListGroupPendingInvites(ctx, groupID)
}

func CreateInviteMagicLink(ctx context.Context, arg CreateInviteMagicLinkParams) (MagicLink, error) {
	return q().CreateInviteMagicLink(ctx, arg)
}

func UpdateGroupAdmin(ctx context.Context, arg UpdateGroupAdminParams) error {
	return q().UpdateGroupAdmin(ctx, arg)
}

func DeleteGroupPendingInvite(ctx context.Context, arg DeleteGroupPendingInviteParams) error {
	return q().DeleteGroupPendingInvite(ctx, arg)
}

func ListGroupUserAccess(ctx context.Context, groupID string) ([]ListGroupUserAccessRow, error) {
	return q().ListGroupUserAccess(ctx, groupID)
}

func ListGroupAdmins(ctx context.Context, groupID string) ([]User, error) {
	return q().ListGroupAdmins(ctx, groupID)
}

func IsGroupReader(ctx context.Context, arg IsGroupReaderParams) (int64, error) {
	return q().IsGroupReader(ctx, arg)
}

func ListUnpaidEventsByGroup(ctx context.Context, groupID string) ([]Event, error) {
	return q().ListUnpaidEventsByGroup(ctx, groupID)
}

func ListUnpaidOutgoingPaymentsByGroup(ctx context.Context, groupID string) ([]GroupOutgoingPayment, error) {
	return q().ListUnpaidOutgoingPaymentsByGroup(ctx, groupID)
}

func ListPaidEventsByGroup(ctx context.Context, groupID string) ([]Event, error) {
	return q().ListPaidEventsByGroup(ctx, groupID)
}

func ListPaidOutgoingPaymentsByGroup(ctx context.Context, groupID string) ([]GroupOutgoingPayment, error) {
	return q().ListPaidOutgoingPaymentsByGroup(ctx, groupID)
}
