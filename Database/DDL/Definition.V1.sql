/*
    Version: 1
*/

/*
    Notes:

    - Identifiers in the system are UUID strings (VARCHAR with the size of 36).
    - Mapping tables are used for binding entities and defining relationships.
        Mapping tables are used as well to append properties to the entities.
*/

/*
    TODOs:

    - TODO: Descriptions
    - TODO: Indexes

      Features:

      Cluster I:

    - TODO: Feature to define -> Reports (Time tracking reports, current status(es), etc.)

      Cluster II:

    - TODO: Feature to define -> Backlog and boards
    - TODO: Feature to define -> Sprint, Milestone, Release (version associated)
    - TODO: Feature to define -> Scrum vs Kanban
        - Planning
        - Grooming
        - Retrospective
    - TODO: Feature to define -> Roadmaps
    - TODO: Feature to define -> Workflows

      Cluster III:

    - TODO: Feature to define -> Connecting with other systems (products) (Integrations)
        - Repository connection and intelligence
        - Connecting with other products
            - Connect chats with entities (for example) (addition to the comments feature)

      Other:

    - TODO: API generator - generate models and API from the definition SQL.
        SQL -> OpenApi definition -> Generated code
*/

/*
    Cleaning up:
        TODO: Comment out before the production.
            Shell scripts should execute each SQl with TO-DO check. If TO-DO found -> fail.
*/
DROP TABLE IF EXISTS system_info;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS ticket_types;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS ticket_relationship_types;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS label_categories;
DROP TABLE IF EXISTS repositories;
DROP TABLE IF EXISTS components;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS permission_contexts;
DROP TABLE IF EXISTS audit;
DROP TABLE IF EXISTS times;
DROP TABLE IF EXISTS users_yandex_mappings;
DROP TABLE IF EXISTS users_google_mappings;
DROP TABLE IF EXISTS user_organization_mappings;
DROP TABLE IF EXISTS user_team_mappings;
DROP TABLE IF EXISTS permission_user_mappings;
DROP TABLE IF EXISTS project_organization_mappings;
DROP TABLE IF EXISTS ticket_relationships;
DROP TABLE IF EXISTS ticket_type_project_mappings;
DROP TABLE IF EXISTS audit_meta_data;
DROP TABLE IF EXISTS tickets_meta_data;
DROP TABLE IF EXISTS team_organization_mappings;
DROP TABLE IF EXISTS team_project_mappings;
DROP TABLE IF EXISTS repository_project_mappings;
DROP TABLE IF EXISTS component_ticket_mappings;
DROP TABLE IF EXISTS time_ticket_mappings;
DROP TABLE IF EXISTS components_meta_data;
DROP TABLE IF EXISTS asset_project_mappings;
DROP TABLE IF EXISTS asset_team_mappings;
DROP TABLE IF EXISTS asset_ticket_mappings;
DROP TABLE IF EXISTS asset_comment_mappings;
DROP TABLE IF EXISTS label_label_category_mappings;
DROP TABLE IF EXISTS label_project_mappings;
DROP TABLE IF EXISTS label_team_mappings;
DROP TABLE IF EXISTS label_ticket_mappings;
DROP TABLE IF EXISTS label_asset_mappings;
DROP TABLE IF EXISTS comment_ticket_mappings;
DROP TABLE IF EXISTS permission_team_mappings;

/*
  Identifies the version of the database (system).
  After each SQL script execution the version will be increased and execution description provided.
  TODO: To be connected to shell script runners
 */
CREATE TABLE system_info
(

    id          INTEGER PRIMARY KEY UNIQUE,
    description VARCHAR NOT NULL UNIQUE,
    created     INTEGER NOT NULL
);

/*
    The system entities:
 */

/*
     System's users.
     User is identified by the unique identifier (id).
     Since there may be different types of users, different kinds of data
     can be mapped (associated) with the user ID.
     For that purpose there are other mappings to the user ID such as Yandex OAuth2 mappings for example.
 */
CREATE TABLE users
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL,
    deleted  BOOLEAN     NOT NULL CHECK (deleted IN (0, 1))
);

/*
    The basic project definition.
 */
CREATE TABLE projects
(

    id          VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title       VARCHAR     NOT NULL,
    description VARCHAR     NOT NULL,
    created     INTEGER     NOT NULL,
    modified    INTEGER     NOT NULL
);

/*
    Ticket type definitions.
 */
CREATE TABLE ticket_types
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Tickets.
    Tickets belong to the project.
    Each ticket has its ticket type anf children if supported.
 */
