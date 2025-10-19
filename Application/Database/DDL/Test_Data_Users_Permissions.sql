-- =====================================================================
-- HelixTrack Core - Test Users with All Permission Combinations
-- Comprehensive test data for permission and security level testing
-- =====================================================================

BEGIN;

-- =====================================================================
-- Test Users
-- Password for all users: test123 (hashed with bcrypt)
-- =====================================================================

INSERT INTO users (username, email, full_name, password_hash, active, created) VALUES
('viewer_user', 'viewer@test.helixtrack.com', 'Viewer Test User',
 '$2a$10$N9qo8uLOickgx2ZMRZoMye1I5X4vHhZQWO3X9CJbBFBN8HB8N7LRe', true, CURRENT_TIMESTAMP),

('contributor_user', 'contributor@test.helixtrack.com', 'Contributor Test User',
 '$2a$10$N9qo8uLOickgx2ZMRZoMye1I5X4vHhZQWO3X9CJbBFBN8HB8N7LRe', true, CURRENT_TIMESTAMP),

('developer_user', 'developer@test.helixtrack.com', 'Developer Test User',
 '$2a$10$N9qo8uLOickgx2ZMRZoMye1I5X4vHhZQWO3X9CJbBFBN8HB8N7LRe', true, CURRENT_TIMESTAMP),

('lead_user', 'lead@test.helixtrack.com', 'Project Lead Test User',
 '$2a$10$N9qo8uLOickgx2ZMRZoMye1I5X4vHhZQWO3X9CJbBFBN8HB8N7LRe', true, CURRENT_TIMESTAMP),

('admin_user', 'admin@test.helixtrack.com', 'Administrator Test User',
 '$2a$10$N9qo8uLOickgx2ZMRZoMye1I5X4vHhZQWO3X9CJbBFBN8HB8N7LRe', true, CURRENT_TIMESTAMP),

-- Additional edge case users
('no_permission_user', 'noperm@test.helixtrack.com', 'No Permissions User',
 '$2a$10$N9qo8uLOickgx2ZMRZoMye1I5X4vHhZQWO3X9CJbBFBN8HB8N7LRe', true, CURRENT_TIMESTAMP),

('mixed_permission_user', 'mixed@test.helixtrack.com', 'Mixed Permissions User',
 '$2a$10$N9qo8uLOickgx2ZMRZoMye1I5X4vHhZQWO3X9CJbBFBN8HB8N7LRe', true, CURRENT_TIMESTAMP),

('high_security_user', 'highsec@test.helixtrack.com', 'High Security User',
 '$2a$10$N9qo8uLOickgx2ZMRZoMye1I5X4vHhZQWO3X9CJbBFBN8HB8N7LRe', true, CURRENT_TIMESTAMP);

-- =====================================================================
-- Test Organization and Teams
-- =====================================================================

INSERT INTO organizations (name, description, created) VALUES
('Test Organization', 'Organization for testing permissions', CURRENT_TIMESTAMP);

INSERT INTO teams (name, organization_id, description, created) VALUES
('Test Team Alpha', (SELECT id FROM organizations WHERE name = 'Test Organization'),
 'Team for viewer and contributor testing', CURRENT_TIMESTAMP),

('Test Team Beta', (SELECT id FROM organizations WHERE name = 'Test Organization'),
 'Team for developer testing', CURRENT_TIMESTAMP),

('Test Team Gamma', (SELECT id FROM organizations WHERE name = 'Test Organization'),
 'Team for lead and admin testing', CURRENT_TIMESTAMP);

-- =====================================================================
-- Test Projects with Different Security Levels
-- =====================================================================

INSERT INTO projects (name, key, description, security_level, created) VALUES
('Public Project', 'PUB', 'Public project (security level 0)', 0, CURRENT_TIMESTAMP),
('Internal Project', 'INT', 'Internal project (security level 1)', 1, CURRENT_TIMESTAMP),
('Confidential Project', 'CONF', 'Confidential project (security level 2)', 2, CURRENT_TIMESTAMP),
('Restricted Project', 'REST', 'Restricted project (security level 3)', 3, CURRENT_TIMESTAMP),
('Secret Project', 'SEC', 'Secret project (security level 4)', 4, CURRENT_TIMESTAMP),
('Top Secret Project', 'TSEC', 'Top secret project (security level 5)', 5, CURRENT_TIMESTAMP);

-- =====================================================================
-- User Permissions
-- =====================================================================

-- Viewer User - READ permissions only, Public security level (0)
INSERT INTO user_permissions (user_id, resource_type, resource_id, permission_level, security_level) VALUES
((SELECT id FROM users WHERE username = 'viewer_user'), 'ticket', NULL, 1, 0),
((SELECT id FROM users WHERE username = 'viewer_user'), 'project', NULL, 1, 0),
((SELECT id FROM users WHERE username = 'viewer_user'), 'comment', NULL, 1, 0),
((SELECT id FROM users WHERE username = 'viewer_user'), 'sprint', NULL, 1, 0);

