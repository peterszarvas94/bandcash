DROP VIEW IF EXISTS group_outgoing_payments;

DROP TRIGGER IF EXISTS trg_groups_owner_access_update;
DROP TRIGGER IF EXISTS trg_groups_owner_access_insert;
DROP TRIGGER IF EXISTS trg_expenses_updated_at;
DROP TRIGGER IF EXISTS trg_participants_updated_at;
DROP TRIGGER IF EXISTS trg_members_updated_at;
DROP TRIGGER IF EXISTS trg_events_updated_at;

DROP INDEX IF EXISTS idx_group_access_owner_per_group;
DROP INDEX IF EXISTS idx_group_access_role;
DROP INDEX IF EXISTS idx_group_access_group_id;
DROP INDEX IF EXISTS idx_group_access_user_id;

DROP INDEX IF EXISTS idx_user_sessions_expires;
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_token;

DROP INDEX IF EXISTS idx_user_detail_card_states_user_id;
DROP INDEX IF EXISTS idx_expenses_group_id;
DROP INDEX IF EXISTS idx_banned_users_user_id;
DROP INDEX IF EXISTS idx_participants_member_id;
DROP INDEX IF EXISTS idx_participants_event_id;
DROP INDEX IF EXISTS idx_participants_group_id;
DROP INDEX IF EXISTS idx_members_group_id;
DROP INDEX IF EXISTS idx_events_group_id;
DROP INDEX IF EXISTS idx_magic_links_email;
DROP INDEX IF EXISTS idx_magic_links_token;

DROP TABLE IF EXISTS group_access;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS user_detail_card_states;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS banned_users;
DROP TABLE IF EXISTS app_flags;
DROP TABLE IF EXISTS participants;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS magic_links;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;