CREATE TABLE tickets
(

    id             VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title          VARCHAR,
    description    VARCHAR,
    created        INTEGER     NOT NULL,
    modified       INTEGER     NOT NULL,
    ticket_type_id VARCHAR(36) NOT NULL,
    project_id     VARCHAR(36) NOT NULL
);

/*
    Ticket relationship types.
    For example:
        - Blocked by
        - Blocks
        - Cloned by
        - Clones, etc.
 */
CREATE TABLE ticket_relationship_types
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Images, attachments, etc.
    Defined by the identifier and the resource url.
 */
CREATE TABLE assets
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    url      VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Labels.
    Label can be associated with the almost everything:
        - Project
        - Team
        - Ticket
        - Asset
 */
CREATE TABLE labels
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Labels can be divided into categories (which is optional).
    TODO: Populate from the initialization JSON by the proper shell script.
 */
CREATE TABLE label_categories
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
      The code repositories - Identified by the identifier and the repository URL.
      Default repository type is Git repository.
 */
CREATE TABLE repositories
(

    id         VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    repository VARCHAR     NOT NULL UNIQUE,

    type       VARCHAR CHECK ( type IN
                               ('Git', 'CVS', 'SVN', 'Mercurial',
                                'Perforce', 'Monotone', 'Bazaar',
                                'TFS', 'VSTS', 'IBM Rational ClearCase',
                                'Revision Control System', 'VSS',
                                'CA Harvest Software Change Manager',
                                'PVCS', 'darcs')
        )                  NOT NULL DEFAULT 'Git',

    created    INTEGER     NOT NULL,
    modified   INTEGER     NOT NULL
);

/*
    Components.
    Components are associated with the tickets.
    For example:
        - Backend
        - Android Client
        - Core Engine
        - Webapp, etc.
 */
CREATE TABLE components
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    The organization definition. Organization is the owner of the project.
 */
CREATE TABLE organizations
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    The team definition. Organization is the owner of the team.
 */
CREATE TABLE teams
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Permission definitions.
    Permissions are (for example):

        CREATE
        UPDATE
        DELETE
        etc.
 */
CREATE TABLE permissions
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);


/*
    Comments.
    Users can comment on:
        - Tickets
        - Tbd.
 */
CREATE TABLE comments
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    comment  VARCHAR     NOT NULL,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Permission contexts.
    Each permission must assigned to the permission owner must have a valid context.
    Permission contexts are (for example):

        organization.project
        organization.team
 */
CREATE TABLE permission_contexts
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title    VARCHAR     NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Audit trail.
 */
CREATE TABLE times
(

    id          VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    timestamp   INTEGER     NOT NULL,
    how_mush    INTEGER     NOT NULL,

    unit        VARCHAR CHECK (
            unit IN (
                     'Minute', 'Hour', 'Day', 'Week', 'Month'
            )
        )                   NOT NULL DEFAULT 'Hour',

    title       VARCHAR,
    description VARCHAR
);

/*
    Time tracking.
    Time is tracked against the tickets.
    One entry is associated with the parent ticket and it contains the information:
        - How much time
        - Unit (time unit)
        - The title for the performed work (optional)
        - The description for the performed work (optional)
 */
CREATE TABLE audit
(

    id        VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    created   INTEGER     NOT NULL,
    entity    VARCHAR,
    operation VARCHAR
);

/*
    Mappings:
 */

/*
    Project belongs to the organization. Multiple projects can belong to one organization.
 */
CREATE TABLE project_organization_mappings
(

    id              VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    organization_id VARCHAR(36) NOT NULL,
    project_id      VARCHAR(36) NOT NULL,
    created         INTEGER     NOT NULL,
    modified        INTEGER     NOT NULL,
    UNIQUE (organization_id, project_id) ON CONFLICT ABORT
);

/*
    Each project has the ticket types that it supports.
 */
CREATE TABLE ticket_type_project_mappings
(

    id             VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    ticket_type_id VARCHAR(36) NOT NULL,
    project_id     VARCHAR(36) NOT NULL,
    created        INTEGER     NOT NULL,
    modified       INTEGER     NOT NULL,
    UNIQUE (ticket_type_id, project_id) ON CONFLICT ABORT
);

/*
    Audit trail meta-data.

    TODO: All meta data to be instantiated in SQL from the JSON definition and processed by shell script.
 */
CREATE TABLE audit_meta_data
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    audit_id VARCHAR(36) NOT NULL,
    property VARCHAR     NOT NULL,
    value    VARCHAR,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Tickets meta-data.
 */
CREATE TABLE tickets_meta_data
(

    id        VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    ticket_id VARCHAR(36) NOT NULL,
    property  VARCHAR     NOT NULL,
    value     VARCHAR,
    created   INTEGER     NOT NULL,
    modified  INTEGER     NOT NULL
);

