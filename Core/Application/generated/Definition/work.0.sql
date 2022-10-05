CREATE TABLE system_info
(
    id          INTEGER PRIMARY KEY UNIQUE,
    description TEXT    NOT NULL,
    created     INTEGER NOT NULL);
CREATE TABLE users
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE projects
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    identifier  TEXT    NOT NULL UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    workflow_id TEXT    NOT NULL,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE ticket_types
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE ticket_statuses
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE tickets
(
    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_number    INTEGER NOT NULL,
    position         INTEGER NOT NULL,
    title            TEXT,
    description      TEXT,
    created          INTEGER NOT NULL,
    modified         INTEGER NOT NULL,
    ticket_type_id   TEXT    NOT NULL,
    ticket_status_id TEXT    NOT NULL,
    project_id       TEXT    NOT NULL,
    user_id          TEXT,
    estimation       REAL    NOT NULL,
    story_points     INTEGER NOT NULL,
    creator          TEXT    NOT NULL,
    deleted          BOOLEAN NOT NULL );
CREATE TABLE ticket_relationship_types
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE boards
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE workflows
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE assets
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    url         TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE labels
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE label_categories
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE repositories
(
    id                 TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository         TEXT    NOT NULL UNIQUE,
    description        TEXT,
    repository_type_id TEXT    NOT NULL,
    created            INTEGER NOT NULL,
    modified           INTEGER NOT NULL,
    deleted            BOOLEAN NOT NULL );
CREATE TABLE repository_types
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE components
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE organizations
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE teams
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE permissions
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE comments
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE permission_contexts
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE workflow_steps
(
    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title            TEXT    NOT NULL UNIQUE,
    description      TEXT,
    workflow_id      TEXT    NOT NULL,
    workflow_step_id TEXT,
    ticket_status_id TEXT    NOT NULL,
    created          INTEGER NOT NULL,
    modified         INTEGER NOT NULL,
    deleted          BOOLEAN NOT NULL );
CREATE INDEX workflow_steps_get_by_workflow_id_and_workflow_step_id_and_ticket_status_id ON workflow_steps
(workflow_id, workflow_step_id, ticket_status_id);
CREATE TABLE reports
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    title       TEXT,
    description TEXT,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE cycles
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    title       TEXT,
    description TEXT,
    cycle_id    TEXT    NOT NULL UNIQUE,
    type        INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE extensions
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    title         TEXT,
    description   TEXT,
    extension_key TEXT    NOT NULL UNIQUE,
    enabled       BOOLEAN NOT NULL ,
    deleted       BOOLEAN NOT NULL );
CREATE TABLE audit
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created   INTEGER NOT NULL,
    entity    TEXT,
    operation TEXT);
CREATE TABLE project_organization_mappings
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_id      TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL );
CREATE INDEX project_organization_mappings_get_by_project_id_and_organization_id ON
    project_organization_mappings (project_id, organization_id);
CREATE INDEX project_organization_mappings_get_by_created_and_modified ON
    project_organization_mappings (created, modified);
CREATE TABLE ticket_type_project_mappings
(
    id             TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_type_id TEXT    NOT NULL,
    project_id     TEXT    NOT NULL,
    created        INTEGER NOT NULL,
    modified       INTEGER NOT NULL,
    deleted        BOOLEAN NOT NULL );
CREATE INDEX ticket_type_project_mappings_get_by_ticket_type_id_and_project_id
    ON ticket_type_project_mappings (ticket_type_id, project_id);
CREATE INDEX ticket_type_project_mappings_get_by_created_and_modified
    ON ticket_type_project_mappings (created, modified);
CREATE TABLE audit_meta_data
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    audit_id TEXT    NOT NULL,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL);
CREATE TABLE reports_meta_data
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    report_id TEXT    NOT NULL,
    property  TEXT    NOT NULL,
    value     TEXT,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL);
CREATE TABLE boards_meta_data
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id TEXT    NOT NULL,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL);
CREATE TABLE tickets_meta_data
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    property  TEXT    NOT NULL,
    value     TEXT,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL );
CREATE TABLE ticket_relationships
(
    id                          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_relationship_type_id TEXT    NOT NULL,
    ticket_id                   TEXT    NOT NULL,
    child_ticket_id             TEXT    NOT NULL,
    created                     INTEGER NOT NULL,
    modified                    INTEGER NOT NULL,
    deleted                     BOOLEAN NOT NULL );
CREATE INDEX ticket_relationships_get_by_child_ticket_id_and_child_ticket_id
    ON ticket_relationships (ticket_id, child_ticket_id);
CREATE INDEX ticket_relationships_get_by_ticket_relationship_type_id
    ON ticket_relationships (ticket_relationship_type_id);
CREATE INDEX ticket_relationships_get_by_ticket_id_and_ticket_relationship_type_id
    ON ticket_relationships (ticket_id, ticket_relationship_type_id);
CREATE INDEX ticket_relationships_get_by_ticket_id_and_child_ticket_id_and_ticket_relationship_type_id
    ON ticket_relationships (ticket_id, child_ticket_id, ticket_relationship_type_id);
CREATE TABLE team_organization_mappings
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id         TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL );
CREATE TABLE team_project_mappings
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id    TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL );
CREATE TABLE repository_project_mappings
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository_id TEXT    NOT NULL,
    project_id    TEXT    NOT NULL,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL );
CREATE TABLE repository_commit_ticket_mappings
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository_id TEXT    NOT NULL,
    ticket_id     TEXT    NOT NULL,
    commit_hash   TEXT    NOT NULL UNIQUE,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL );
CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id
    ON repository_commit_ticket_mappings (repository_id);
CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id_and_ticket_id
    ON repository_commit_ticket_mappings (repository_id, ticket_id);
CREATE INDEX repository_commit_ticket_mappings_get_by_ticket_id_commit_hash
    ON repository_commit_ticket_mappings (ticket_id, commit_hash);
CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id_and_ticket_id_commit_hash
    ON repository_commit_ticket_mappings (repository_id, ticket_id, commit_hash);
CREATE INDEX repository_commit_ticket_mappings_get_by_created_and_modified
    ON repository_commit_ticket_mappings (created, modified);
CREATE TABLE component_ticket_mappings
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    component_id TEXT    NOT NULL,
    ticket_id    TEXT    NOT NULL,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL );
CREATE TABLE components_meta_data
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    component_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL );
CREATE INDEX components_meta_data_get_by_component_id_and_property_and_value
    ON components_meta_data (component_id, property, value);
CREATE TABLE asset_project_mappings
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id   TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL );
CREATE TABLE asset_team_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE asset_ticket_mappings
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id  TEXT    NOT NULL,
    ticket_id TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL );
CREATE TABLE asset_comment_mappings
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id   TEXT    NOT NULL,
    comment_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL );
CREATE TABLE label_label_category_mappings
(
    id                TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id          TEXT    NOT NULL,
    label_category_id TEXT    NOT NULL,
    created           INTEGER NOT NULL,
    modified          INTEGER NOT NULL,
    deleted           BOOLEAN NOT NULL );
CREATE INDEX label_label_category_mappings_get_by_label_category_id
    ON label_label_category_mappings (label_category_id);
CREATE INDEX label_label_category_mappings_get_by_created_and_modified
    ON label_label_category_mappings (created, modified);
CREATE TABLE label_project_mappings
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id   TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL );
CREATE TABLE label_team_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE label_ticket_mappings
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id  TEXT    NOT NULL,
    ticket_id TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL );
CREATE TABLE label_asset_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id TEXT    NOT NULL,
    asset_id TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE comment_ticket_mappings
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id TEXT    NOT NULL,
    ticket_id  TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL );
CREATE TABLE ticket_project_mappings
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL );
CREATE TABLE cycle_project_mappings
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    cycle_id   TEXT    NOT NULL UNIQUE,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL );
CREATE TABLE ticket_cycle_mappings
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    cycle_id  TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL );
CREATE TABLE ticket_board_mappings
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    board_id  TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL );
CREATE TABLE users_yandex_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL UNIQUE,
    username TEXT    NOT NULL UNIQUE,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE users_google_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL UNIQUE,
    username TEXT    NOT NULL UNIQUE,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE user_organization_mappings
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id         TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL );
CREATE TABLE user_team_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE permission_user_mappings
(
    id                    TEXT    NOT NULL PRIMARY KEY UNIQUE,
    permission_id         TEXT    NOT NULL,
    user_id               TEXT    NOT NULL,
    permission_context_id TEXT    NOT NULL,
    created               INTEGER NOT NULL,
    modified              INTEGER NOT NULL,
    deleted               BOOLEAN NOT NULL );
CREATE INDEX permission_user_mappings_get_by_user_id_and_permission_id
    ON permission_user_mappings (user_id, permission_id);
CREATE INDEX permission_user_mappings_get_by_user_id_and_permission_context_id
    ON permission_user_mappings (user_id, permission_context_id);
CREATE INDEX permission_user_mappings_get_by_permission_id_and_permission_context_id
    ON permission_user_mappings (permission_id, permission_context_id);
CREATE TABLE permission_team_mappings
(
    id                    TEXT    NOT NULL PRIMARY KEY UNIQUE,
    permission_id         TEXT    NOT NULL,
    team_id               TEXT    NOT NULL,
    permission_context_id TEXT    NOT NULL,
    created               INTEGER NOT NULL,
    modified              INTEGER NOT NULL,
    deleted               BOOLEAN NOT NULL );
CREATE INDEX permission_team_mappings_get_by_team_id_and_permission_id
    ON permission_team_mappings (team_id, permission_id);
CREATE INDEX permission_team_mappings_get_by_team_id_and_permission_context_id
    ON permission_team_mappings (team_id, permission_context_id);
CREATE TABLE configuration_data_extension_mappings
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    extension_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    enabled      BOOLEAN NOT NULL ,
    deleted      BOOLEAN NOT NULL );
CREATE INDEX configuration_data_extension_mappings_get_by_extension_id
    ON configuration_data_extension_mappings (extension_id);
CREATE INDEX configuration_data_extension_mappings_get_by_property_and_value
    ON configuration_data_extension_mappings (property, value);
CREATE INDEX configuration_data_extension_mappings_get_by_extension_id_and_property
    ON configuration_data_extension_mappings (extension_id, property);
CREATE INDEX configuration_data_extension_mappings_get_by_extension_id_and_property_and_value
    ON configuration_data_extension_mappings (extension_id, property, value);
CREATE INDEX configuration_data_extension_mappings_get_by_created_and_modified
    ON configuration_data_extension_mappings (created, modified);
CREATE TABLE extensions_meta_data
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    extension_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL );
CREATE INDEX extensions_meta_data_get_by_extension_id_and_property_and_value
    ON extensions_meta_data (extension_id, property, value);