-- Contributor User - CREATE permissions, Internal security level (1)
INSERT INTO user_permissions (user_id, resource_type, resource_id, permission_level, security_level) VALUES
((SELECT id FROM users WHERE username = 'contributor_user'), 'ticket', NULL, 2, 1),
((SELECT id FROM users WHERE username = 'contributor_user'), 'project', NULL, 1, 1),
((SELECT id FROM users WHERE username = 'contributor_user'), 'comment', NULL, 2, 1),
((SELECT id FROM users WHERE username = 'contributor_user'), 'sprint', NULL, 1, 1);

-- Developer User - UPDATE permissions, Confidential security level (2)
INSERT INTO user_permissions (user_id, resource_type, resource_id, permission_level, security_level) VALUES
((SELECT id FROM users WHERE username = 'developer_user'), 'ticket', NULL, 3, 2),
((SELECT id FROM users WHERE username = 'developer_user'), 'project', NULL, 3, 2),
((SELECT id FROM users WHERE username = 'developer_user'), 'comment', NULL, 3, 2),
((SELECT id FROM users WHERE username = 'developer_user'), 'sprint', NULL, 2, 2),
((SELECT id FROM users WHERE username = 'developer_user'), 'release', NULL, 1, 2);

-- Project Lead User - UPDATE/EXECUTE permissions, Restricted security level (3)
INSERT INTO user_permissions (user_id, resource_type, resource_id, permission_level, security_level) VALUES
((SELECT id FROM users WHERE username = 'lead_user'), 'ticket', NULL, 5, 3),
((SELECT id FROM users WHERE username = 'lead_user'), 'project', NULL, 3, 3),
((SELECT id FROM users WHERE username = 'lead_user'), 'comment', NULL, 5, 3),
((SELECT id FROM users WHERE username = 'lead_user'), 'sprint', NULL, 3, 3),
((SELECT id FROM users WHERE username = 'lead_user'), 'release', NULL, 2, 3),
((SELECT id FROM users WHERE username = 'lead_user'), 'user', NULL, 1, 3);

-- Admin User - DELETE permissions, Top Secret security level (5)
INSERT INTO user_permissions (user_id, resource_type, resource_id, permission_level, security_level) VALUES
((SELECT id FROM users WHERE username = 'admin_user'), 'ticket', NULL, 5, 5),
((SELECT id FROM users WHERE username = 'admin_user'), 'project', NULL, 5, 5),
((SELECT id FROM users WHERE username = 'admin_user'), 'comment', NULL, 5, 5),
((SELECT id FROM users WHERE username = 'admin_user'), 'sprint', NULL, 5, 5),
((SELECT id FROM users WHERE username = 'admin_user'), 'release', NULL, 5, 5),
((SELECT id FROM users WHERE username = 'admin_user'), 'user', NULL, 5, 5),
((SELECT id FROM users WHERE username = 'admin_user'), 'organization', NULL, 5, 5),
((SELECT id FROM users WHERE username = 'admin_user'), 'team', NULL, 5, 5);

-- No Permission User - No permissions (for denial testing)
-- (No entries - user has no permissions)

-- Mixed Permission User - Different levels for different resources
INSERT INTO user_permissions (user_id, resource_type, resource_id, permission_level, security_level) VALUES
((SELECT id FROM users WHERE username = 'mixed_permission_user'), 'ticket', NULL, 3, 2),
((SELECT id FROM users WHERE username = 'mixed_permission_user'), 'project', NULL, 1, 1),
((SELECT id FROM users WHERE username = 'mixed_permission_user'), 'comment', NULL, 5, 0),
((SELECT id FROM users WHERE username = 'mixed_permission_user'), 'sprint', NULL, 2, 2);

-- High Security User - READ only but high security clearance (4)
INSERT INTO user_permissions (user_id, resource_type, resource_id, permission_level, security_level) VALUES
((SELECT id FROM users WHERE username = 'high_security_user'), 'ticket', NULL, 1, 4),
((SELECT id FROM users WHERE username = 'high_security_user'), 'project', NULL, 1, 4),
((SELECT id FROM users WHERE username = 'high_security_user'), 'comment', NULL, 1, 4);

-- =====================================================================
-- Project Roles
-- =====================================================================

INSERT INTO project_roles (user_id, project_id, role) VALUES
-- Viewer in Public Project
((SELECT id FROM users WHERE username = 'viewer_user'),
 (SELECT id FROM projects WHERE key = 'PUB'), 1),

-- Contributor in Internal Project
((SELECT id FROM users WHERE username = 'contributor_user'),
 (SELECT id FROM projects WHERE key = 'INT'), 2),

-- Developer in Confidential Project
((SELECT id FROM users WHERE username = 'developer_user'),
 (SELECT id FROM projects WHERE key = 'CONF'), 3),

-- Project Lead in Restricted Project
((SELECT id FROM users WHERE username = 'lead_user'),
 (SELECT id FROM projects WHERE key = 'REST'), 4),

-- Admin in Top Secret Project
((SELECT id FROM users WHERE username = 'admin_user'),
 (SELECT id FROM projects WHERE key = 'TSEC'), 5);