/*
    All relationships between the tickets.
 */
CREATE TABLE ticket_relationships
(

    id                          VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    ticket_relationship_type_id VARCHAR(36) NOT NULL,
    ticket_id                   VARCHAR(36) NOT NULL,
    child_ticket_id             VARCHAR(36) NOT NULL,
    created                     INTEGER     NOT NULL,
    modified                    INTEGER     NOT NULL
);

/*
    Team belongs to the organization. Multiple teams can belong to one organization.
 */
CREATE TABLE team_organization_mappings
(

    id              VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    team_id         VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36) NOT NULL,
    created         INTEGER     NOT NULL,
    modified        INTEGER     NOT NULL,
    UNIQUE (team_id, organization_id) ON CONFLICT ABORT
);

/*
    Team belongs to one or more projects. Multiple teams can work on multiple projects.
 */
CREATE TABLE team_project_mappings
(

    id         VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    team_id    VARCHAR(36) NOT NULL,
    project_id VARCHAR(36) NOT NULL,
    created    INTEGER     NOT NULL,
    modified   INTEGER     NOT NULL,
    UNIQUE (team_id, project_id) ON CONFLICT ABORT
);

/*
     Repository belongs to project. Multiple repositories can belong to multiple projects.
     So, two projects can actually have the same repository.
 */
CREATE TABLE repository_project_mappings
(

    id            VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    repository_id VARCHAR(36) NOT NULL,
    project_id    VARCHAR(36) NOT NULL,
    created       INTEGER     NOT NULL,
    modified      INTEGER     NOT NULL,
    UNIQUE (repository_id, project_id) ON CONFLICT ABORT
);

/*
    Components to the tickets mappings.
    Component can be mapped to the multiple tickets.
*/
CREATE TABLE component_ticket_mappings
(

    id           VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    component_id VARCHAR(36) NOT NULL,
    ticket_id    VARCHAR(36) NOT NULL,
    created      INTEGER     NOT NULL,
    modified     INTEGER     NOT NULL,
    UNIQUE (component_id, ticket_id) ON CONFLICT ABORT
);

/*
    Time is associated with the tickets.
*/
CREATE TABLE time_ticket_mappings
(

    id        VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    time_id   VARCHAR(36) NOT NULL,
    ticket_id VARCHAR(36) NOT NULL,
    created   INTEGER     NOT NULL,
    modified  INTEGER     NOT NULL,
    UNIQUE (time_id, ticket_id) ON CONFLICT ABORT
);

/*
    Components meta-data:
    Associate the various information with different components.
*/
CREATE TABLE components_meta_data
(

    id           VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    component_id VARCHAR(36) NOT NULL,
    property     VARCHAR     NOT NULL,
    value        VARCHAR,
    created      INTEGER     NOT NULL,
    modified     INTEGER     NOT NULL
);

/*
    Assets can belong to the multiple projects.
    One example of the image used in the context of the project is the project's avatar.
    Projects may have various other assets associated to itself.
    Various documentation for example.
*/
CREATE TABLE asset_project_mappings
(

    id         VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    asset_id   VARCHAR(36) NOT NULL,
    project_id VARCHAR(36) NOT NULL,
    created    INTEGER     NOT NULL,
    modified   INTEGER     NOT NULL,
    UNIQUE (asset_id, project_id) ON CONFLICT ABORT /* TODO: Create the conflict(s) unit test(s). */
);

/*
    Assets can belong to the multiple teams.
    The image used in the context of the team is the team's avatar, for example.
    Teams may have other additions associated to itself. Various documents for example,
*/
CREATE TABLE asset_team_mappings
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    asset_id VARCHAR(36) NOT NULL,
    team_id  VARCHAR(36) NOT NULL,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL,
    UNIQUE (asset_id, team_id) ON CONFLICT ABORT
);

/*
    Assets (attachments for example) can belong to the multiple tickets.
*/
CREATE TABLE asset_ticket_mappings
(

    id        VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    asset_id  VARCHAR(36) NOT NULL,
    ticket_id VARCHAR(36) NOT NULL,
    created   INTEGER     NOT NULL,
    modified  INTEGER     NOT NULL,
    UNIQUE (asset_id, ticket_id) ON CONFLICT ABORT
);

/*
    Assets (attachments for example) can belong to the multiple comments.
*/
CREATE TABLE asset_comment_mappings
(

    id         VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    asset_id   VARCHAR(36) NOT NULL,
    comment_id VARCHAR(36) NOT NULL,
    created    INTEGER     NOT NULL,
    modified   INTEGER     NOT NULL,
    UNIQUE (asset_id, comment_id) ON CONFLICT ABORT
);

