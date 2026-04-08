package db

import (
	"context"
	"database/sql"
)

// Compatibility wrappers for legacy query methods.
// TODO: replace call sites with Bun-native implementations, then remove.

func AddParticipant(ctx context.Context, arg AddParticipantParams) (Participant, error) {
	return New(DB).AddParticipant(ctx, arg)
}

func BanUser(ctx context.Context, arg BanUserParams) error {
	return New(DB).BanUser(ctx, arg)
}

func CountEvents(ctx context.Context) (int64, error) {
	return New(DB).CountEvents(ctx)
}

func CountEventsFiltered(ctx context.Context, arg CountEventsFilteredParams) (int64, error) {
	return New(DB).CountEventsFiltered(ctx, arg)
}

func CountExpensesFiltered(ctx context.Context, arg CountExpensesFilteredParams) (int64, error) {
	return New(DB).CountExpensesFiltered(ctx, arg)
}

func CountGroupAdminsFiltered(ctx context.Context, arg CountGroupAdminsFilteredParams) (int64, error) {
	return New(DB).CountGroupAdminsFiltered(ctx, arg)
}

func CountGroupPendingInvitesFiltered(ctx context.Context, arg CountGroupPendingInvitesFilteredParams) (int64, error) {
	return New(DB).CountGroupPendingInvitesFiltered(ctx, arg)
}

func CountGroupReadersFiltered(ctx context.Context, arg CountGroupReadersFilteredParams) (int64, error) {
	return New(DB).CountGroupReadersFiltered(ctx, arg)
}

func CountGroups(ctx context.Context) (int64, error) {
	return New(DB).CountGroups(ctx)
}

func CountGroupsFiltered(ctx context.Context, search interface{}) (int64, error) {
	return New(DB).CountGroupsFiltered(ctx, search)
}

func CountMembers(ctx context.Context) (int64, error) {
	return New(DB).CountMembers(ctx)
}

func CountMembersFiltered(ctx context.Context, arg CountMembersFilteredParams) (int64, error) {
	return New(DB).CountMembersFiltered(ctx, arg)
}

func CountParticipantsByMemberFiltered(ctx context.Context, arg CountParticipantsByMemberFilteredParams) (int64, error) {
	return New(DB).CountParticipantsByMemberFiltered(ctx, arg)
}

func CountSessionsFiltered(ctx context.Context, search interface{}) (int64, error) {
	return New(DB).CountSessionsFiltered(ctx, search)
}

func CountUserGroupsFiltered(ctx context.Context, arg CountUserGroupsFilteredParams) (int64, error) {
	return New(DB).CountUserGroupsFiltered(ctx, arg)
}

func CountUsers(ctx context.Context) (int64, error) {
	return New(DB).CountUsers(ctx)
}

func CountUsersFiltered(ctx context.Context, search interface{}) (int64, error) {
	return New(DB).CountUsersFiltered(ctx, search)
}

func CreateGroup(ctx context.Context, arg CreateGroupParams) (Group, error) {
	return New(DB).CreateGroup(ctx, arg)
}

func CreateGroupAdmin(ctx context.Context, arg CreateGroupAdminParams) (CreateGroupAdminRow, error) {
	return New(DB).CreateGroupAdmin(ctx, arg)
}

func CreateGroupReader(ctx context.Context, arg CreateGroupReaderParams) (CreateGroupReaderRow, error) {
	return New(DB).CreateGroupReader(ctx, arg)
}

func CreateInviteMagicLink(ctx context.Context, arg CreateInviteMagicLinkParams) (MagicLink, error) {
	return New(DB).CreateInviteMagicLink(ctx, arg)
}

func CreateMagicLink(ctx context.Context, arg CreateMagicLinkParams) (MagicLink, error) {
	return New(DB).CreateMagicLink(ctx, arg)
}

func CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	return New(DB).CreateUser(ctx, arg)
}

func CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (UserSession, error) {
	return New(DB).CreateUserSession(ctx, arg)
}

func DeleteAllEvents(ctx context.Context) error {
	return New(DB).DeleteAllEvents(ctx)
}

func DeleteAllUserSessions(ctx context.Context, userID string) error {
	return New(DB).DeleteAllUserSessions(ctx, userID)
}

func DeleteExpiredMagicLinks(ctx context.Context) error {
	return New(DB).DeleteExpiredMagicLinks(ctx)
}

