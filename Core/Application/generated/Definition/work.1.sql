CREATE TABLE system_info
(
    id          INTEGER PRIMARY KEY UNIQUE,
    description TEXT    NOT NULL,
    created     INTEGER NOT NULL);
CREATE TABLE project
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    identifier  TEXT    NOT NULL UNIQUE,
    title       TEXT    NOT NULL,
    description TEXT,
    workflow_id TEXT    NOT NULL,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE ticket_type
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE ticket_status
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE ticket
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
    deleted          BOOLEAN NOT NULL);
CREATE TABLE ticket_relationship_type
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE board
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE workflow
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE asset
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    url         TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE label
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE label_category
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE repository
(
    id                 TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository         TEXT    NOT NULL UNIQUE,
    description        TEXT,
    repository_type_id TEXT    NOT NULL,
    created            INTEGER NOT NULL,
    modified           INTEGER NOT NULL,
    deleted            BOOLEAN NOT NULL);
CREATE TABLE repository_type
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE component
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE organization
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE team
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE permission
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE comment
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL);
CREATE TABLE permission_context
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE workflow_step
(
    id               TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title            TEXT    NOT NULL UNIQUE,
    description      TEXT,
    workflow_id      TEXT    NOT NULL,
    workflow_step_id TEXT,
    ticket_status_id TEXT    NOT NULL,
    created          INTEGER NOT NULL,
    modified         INTEGER NOT NULL,
    deleted          BOOLEAN NOT NULL);
CREATE TABLE report
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    title       TEXT,
    description TEXT,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE cycle
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    title       TEXT,
    description TEXT,
    cycle_id    TEXT    NOT NULL UNIQUE,
    type        INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
CREATE TABLE extension
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    title         TEXT,
    description   TEXT,
    extension_key TEXT    NOT NULL UNIQUE,
    enabled       BOOLEAN NOT NULL,
    deleted       BOOLEAN NOT NULL);
CREATE TABLE audit
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created   INTEGER NOT NULL,
    entity    TEXT,
    operation TEXT);
CREATE TABLE project_organization_mapping
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    project_id      TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL);
CREATE INDEX project_organization_mappings_get_by_created_and_modified ON
    project_organization_mapping (created, modified);
CREATE TABLE ticket_type_project_mapping
(
    id             TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_type_id TEXT    NOT NULL,
    project_id     TEXT    NOT NULL,
    created        INTEGER NOT NULL,
    modified       INTEGER NOT NULL,
    deleted        BOOLEAN NOT NULL);
CREATE INDEX ticket_type_project_mappings_get_by_created_and_modified
    ON ticket_type_project_mapping (created, modified);
CREATE TABLE audit_meta_data
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    audit_id TEXT    NOT NULL,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL);
CREATE TABLE report_meta_data
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    report_id TEXT    NOT NULL,
    property  TEXT    NOT NULL,
    value     TEXT,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL);
CREATE TABLE board_meta_data
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    board_id TEXT    NOT NULL,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL);
CREATE TABLE ticket_meta_data
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    property  TEXT    NOT NULL,
    value     TEXT,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL);
CREATE TABLE ticket_relationship
(
    id                          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_relationship_type_id TEXT    NOT NULL,
    ticket_id                   TEXT    NOT NULL,
    child_ticket_id             TEXT    NOT NULL,
    created                     INTEGER NOT NULL,
    modified                    INTEGER NOT NULL,
    deleted                     BOOLEAN NOT NULL);
CREATE INDEX ticket_relationships_get_by_child_ticket_id_and_child_ticket_id
    ON ticket_relationship (ticket_id, child_ticket_id);
CREATE INDEX ticket_relationships_get_by_ticket_relationship_type_id
    ON ticket_relationship (ticket_relationship_type_id);
CREATE TABLE team_organization_mapping
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id         TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL);
CREATE TABLE team_project_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    team_id    TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL);
CREATE TABLE repository_project_mapping
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository_id TEXT    NOT NULL,
    project_id    TEXT    NOT NULL,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL);
CREATE TABLE repository_commit_ticket_mapping
(
    id            TEXT    NOT NULL PRIMARY KEY UNIQUE,
    repository_id TEXT    NOT NULL,
    ticket_id     TEXT    NOT NULL,
    commit_hash   TEXT    NOT NULL UNIQUE,
    created       INTEGER NOT NULL,
    modified      INTEGER NOT NULL,
    deleted       BOOLEAN NOT NULL);
CREATE INDEX repository_commit_ticket_mappings_get_by_repository_id
    ON repository_commit_ticket_mapping (repository_id);
CREATE INDEX repository_commit_ticket_mappings_get_by_ticket_id_commit_hash
    ON repository_commit_ticket_mapping (ticket_id, commit_hash);
CREATE INDEX repository_commit_ticket_mappings_get_by_created_and_modified
    ON repository_commit_ticket_mapping (created, modified);