/*
    Labels can belong to the label category.
    One single asset can belong to multiple categories.
*/
CREATE TABLE label_label_category_mappings
(

    id                VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    label_id          VARCHAR(36) NOT NULL,
    label_category_id VARCHAR(36) NOT NULL,
    created           INTEGER     NOT NULL,
    modified          INTEGER     NOT NULL,
    UNIQUE (label_id, label_category_id) ON CONFLICT ABORT
);

/*
    Label can be associated with one or more projects.
*/
CREATE TABLE label_project_mappings
(

    id         VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    label_id   VARCHAR(36) NOT NULL,
    project_id VARCHAR(36) NOT NULL,
    created    INTEGER     NOT NULL,
    modified   INTEGER     NOT NULL,
    UNIQUE (label_id, project_id) ON CONFLICT ABORT
);

/*
    Label can be associated with one or more teams.
*/
CREATE TABLE label_team_mappings
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    label_id VARCHAR(36) NOT NULL,
    team_id  VARCHAR(36) NOT NULL,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL,
    UNIQUE (label_id, team_id) ON CONFLICT ABORT
);

/*
    Label can be associated with one or more tickets.
*/
CREATE TABLE label_ticket_mappings
(

    id        VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    label_id  VARCHAR(36) NOT NULL,
    ticket_id VARCHAR(36) NOT NULL,
    created   INTEGER     NOT NULL,
    modified  INTEGER     NOT NULL,
    UNIQUE (label_id, ticket_id) ON CONFLICT ABORT
);

/*
    Label can be associated with one or more assets.
*/
CREATE TABLE label_asset_mappings
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    label_id VARCHAR(36) NOT NULL,
    asset_id VARCHAR(36) NOT NULL,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL,
    UNIQUE (label_id, asset_id) ON CONFLICT ABORT
);

/*
    Comments are usually associated with project tickets:
*/
CREATE TABLE comment_ticket_mappings
(

    id         VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    comment_id VARCHAR(36) NOT NULL,
    ticket_id  VARCHAR(36) NOT NULL,
    created    INTEGER     NOT NULL,
    modified   INTEGER     NOT NULL,
    UNIQUE (comment_id, ticket_id) ON CONFLICT ABORT
);

/*
    OAuth2 mappings:
*/

/*
    Users can be Yandex OAuth2 account users:
 */
CREATE TABLE users_yandex_mappings
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    user_id  VARCHAR(36) NOT NULL UNIQUE,
    username VARCHAR(36) NOT NULL UNIQUE, /* TODO: Populate user mappings with additional meta-data/
                                                The proper JSON file per users mapping/
                                                For example: yandex_users.json
                                        */
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    Users can be Google OAuth2 account users:
 */
CREATE TABLE users_google_mappings
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    user_id  VARCHAR(36) NOT NULL UNIQUE,
    username VARCHAR(36) NOT NULL UNIQUE,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL
);

/*
    User access rights:
*/

/*
    User belongs to organizations:
*/
CREATE TABLE user_organization_mappings
(

    id              VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    user_id         VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36) NOT NULL,
    created         INTEGER     NOT NULL,
    modified        INTEGER     NOT NULL,
    UNIQUE (user_id, organization_id) ON CONFLICT ABORT
);

/*
    User belongs to the organization's teams:
*/
CREATE TABLE user_team_mappings
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    user_id  VARCHAR(36) NOT NULL,
    team_id  VARCHAR(36) NOT NULL,
    created  INTEGER     NOT NULL,
    modified INTEGER     NOT NULL,
    UNIQUE (user_id, team_id) ON CONFLICT ABORT
);

/*
    User has the permissions.
    Each permission has be associated to the proper permission context.
*/
CREATE TABLE permission_user_mappings
(

    id                    VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    permission_id         VARCHAR(36) NOT NULL,
    user_id               VARCHAR(36) NOT NULL,
    permission_context_id VARCHAR(36) NOT NULL,
    created               INTEGER     NOT NULL,
    modified              INTEGER     NOT NULL,
    UNIQUE (user_id, permission_id, permission_context_id) ON CONFLICT ABORT
);


/*
    Team has the permissions.
    Each team permission has be associated to the proper permission context.
    All team members (users) will inherit team's permissions.
*/
CREATE TABLE permission_team_mappings
(

    id                    VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    permission_id         VARCHAR(36) NOT NULL,
    team_id               VARCHAR(36) NOT NULL,
    permission_context_id VARCHAR(36) NOT NULL,
    created               INTEGER     NOT NULL,
    modified              INTEGER     NOT NULL,
    UNIQUE (team_id, permission_id, permission_context_id) ON CONFLICT ABORT
);