func DeleteExpiredUserSessions(ctx context.Context) error {
	return New(DB).DeleteExpiredUserSessions(ctx)
}

func DeleteGroup(ctx context.Context, id string) error {
	return New(DB).DeleteGroup(ctx, id)
}

func DeleteGroupPendingInvite(ctx context.Context, arg DeleteGroupPendingInviteParams) error {
	return New(DB).DeleteGroupPendingInvite(ctx, arg)
}

func DeleteOtherUserSessions(ctx context.Context, arg DeleteOtherUserSessionsParams) error {
	return New(DB).DeleteOtherUserSessions(ctx, arg)
}

func DeleteUserSession(ctx context.Context, arg DeleteUserSessionParams) error {
	return New(DB).DeleteUserSession(ctx, arg)
}

func DeleteUserSessionByID(ctx context.Context, id string) error {
	return New(DB).DeleteUserSessionByID(ctx, id)
}

func GetAppFlagBool(ctx context.Context, key string) (int64, error) {
	return New(DB).GetAppFlagBool(ctx, key)
}

func GetGroupAccessRole(ctx context.Context, arg GetGroupAccessRoleParams) (string, error) {
	return New(DB).GetGroupAccessRole(ctx, arg)
}

func GetGroupByAdmin(ctx context.Context, userID string) (Group, error) {
	return New(DB).GetGroupByAdmin(ctx, userID)
}

func GetGroupReaders(ctx context.Context, groupID string) ([]User, error) {
	return New(DB).GetGroupReaders(ctx, groupID)
}

func GetMagicLinkByToken(ctx context.Context, token string) (MagicLink, error) {
	return New(DB).GetMagicLinkByToken(ctx, token)
}

func GetUserByEmail(ctx context.Context, email string) (User, error) {
	return New(DB).GetUserByEmail(ctx, email)
}

func GetUserSessionByToken(ctx context.Context, token string) (UserSession, error) {
	return New(DB).GetUserSessionByToken(ctx, token)
}

func IsGroupAdmin(ctx context.Context, arg IsGroupAdminParams) (int64, error) {
	return New(DB).IsGroupAdmin(ctx, arg)
}

func IsGroupReader(ctx context.Context, arg IsGroupReaderParams) (int64, error) {
	return New(DB).IsGroupReader(ctx, arg)
}

func IsUserBanned(ctx context.Context, userID string) (int64, error) {
	return New(DB).IsUserBanned(ctx, userID)
}

func ListAllGroups(ctx context.Context) ([]Group, error) {
	return New(DB).ListAllGroups(ctx)
}

func ListEvents(ctx context.Context, groupID string) ([]Event, error) {
	return New(DB).ListEvents(ctx, groupID)
}

func ListEventsByAmountAscFiltered(ctx context.Context, arg ListEventsByAmountAscFilteredParams) ([]Event, error) {
	return New(DB).ListEventsByAmountAscFiltered(ctx, arg)
}

func ListEventsByAmountDescFiltered(ctx context.Context, arg ListEventsByAmountDescFilteredParams) ([]Event, error) {
	return New(DB).ListEventsByAmountDescFiltered(ctx, arg)
}

func ListEventsByDescriptionAscFiltered(ctx context.Context, arg ListEventsByDescriptionAscFilteredParams) ([]Event, error) {
	return New(DB).ListEventsByDescriptionAscFiltered(ctx, arg)
}

func ListEventsByDescriptionDescFiltered(ctx context.Context, arg ListEventsByDescriptionDescFilteredParams) ([]Event, error) {
	return New(DB).ListEventsByDescriptionDescFiltered(ctx, arg)
}

func ListEventsByTimeAscFiltered(ctx context.Context, arg ListEventsByTimeAscFilteredParams) ([]Event, error) {
	return New(DB).ListEventsByTimeAscFiltered(ctx, arg)
}

func ListEventsByTimeDescFiltered(ctx context.Context, arg ListEventsByTimeDescFilteredParams) ([]Event, error) {
	return New(DB).ListEventsByTimeDescFiltered(ctx, arg)
}

func ListEventsByTitleAscFiltered(ctx context.Context, arg ListEventsByTitleAscFilteredParams) ([]Event, error) {
	return New(DB).ListEventsByTitleAscFiltered(ctx, arg)
}

