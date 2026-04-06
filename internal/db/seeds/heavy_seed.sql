PRAGMA foreign_keys = ON;

BEGIN;

INSERT OR IGNORE INTO users (id, email)
VALUES
  ('usr_SuperAdminSeed000001', 'admin@bandcash.localhost'),
  ('usr_AdminUserOneSeed0001', 'admin1@bandcash.local'),
  ('usr_AdminUserTwoSeed0002', 'admin2@bandcash.local'),
  ('usr_AdminUserTriSeed0003', 'admin3@bandcash.local'),
  ('usr_ViewerSeedUser00001A', 'viewer01@bandcash.local'),
  ('usr_ViewerSeedUser00002A', 'viewer02@bandcash.local'),
  ('usr_ViewerSeedUser00003A', 'viewer03@bandcash.local');

INSERT OR IGNORE INTO groups (id, name, admin_user_id)
VALUES
  ('grp_SeedDataLabGroup0001', 'Seed Data Lab', 'usr_SuperAdminSeed000001'),
  ('grp_SeedRoadCrewGroup002', 'Seed Road Crew', 'usr_AdminUserOneSeed0001'),
  ('grp_SeedAcousticGroup003', 'Seed Acoustic Duo', 'usr_AdminUserTwoSeed0002');

INSERT OR IGNORE INTO group_access (id, user_id, group_id, role)
VALUES
  ('gac_SeedViewerLink00001A', 'usr_ViewerSeedUser00001A', 'grp_SeedDataLabGroup0001', 'viewer'),
  ('gac_SeedViewerLink00002A', 'usr_ViewerSeedUser00002A', 'grp_SeedDataLabGroup0001', 'viewer'),
  ('gac_SeedViewerLink00003A', 'usr_ViewerSeedUser00003A', 'grp_SeedDataLabGroup0001', 'viewer'),
  ('gac_SeedViewerLink00004A', 'usr_SuperAdminSeed000001', 'grp_SeedAcousticGroup003', 'viewer');

INSERT OR IGNORE INTO group_access (id, user_id, group_id, role)
VALUES
  ('gac_SeedAdminLink00001A', 'usr_AdminUserOneSeed0001', 'grp_SeedDataLabGroup0001', 'admin'),
  ('gac_SeedAdminLink00002A', 'usr_AdminUserTwoSeed0002', 'grp_SeedDataLabGroup0001', 'admin'),
  ('gac_SeedAdminLink00003A', 'usr_AdminUserTriSeed0003', 'grp_SeedDataLabGroup0001', 'admin'),
  ('gac_SeedAdminLink00004A', 'usr_SuperAdminSeed000001', 'grp_SeedRoadCrewGroup002', 'admin');

INSERT OR IGNORE INTO magic_links (id, token, email, action, group_id, expires_at)
VALUES
  ('mag_SeedInviteLink00001A', 'tok_SeedPendingInvite001', 'pending01@bandcash.local', 'invite', 'grp_SeedDataLabGroup0001', datetime('now', '+2 days')),
  ('mag_SeedInviteLink00002A', 'tok_SeedPendingInvite002', 'pending02@bandcash.local', 'invite', 'grp_SeedDataLabGroup0001', datetime('now', '+2 days')),
  ('mag_SeedInviteLink00003A', 'tok_SeedPendingInvite003', 'pending03@bandcash.local', 'invite', 'grp_SeedDataLabGroup0001', datetime('now', '+2 days')),
  ('mag_SeedInviteLink00004A', 'tok_SeedPendingInvite004', 'pending04@bandcash.local', 'invite', 'grp_SeedDataLabGroup0001', datetime('now', '+2 days')),
  ('mag_SeedInviteLink00005A', 'tok_SeedPendingInvite005', 'pending05@bandcash.local', 'invite', 'grp_SeedDataLabGroup0001', datetime('now', '+2 days')),
  ('mag_SeedInviteLink00006A', 'tok_SeedPendingInvite006', 'pending06@bandcash.local', 'invite', 'grp_SeedDataLabGroup0001', datetime('now', '+2 days'));

INSERT OR REPLACE INTO app_flags (key, bool_value)
VALUES ('enable_signup', 1);