CREATE TABLE component_ticket_mapping
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    component_id TEXT    NOT NULL,
    ticket_id    TEXT    NOT NULL,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL);
CREATE TABLE component_meta_data
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    component_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL);
CREATE INDEX components_meta_data_get_by_component_id_and_property_and_value
    ON component_meta_data (component_id, property, value);
CREATE TABLE asset_project_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id   TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL);
CREATE TABLE asset_team_mapping
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL);
CREATE TABLE asset_ticket_mapping
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id  TEXT    NOT NULL,
    ticket_id TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL);
CREATE TABLE asset_comment_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    asset_id   TEXT    NOT NULL,
    comment_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL);
CREATE TABLE label_label_category_mapping
(
    id                TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id          TEXT    NOT NULL,
    label_category_id TEXT    NOT NULL,
    created           INTEGER NOT NULL,
    modified          INTEGER NOT NULL,
    deleted           BOOLEAN NOT NULL);
CREATE INDEX label_label_category_mappings_get_by_label_category_id
    ON label_label_category_mapping (label_category_id);
CREATE INDEX label_label_category_mappings_get_by_created_and_modified
    ON label_label_category_mapping (created, modified);
CREATE TABLE label_project_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id   TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL);
CREATE TABLE label_team_mapping
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL);
CREATE TABLE label_ticket_mapping
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id  TEXT    NOT NULL,
    ticket_id TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL);
CREATE TABLE label_asset_mapping
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    label_id TEXT    NOT NULL,
    asset_id TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL);
CREATE TABLE comment_ticket_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    comment_id TEXT    NOT NULL,
    ticket_id  TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL);
CREATE TABLE ticket_project_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id  TEXT    NOT NULL,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL);
CREATE TABLE cycle_project_mapping
(
    id         TEXT    NOT NULL PRIMARY KEY UNIQUE,
    cycle_id   TEXT    NOT NULL UNIQUE,
    project_id TEXT    NOT NULL,
    created    INTEGER NOT NULL,
    modified   INTEGER NOT NULL,
    deleted    BOOLEAN NOT NULL);
CREATE TABLE ticket_cycle_mapping
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    cycle_id  TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL);
CREATE TABLE ticket_board_mapping
(
    id        TEXT    NOT NULL PRIMARY KEY UNIQUE,
    ticket_id TEXT    NOT NULL,
    board_id  TEXT    NOT NULL,
    created   INTEGER NOT NULL,
    modified  INTEGER NOT NULL,
    deleted   BOOLEAN NOT NULL);
CREATE TABLE user_default_mapping
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL UNIQUE,
    username TEXT    NOT NULL UNIQUE,
    secret   TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL);
CREATE TABLE user_organization_mapping
(
    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id         TEXT    NOT NULL,
    organization_id TEXT    NOT NULL,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL);
CREATE TABLE user_team_mapping
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    user_id  TEXT    NOT NULL,
    team_id  TEXT    NOT NULL,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL);
CREATE TABLE permission_user_mapping
(
    id                    TEXT    NOT NULL PRIMARY KEY UNIQUE,
    permission_id         TEXT    NOT NULL,
    user_id               TEXT    NOT NULL,
    permission_context_id TEXT    NOT NULL,
    created               INTEGER NOT NULL,
    modified              INTEGER NOT NULL,
    deleted               BOOLEAN NOT NULL);
CREATE INDEX permission_user_mappings_get_by_user_id_and_permission_id
    ON permission_user_mapping (user_id, permission_id);
CREATE TABLE permission_team_mapping
(
    id                    TEXT    NOT NULL PRIMARY KEY UNIQUE,
    permission_id         TEXT    NOT NULL,
    team_id               TEXT    NOT NULL,
    permission_context_id TEXT    NOT NULL,
    created               INTEGER NOT NULL,
    modified              INTEGER NOT NULL,
    deleted               BOOLEAN NOT NULL);
CREATE INDEX permission_team_mappings_get_by_team_id_and_permission_id
    ON permission_team_mapping (team_id, permission_id);
CREATE TABLE configuration_data_extension_mapping
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    extension_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    enabled      BOOLEAN NOT NULL,
    deleted      BOOLEAN NOT NULL);
CREATE INDEX configuration_data_extension_mappings_get_by_extension_id
    ON configuration_data_extension_mapping (extension_id);
CREATE INDEX configuration_data_extension_mappings_get_by_property_and_value
    ON configuration_data_extension_mapping (property, value);
CREATE TABLE extension_meta_data
(
    id           TEXT    NOT NULL PRIMARY KEY UNIQUE,
    extension_id TEXT    NOT NULL,
    property     TEXT    NOT NULL,
    value        TEXT,
    created      INTEGER NOT NULL,
    modified     INTEGER NOT NULL,
    deleted      BOOLEAN NOT NULL);
CREATE INDEX extensions_meta_data_get_by_extension_id_and_property_and_value
    ON extension_meta_data (extension_id, property, value);