func ListEventsByTitleDescFiltered(ctx context.Context, arg ListEventsByTitleDescFilteredParams) ([]Event, error) {
	return New(DB).ListEventsByTitleDescFiltered(ctx, arg)
}

func ListExpenses(ctx context.Context, groupID string) ([]Expense, error) {
	return New(DB).ListExpenses(ctx, groupID)
}

func ListExpensesByAmountAscFiltered(ctx context.Context, arg ListExpensesByAmountAscFilteredParams) ([]Expense, error) {
	return New(DB).ListExpensesByAmountAscFiltered(ctx, arg)
}

func ListExpensesByAmountDescFiltered(ctx context.Context, arg ListExpensesByAmountDescFilteredParams) ([]Expense, error) {
	return New(DB).ListExpensesByAmountDescFiltered(ctx, arg)
}

func ListExpensesByDateAscFiltered(ctx context.Context, arg ListExpensesByDateAscFilteredParams) ([]Expense, error) {
	return New(DB).ListExpensesByDateAscFiltered(ctx, arg)
}

func ListExpensesByDateDescFiltered(ctx context.Context, arg ListExpensesByDateDescFilteredParams) ([]Expense, error) {
	return New(DB).ListExpensesByDateDescFiltered(ctx, arg)
}

func ListExpensesByDescriptionAscFiltered(ctx context.Context, arg ListExpensesByDescriptionAscFilteredParams) ([]Expense, error) {
	return New(DB).ListExpensesByDescriptionAscFiltered(ctx, arg)
}

func ListExpensesByDescriptionDescFiltered(ctx context.Context, arg ListExpensesByDescriptionDescFilteredParams) ([]Expense, error) {
	return New(DB).ListExpensesByDescriptionDescFiltered(ctx, arg)
}

func ListExpensesByTitleAscFiltered(ctx context.Context, arg ListExpensesByTitleAscFilteredParams) ([]Expense, error) {
	return New(DB).ListExpensesByTitleAscFiltered(ctx, arg)
}

func ListExpensesByTitleDescFiltered(ctx context.Context, arg ListExpensesByTitleDescFilteredParams) ([]Expense, error) {
	return New(DB).ListExpensesByTitleDescFiltered(ctx, arg)
}

func ListGroupAdminUserIDs(ctx context.Context, groupID string) ([]string, error) {
	return New(DB).ListGroupAdminUserIDs(ctx, groupID)
}

func ListGroupAdmins(ctx context.Context, groupID string) ([]User, error) {
	return New(DB).ListGroupAdmins(ctx, groupID)
}

func ListGroupAdminsByEmailAscFiltered(ctx context.Context, arg ListGroupAdminsByEmailAscFilteredParams) ([]User, error) {
	return New(DB).ListGroupAdminsByEmailAscFiltered(ctx, arg)
}

func ListGroupAdminsByEmailDescFiltered(ctx context.Context, arg ListGroupAdminsByEmailDescFilteredParams) ([]User, error) {
	return New(DB).ListGroupAdminsByEmailDescFiltered(ctx, arg)
}

func ListGroupPendingInvites(ctx context.Context, groupID sql.NullString) ([]MagicLink, error) {
	return New(DB).ListGroupPendingInvites(ctx, groupID)
}

func ListGroupPendingInvitesByCreatedAscFiltered(ctx context.Context, arg ListGroupPendingInvitesByCreatedAscFilteredParams) ([]MagicLink, error) {
	return New(DB).ListGroupPendingInvitesByCreatedAscFiltered(ctx, arg)
}

func ListGroupPendingInvitesByCreatedDescFiltered(ctx context.Context, arg ListGroupPendingInvitesByCreatedDescFilteredParams) ([]MagicLink, error) {
	return New(DB).ListGroupPendingInvitesByCreatedDescFiltered(ctx, arg)
}

func ListGroupPendingInvitesByEmailAscFiltered(ctx context.Context, arg ListGroupPendingInvitesByEmailAscFilteredParams) ([]MagicLink, error) {
	return New(DB).ListGroupPendingInvitesByEmailAscFiltered(ctx, arg)
}

func ListGroupPendingInvitesByEmailDescFiltered(ctx context.Context, arg ListGroupPendingInvitesByEmailDescFilteredParams) ([]MagicLink, error) {
	return New(DB).ListGroupPendingInvitesByEmailDescFiltered(ctx, arg)
}

