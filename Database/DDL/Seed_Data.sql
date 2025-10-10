/*
    Seed Data for HelixTrack Core

    This file contains initial/default data for the system:
    - System information
    - Default ticket types (Bug, Story, Task, Epic)
    - Default ticket statuses (Open, In Progress, Resolved, Closed)
    - Default workflow
*/

-- System Information
INSERT OR IGNORE INTO system_info (id, description, created)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'HelixTrack Core System - Initial Setup',
    strftime('%s', 'now')
);

-- Default Workflow
INSERT OR IGNORE INTO workflow (id, title, description, created, modified, deleted)
VALUES (
    'default-workflow',
    'Default Workflow',
    'Standard workflow for all projects',
    strftime('%s', 'now'),
    strftime('%s', 'now'),
    0
);

-- Default Ticket Types
INSERT OR IGNORE INTO ticket_type (id, title, description, created, modified, deleted)
VALUES
    ('type-bug', 'Bug', 'A problem which impairs or prevents function', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('type-story', 'Story', 'User story or feature request', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('type-task', 'Task', 'A task that needs to be done', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('type-epic', 'Epic', 'A large feature or initiative', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('default-type', 'Default', 'Default ticket type', strftime('%s', 'now'), strftime('%s', 'now'), 0);

-- Default Ticket Statuses
INSERT OR IGNORE INTO ticket_status (id, title, description, created, modified, deleted)
VALUES
    ('status-open', 'Open', 'Issue is open and ready to be worked on', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('status-in-progress', 'In Progress', 'Work is currently being done', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('status-resolved', 'Resolved', 'Issue has been resolved', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('status-closed', 'Closed', 'Issue is closed and verified', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('status-reopened', 'Reopened', 'Issue was closed but has been reopened', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('open', 'Open', 'Default open status', strftime('%s', 'now'), strftime('%s', 'now'), 0);

-- Default Workflow Steps
INSERT OR IGNORE INTO workflow_step (id, workflow_id, title, description, step_order, created, modified, deleted)
VALUES
    ('step-1', 'default-workflow', 'Open', 'Initial state', 1, strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('step-2', 'default-workflow', 'In Progress', 'Work in progress', 2, strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('step-3', 'default-workflow', 'Resolved', 'Work completed', 3, strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('step-4', 'default-workflow', 'Closed', 'Final state', 4, strftime('%s', 'now'), strftime('%s', 'now'), 0);

-- Default Ticket Relationship Types
INSERT OR IGNORE INTO ticket_relationship_type (id, title, description, created, modified, deleted)
VALUES
    ('rel-blocks', 'Blocks', 'This ticket blocks another', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('rel-is-blocked-by', 'Is Blocked By', 'This ticket is blocked by another', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('rel-relates-to', 'Relates To', 'This ticket relates to another', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('rel-duplicates', 'Duplicates', 'This ticket is a duplicate of another', strftime('%s', 'now'), strftime('%s', 'now'), 0),
    ('rel-is-duplicated-by', 'Is Duplicated By', 'This ticket is duplicated by another', strftime('%s', 'now'), strftime('%s', 'now'), 0);