-- 10 members
WITH RECURSIVE seq(n) AS (
  SELECT 1
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 10
)
INSERT OR IGNORE INTO members (id, group_id, name, description)
SELECT
  printf('mem_%020d', n),
  'grp_SeedDataLabGroup0001',
  CASE (n % 8)
    WHEN 0 THEN 'Alex Harper'
    WHEN 1 THEN 'Jordan Vale'
    WHEN 2 THEN 'Casey Stone'
    WHEN 3 THEN 'Morgan Hale'
    WHEN 4 THEN 'Taylor Finch'
    WHEN 5 THEN 'Riley Knox'
    WHEN 6 THEN 'Avery Lane'
    ELSE 'Quinn North'
  END || ' #' || n,
  CASE
    WHEN n % 5 = 0 THEN ''
    WHEN n % 5 = 1 THEN 'vocals'
    WHEN n % 5 = 2 THEN 'guitar and keys'
    WHEN n % 5 = 3 THEN 'drums and percussion'
    ELSE 'bass and backing vocals'
  END
FROM seq;

-- 400 events
WITH RECURSIVE seq(n) AS (
  SELECT 1
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 400
)
INSERT OR IGNORE INTO events (id, group_id, title, time, place, description, amount, paid, paid_at)
SELECT
  printf('evt_%020d', n),
  'grp_SeedDataLabGroup0001',
  CASE (n % 10)
    WHEN 0 THEN 'Downtown Club Night'
    WHEN 1 THEN 'Festival Side Stage'
    WHEN 2 THEN 'Private Wedding Set'
    WHEN 3 THEN 'Riverside Acoustic Session'
    WHEN 4 THEN 'Corporate Launch Event'
    WHEN 5 THEN 'Late Night Jazz Room'
    WHEN 6 THEN 'Summer Street Concert'
    WHEN 7 THEN 'Studio Showcase'
    WHEN 8 THEN 'University Hall Gig'
    ELSE 'Weekend Headliner'
  END || ' #' || n,
  strftime('%Y-%m-%dT%H:%M', datetime('now', printf('-%d days', n % 365), printf('-%d hours', n % 24))),
  CASE (n % 7)
    WHEN 0 THEN 'A38 Ship'
    WHEN 1 THEN 'Durer Kert'
    WHEN 2 THEN 'Akvarium Klub'
    WHEN 3 THEN 'Kobuci Kert'
    WHEN 4 THEN 'Budapest Park'
    WHEN 5 THEN ''
    ELSE 'MOMKult'
  END,
  CASE
    WHEN n % 6 = 0 THEN ''
    WHEN n % 6 = 1 THEN 'ticketed event with local support acts'
    WHEN n % 6 = 2 THEN 'private booking with extended encore'
    WHEN n % 6 = 3 THEN 'festival slot with shared backline'
    WHEN n % 6 = 4 THEN 'corporate set with custom playlist'
    ELSE 'bar show with reduced setup'
  END,
  20000 + (ABS(RANDOM()) % 280000),
  CASE
    WHEN n % 4 = 0 THEN 1
    WHEN n % 4 = 1 THEN 1
    ELSE 0
  END,
  CASE
    WHEN n % 4 IN (0, 1) THEN strftime('%Y-%m-%dT%H:%M:%SZ', datetime('now', printf('-%d days', (n % 300) + 1), printf('-%d hours', n % 12)))
    ELSE NULL
  END
FROM seq;

-- 300 expenses
WITH RECURSIVE seq(n) AS (
  SELECT 1
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 300
)
INSERT OR IGNORE INTO expenses (id, group_id, title, description, amount, date)
SELECT
  printf('exp_%020d', n),
  'grp_SeedDataLabGroup0001',
  CASE (n % 9)
    WHEN 0 THEN 'Van rental'
    WHEN 1 THEN 'Fuel and parking'
    WHEN 2 THEN 'Sound engineer fee'
    WHEN 3 THEN 'Strings and drum heads'
    WHEN 4 THEN 'Promo materials'
    WHEN 5 THEN 'Backline rental'
    WHEN 6 THEN 'Venue consumables'
    WHEN 7 THEN 'Accommodation'
    ELSE 'Session catering'
  END || ' #' || n,
  CASE
    WHEN n % 4 = 0 THEN ''
    WHEN n % 4 = 1 THEN 'routine operating cost'
    WHEN n % 4 = 2 THEN 'one-off event expense'
    ELSE 'shared trip logistics'
  END,
  1500 + (ABS(RANDOM()) % 95000),
  date('now', printf('-%d days', n % 365))
FROM seq;

-- 1800 participant rows (dense enough for testing lists/joins)
WITH RECURSIVE seq(n) AS (
  SELECT 1
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 1800
)
INSERT OR IGNORE INTO participants (group_id, event_id, member_id, amount, expense)
SELECT
  'grp_SeedDataLabGroup0001',
  printf('evt_%020d', ((n - 1) % 400) + 1),
  printf('mem_%020d', ((n * 7 - 1) % 10) + 1),
  500 + (ABS(RANDOM()) % 25000),
  ABS(RANDOM()) % 2000
FROM seq;

COMMIT;