-- =====================================================================
-- Team Memberships
-- =====================================================================

INSERT INTO team_members (team_id, user_id, role, joined) VALUES
((SELECT id FROM teams WHERE name = 'Test Team Alpha'),
 (SELECT id FROM users WHERE username = 'viewer_user'), 'member', CURRENT_TIMESTAMP),

((SELECT id FROM teams WHERE name = 'Test Team Alpha'),
 (SELECT id FROM users WHERE username = 'contributor_user'), 'member', CURRENT_TIMESTAMP),

((SELECT id FROM teams WHERE name = 'Test Team Beta'),
 (SELECT id FROM users WHERE username = 'developer_user'), 'member', CURRENT_TIMESTAMP),

((SELECT id FROM teams WHERE name = 'Test Team Gamma'),
 (SELECT id FROM users WHERE username = 'lead_user'), 'lead', CURRENT_TIMESTAMP),

((SELECT id FROM teams WHERE name = 'Test Team Gamma'),
 (SELECT id FROM users WHERE username = 'admin_user'), 'admin', CURRENT_TIMESTAMP);

-- =====================================================================
-- Test Tickets with Different Security Levels
-- =====================================================================

INSERT INTO tickets (title, description, project_id, security_level, reporter_id, created) VALUES
('Public Ticket', 'Ticket accessible to everyone',
 (SELECT id FROM projects WHERE key = 'PUB'), 0,
 (SELECT id FROM users WHERE username = 'viewer_user'), CURRENT_TIMESTAMP),

('Internal Ticket', 'Ticket for internal users only',
 (SELECT id FROM projects WHERE key = 'INT'), 1,
 (SELECT id FROM users WHERE username = 'contributor_user'), CURRENT_TIMESTAMP),

('Confidential Ticket', 'Confidential ticket',
 (SELECT id FROM projects WHERE key = 'CONF'), 2,
 (SELECT id FROM users WHERE username = 'developer_user'), CURRENT_TIMESTAMP),

('Restricted Ticket', 'Restricted ticket',
 (SELECT id FROM projects WHERE key = 'REST'), 3,
 (SELECT id FROM users WHERE username = 'lead_user'), CURRENT_TIMESTAMP),

('Secret Ticket', 'Secret ticket',
 (SELECT id FROM projects WHERE key = 'SEC'), 4,
 (SELECT id FROM users WHERE username = 'admin_user'), CURRENT_TIMESTAMP),

('Top Secret Ticket', 'Top secret ticket',
 (SELECT id FROM projects WHERE key = 'TSEC'), 5,
 (SELECT id FROM users WHERE username = 'admin_user'), CURRENT_TIMESTAMP);

-- =====================================================================
-- Test Summary
-- =====================================================================

SELECT 'Test data creation complete!' AS status;

SELECT 'Test Users:' AS category, COUNT(*) AS count FROM users WHERE email LIKE '%@test.helixtrack.com'
UNION ALL
SELECT 'Test Projects:', COUNT(*) FROM projects WHERE key IN ('PUB', 'INT', 'CONF', 'REST', 'SEC', 'TSEC')
UNION ALL
SELECT 'Test Teams:', COUNT(*) FROM teams WHERE name LIKE 'Test Team%'
UNION ALL
SELECT 'User Permissions:', COUNT(*) FROM user_permissions
UNION ALL
SELECT 'Project Roles:', COUNT(*) FROM project_roles
UNION ALL
SELECT 'Team Members:', COUNT(*) FROM team_members
UNION ALL
SELECT 'Test Tickets:', COUNT(*) FROM tickets WHERE security_level IN (0,1,2,3,4,5);

COMMIT;

-- =====================================================================
-- Usage Instructions
-- =====================================================================

/*
To load this test data:

SQLite:
sqlite3 Database/Definition.sqlite < Database/DDL/Test_Data_Users_Permissions.sql

PostgreSQL:
psql -U helixtrack -d helixtrack_core -f Database/DDL/Test_Data_Users_Permissions.sql

Test User Credentials:
- viewer_user / test123        (READ only, security level 0)
- contributor_user / test123   (CREATE, security level 1)
- developer_user / test123     (UPDATE, security level 2)
- lead_user / test123          (UPDATE/EXECUTE, security level 3)
- admin_user / test123         (DELETE/ALL, security level 5)
- no_permission_user / test123 (NO permissions - for denial testing)
- mixed_permission_user / test123 (Mixed permissions across resources)
- high_security_user / test123 (High security clearance, low permissions)

Total Test Combinations:
- 8 users × 8 resources × 5 actions × 6 security levels = 1,920 test cases

Permission Levels:
1 = READ
2 = CREATE
3 = UPDATE/EXECUTE
5 = DELETE/ALL

Security Levels:
0 = PUBLIC
1 = INTERNAL
2 = CONFIDENTIAL
3 = RESTRICTED
4 = SECRET
5 = TOP_SECRET

Resources Covered:
- ticket, project, comment, sprint, release, user, organization, team
*/