func ListGroupReadersByEmailAscFiltered(ctx context.Context, arg ListGroupReadersByEmailAscFilteredParams) ([]User, error) {
	return New(DB).ListGroupReadersByEmailAscFiltered(ctx, arg)
}

func ListGroupReadersByEmailDescFiltered(ctx context.Context, arg ListGroupReadersByEmailDescFilteredParams) ([]User, error) {
	return New(DB).ListGroupReadersByEmailDescFiltered(ctx, arg)
}

func ListGroupUserAccess(ctx context.Context, groupID string) ([]ListGroupUserAccessRow, error) {
	return New(DB).ListGroupUserAccess(ctx, groupID)
}

func ListGroupsByAdmin(ctx context.Context, userID string) ([]Group, error) {
	return New(DB).ListGroupsByAdmin(ctx, userID)
}

func ListGroupsByCreatedAscFiltered(ctx context.Context, arg ListGroupsByCreatedAscFilteredParams) ([]Group, error) {
	return New(DB).ListGroupsByCreatedAscFiltered(ctx, arg)
}

func ListGroupsByCreatedDescFiltered(ctx context.Context, arg ListGroupsByCreatedDescFilteredParams) ([]Group, error) {
	return New(DB).ListGroupsByCreatedDescFiltered(ctx, arg)
}

func ListGroupsByNameAscFiltered(ctx context.Context, arg ListGroupsByNameAscFilteredParams) ([]Group, error) {
	return New(DB).ListGroupsByNameAscFiltered(ctx, arg)
}

func ListGroupsByNameDescFiltered(ctx context.Context, arg ListGroupsByNameDescFilteredParams) ([]Group, error) {
	return New(DB).ListGroupsByNameDescFiltered(ctx, arg)
}

func ListGroupsByReader(ctx context.Context, userID string) ([]Group, error) {
	return New(DB).ListGroupsByReader(ctx, userID)
}

func ListMembersByCreatedAtAscFiltered(ctx context.Context, arg ListMembersByCreatedAtAscFilteredParams) ([]Member, error) {
	return New(DB).ListMembersByCreatedAtAscFiltered(ctx, arg)
}

func ListMembersByCreatedAtDescFiltered(ctx context.Context, arg ListMembersByCreatedAtDescFilteredParams) ([]Member, error) {
	return New(DB).ListMembersByCreatedAtDescFiltered(ctx, arg)
}

func ListMembersByDescriptionAscFiltered(ctx context.Context, arg ListMembersByDescriptionAscFilteredParams) ([]Member, error) {
	return New(DB).ListMembersByDescriptionAscFiltered(ctx, arg)
}

func ListMembersByDescriptionDescFiltered(ctx context.Context, arg ListMembersByDescriptionDescFilteredParams) ([]Member, error) {
	return New(DB).ListMembersByDescriptionDescFiltered(ctx, arg)
}

func ListMembersByNameAscFiltered(ctx context.Context, arg ListMembersByNameAscFilteredParams) ([]Member, error) {
	return New(DB).ListMembersByNameAscFiltered(ctx, arg)
}

func ListMembersByNameDescFiltered(ctx context.Context, arg ListMembersByNameDescFilteredParams) ([]Member, error) {
	return New(DB).ListMembersByNameDescFiltered(ctx, arg)
}

func ListPaidEventsByGroup(ctx context.Context, groupID string) ([]Event, error) {
	return New(DB).ListPaidEventsByGroup(ctx, groupID)
}

func ListPaidExpensesByGroup(ctx context.Context, groupID string) ([]Expense, error) {
	return New(DB).ListPaidExpensesByGroup(ctx, groupID)
}

func ListPaidOutgoingParticipantsByGroup(ctx context.Context, groupID string) ([]ListPaidOutgoingParticipantsByGroupRow, error) {
	return New(DB).ListPaidOutgoingParticipantsByGroup(ctx, groupID)
}

func ListPaidOutgoingPaymentsByGroup(ctx context.Context, groupID string) ([]GroupOutgoingPayment, error) {
	return New(DB).ListPaidOutgoingPaymentsByGroup(ctx, groupID)
}

func ListParticipantsByEvent(ctx context.Context, arg ListParticipantsByEventParams) ([]ListParticipantsByEventRow, error) {
	return New(DB).ListParticipantsByEvent(ctx, arg)
}

func ListParticipantsByMember(ctx context.Context, arg ListParticipantsByMemberParams) ([]ListParticipantsByMemberRow, error) {
	return New(DB).ListParticipantsByMember(ctx, arg)
}

func ListParticipantsByMemberByAmountAscFiltered(ctx context.Context, arg ListParticipantsByMemberByAmountAscFilteredParams) ([]ListParticipantsByMemberByAmountAscFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByAmountAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByAmountDescFiltered(ctx context.Context, arg ListParticipantsByMemberByAmountDescFilteredParams) ([]ListParticipantsByMemberByAmountDescFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByAmountDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByCutAscFiltered(ctx context.Context, arg ListParticipantsByMemberByCutAscFilteredParams) ([]ListParticipantsByMemberByCutAscFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByCutAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByCutDescFiltered(ctx context.Context, arg ListParticipantsByMemberByCutDescFilteredParams) ([]ListParticipantsByMemberByCutDescFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByCutDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByExpenseAscFiltered(ctx context.Context, arg ListParticipantsByMemberByExpenseAscFilteredParams) ([]ListParticipantsByMemberByExpenseAscFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByExpenseAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByExpenseDescFiltered(ctx context.Context, arg ListParticipantsByMemberByExpenseDescFilteredParams) ([]ListParticipantsByMemberByExpenseDescFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByExpenseDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByPaidAscFiltered(ctx context.Context, arg ListParticipantsByMemberByPaidAscFilteredParams) ([]ListParticipantsByMemberByPaidAscFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByPaidAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByPaidAtAscFiltered(ctx context.Context, arg ListParticipantsByMemberByPaidAtAscFilteredParams) ([]ListParticipantsByMemberByPaidAtAscFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByPaidAtAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByPaidAtDescFiltered(ctx context.Context, arg ListParticipantsByMemberByPaidAtDescFilteredParams) ([]ListParticipantsByMemberByPaidAtDescFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByPaidAtDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByPaidDescFiltered(ctx context.Context, arg ListParticipantsByMemberByPaidDescFilteredParams) ([]ListParticipantsByMemberByPaidDescFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByPaidDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByTimeAscFiltered(ctx context.Context, arg ListParticipantsByMemberByTimeAscFilteredParams) ([]ListParticipantsByMemberByTimeAscFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByTimeAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByTimeDescFiltered(ctx context.Context, arg ListParticipantsByMemberByTimeDescFilteredParams) ([]ListParticipantsByMemberByTimeDescFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByTimeDescFiltered(ctx, arg)
}

func ListParticipantsByMemberByTitleAscFiltered(ctx context.Context, arg ListParticipantsByMemberByTitleAscFilteredParams) ([]ListParticipantsByMemberByTitleAscFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByTitleAscFiltered(ctx, arg)
}

func ListParticipantsByMemberByTitleDescFiltered(ctx context.Context, arg ListParticipantsByMemberByTitleDescFilteredParams) ([]ListParticipantsByMemberByTitleDescFilteredRow, error) {
	return New(DB).ListParticipantsByMemberByTitleDescFiltered(ctx, arg)
}

func ListRecentGroups(ctx context.Context, limit int64) ([]Group, error) {
	return New(DB).ListRecentGroups(ctx, limit)
}

func ListRecentPaidEventsByGroup(ctx context.Context, arg ListRecentPaidEventsByGroupParams) ([]Event, error) {
	return New(DB).ListRecentPaidEventsByGroup(ctx, arg)
}

func ListRecentPaidExpensesByGroup(ctx context.Context, arg ListRecentPaidExpensesByGroupParams) ([]Expense, error) {
	return New(DB).ListRecentPaidExpensesByGroup(ctx, arg)
}

func ListRecentPaidParticipantsByGroup(ctx context.Context, arg ListRecentPaidParticipantsByGroupParams) ([]ListRecentPaidParticipantsByGroupRow, error) {
	return New(DB).ListRecentPaidParticipantsByGroup(ctx, arg)
}

func ListRecentUsersWithBanStatus(ctx context.Context, limit int64) ([]ListRecentUsersWithBanStatusRow, error) {
	return New(DB).ListRecentUsersWithBanStatus(ctx, limit)
}

func ListSessionsByCreatedAscFiltered(ctx context.Context, arg ListSessionsByCreatedAscFilteredParams) ([]ListSessionsByCreatedAscFilteredRow, error) {
	return New(DB).ListSessionsByCreatedAscFiltered(ctx, arg)
}

func ListSessionsByCreatedDescFiltered(ctx context.Context, arg ListSessionsByCreatedDescFilteredParams) ([]ListSessionsByCreatedDescFilteredRow, error) {
	return New(DB).ListSessionsByCreatedDescFiltered(ctx, arg)
}

func ListSessionsByEmailAscFiltered(ctx context.Context, arg ListSessionsByEmailAscFilteredParams) ([]ListSessionsByEmailAscFilteredRow, error) {
	return New(DB).ListSessionsByEmailAscFiltered(ctx, arg)
}

func ListSessionsByEmailDescFiltered(ctx context.Context, arg ListSessionsByEmailDescFilteredParams) ([]ListSessionsByEmailDescFilteredRow, error) {
	return New(DB).ListSessionsByEmailDescFiltered(ctx, arg)
}

func ListUnpaidEventsByGroup(ctx context.Context, groupID string) ([]Event, error) {
	return New(DB).ListUnpaidEventsByGroup(ctx, groupID)
}

func ListUnpaidExpensesByGroup(ctx context.Context, groupID string) ([]Expense, error) {
	return New(DB).ListUnpaidExpensesByGroup(ctx, groupID)
}

func ListUnpaidOutgoingParticipantsByGroup(ctx context.Context, groupID string) ([]ListUnpaidOutgoingParticipantsByGroupRow, error) {
	return New(DB).ListUnpaidOutgoingParticipantsByGroup(ctx, groupID)
}

func ListUnpaidOutgoingPaymentsByGroup(ctx context.Context, groupID string) ([]GroupOutgoingPayment, error) {
	return New(DB).ListUnpaidOutgoingPaymentsByGroup(ctx, groupID)
}

func ListUserDetailCardStates(ctx context.Context, userID string) ([]ListUserDetailCardStatesRow, error) {
	return New(DB).ListUserDetailCardStates(ctx, userID)
}

func ListUserGroupsByAdminAscFiltered(ctx context.Context, arg ListUserGroupsByAdminAscFilteredParams) ([]ListUserGroupsByAdminAscFilteredRow, error) {
	return New(DB).ListUserGroupsByAdminAscFiltered(ctx, arg)
}

func ListUserGroupsByAdminDescFiltered(ctx context.Context, arg ListUserGroupsByAdminDescFilteredParams) ([]ListUserGroupsByAdminDescFilteredRow, error) {
	return New(DB).ListUserGroupsByAdminDescFiltered(ctx, arg)
}

func ListUserGroupsByCreatedAscFiltered(ctx context.Context, arg ListUserGroupsByCreatedAscFilteredParams) ([]ListUserGroupsByCreatedAscFilteredRow, error) {
	return New(DB).ListUserGroupsByCreatedAscFiltered(ctx, arg)
}

func ListUserGroupsByCreatedDescFiltered(ctx context.Context, arg ListUserGroupsByCreatedDescFilteredParams) ([]ListUserGroupsByCreatedDescFilteredRow, error) {
	return New(DB).ListUserGroupsByCreatedDescFiltered(ctx, arg)
}

func ListUserGroupsByNameAscFiltered(ctx context.Context, arg ListUserGroupsByNameAscFilteredParams) ([]ListUserGroupsByNameAscFilteredRow, error) {
	return New(DB).ListUserGroupsByNameAscFiltered(ctx, arg)
}

func ListUserGroupsByNameDescFiltered(ctx context.Context, arg ListUserGroupsByNameDescFilteredParams) ([]ListUserGroupsByNameDescFilteredRow, error) {
	return New(DB).ListUserGroupsByNameDescFiltered(ctx, arg)
}

func ListUserSessions(ctx context.Context, userID string) ([]UserSession, error) {
	return New(DB).ListUserSessions(ctx, userID)
}

func ListUsersByCreatedAscFiltered(ctx context.Context, arg ListUsersByCreatedAscFilteredParams) ([]ListUsersByCreatedAscFilteredRow, error) {
	return New(DB).ListUsersByCreatedAscFiltered(ctx, arg)
}

func ListUsersByCreatedDescFiltered(ctx context.Context, arg ListUsersByCreatedDescFilteredParams) ([]ListUsersByCreatedDescFilteredRow, error) {
	return New(DB).ListUsersByCreatedDescFiltered(ctx, arg)
}

func ListUsersByEmailAscFiltered(ctx context.Context, arg ListUsersByEmailAscFilteredParams) ([]ListUsersByEmailAscFilteredRow, error) {
	return New(DB).ListUsersByEmailAscFiltered(ctx, arg)
}

func ListUsersByEmailDescFiltered(ctx context.Context, arg ListUsersByEmailDescFilteredParams) ([]ListUsersByEmailDescFilteredRow, error) {
	return New(DB).ListUsersByEmailDescFiltered(ctx, arg)
}

func RemoveGroupAdmin(ctx context.Context, arg RemoveGroupAdminParams) error {
	return New(DB).RemoveGroupAdmin(ctx, arg)
}

func RemoveGroupReader(ctx context.Context, arg RemoveGroupReaderParams) error {
	return New(DB).RemoveGroupReader(ctx, arg)
}

func RemoveParticipant(ctx context.Context, arg RemoveParticipantParams) error {
	return New(DB).RemoveParticipant(ctx, arg)
}

func SumEventsFiltered(ctx context.Context, arg SumEventsFilteredParams) (int64, error) {
	return New(DB).SumEventsFiltered(ctx, arg)
}

func SumExpenseTotalsFiltered(ctx context.Context, arg SumExpenseTotalsFilteredParams) (SumExpenseTotalsFilteredRow, error) {
	return New(DB).SumExpenseTotalsFiltered(ctx, arg)
}

func SumExpensesFiltered(ctx context.Context, arg SumExpensesFilteredParams) (int64, error) {
	return New(DB).SumExpensesFiltered(ctx, arg)
}

func SumParticipantAmountsByGroup(ctx context.Context, groupID string) (int64, error) {
	return New(DB).SumParticipantAmountsByGroup(ctx, groupID)
}

func SumParticipantPaidAmountsByGroup(ctx context.Context, groupID string) (SumParticipantPaidAmountsByGroupRow, error) {
	return New(DB).SumParticipantPaidAmountsByGroup(ctx, groupID)
}

func SumParticipantTotalsByGroupFiltered(ctx context.Context, arg SumParticipantTotalsByGroupFilteredParams) (SumParticipantTotalsByGroupFilteredRow, error) {
	return New(DB).SumParticipantTotalsByGroupFiltered(ctx, arg)
}

func SumParticipantTotalsByMemberFiltered(ctx context.Context, arg SumParticipantTotalsByMemberFilteredParams) (SumParticipantTotalsByMemberFilteredRow, error) {
	return New(DB).SumParticipantTotalsByMemberFiltered(ctx, arg)
}

func UnbanUser(ctx context.Context, userID string) error {
	return New(DB).UnbanUser(ctx, userID)
}

func UpdateGroupAdmin(ctx context.Context, arg UpdateGroupAdminParams) error {
	return New(DB).UpdateGroupAdmin(ctx, arg)
}

func UpdateGroupName(ctx context.Context, arg UpdateGroupNameParams) (Group, error) {
	return New(DB).UpdateGroupName(ctx, arg)
}

func UpdateParticipant(ctx context.Context, arg UpdateParticipantParams) error {
	return New(DB).UpdateParticipant(ctx, arg)
}

func UpdateUserPreferredLang(ctx context.Context, arg UpdateUserPreferredLangParams) error {
	return New(DB).UpdateUserPreferredLang(ctx, arg)
}

func UpsertAppFlagBool(ctx context.Context, arg UpsertAppFlagBoolParams) error {
	return New(DB).UpsertAppFlagBool(ctx, arg)
}

func UpsertUserDetailCardState(ctx context.Context, arg UpsertUserDetailCardStateParams) error {
	return New(DB).UpsertUserDetailCardState(ctx, arg)
}

func UseMagicLink(ctx context.Context, id string) error {
	return New(DB).UseMagicLink(ctx, id)
}

func WithTx(tx *sql.Tx) *Queries {
	return New(DB).WithTx(tx)
}
