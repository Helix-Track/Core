/*
    Version: 1
*/

/*
    Notes:
    - Identifiers in the system are UUID strings (VARCHAR with the size of 36).
*/

/*
    Cleaning up:
        TODO: Comment out before the production.
*/
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS repositories;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS users_yandex_mappings;
DROP TABLE IF EXISTS user_organization_mappings;
DROP TABLE IF EXISTS user_team_mappings;
DROP TABLE IF EXISTS permission_user_mappings;
DROP TABLE IF EXISTS project_organization_mappings;
DROP TABLE IF EXISTS team_organization_mappings;
DROP TABLE IF EXISTS team_project_mappings;
DROP TABLE IF EXISTS repository_project_mappings;
DROP TABLE IF EXISTS asset_project_mappings;

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

    id VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE
);

/*
    The basic project definition.
 */
CREATE TABLE projects
(

    id          VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title       VARCHAR     NOT NULL,
    description VARCHAR     NOT NULL
);

/*
    Images, attachments, etc.
    Defined by the identifier and the resource url.
 */
CREATE TABLE assets
(

    id  VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    url VARCHAR     NOT NULL UNIQUE
);

/*
      The code repositories - Identified by the identifier and the repository URL.
      Default repository type is Git repository.
      TODO:
        For supporting multiple repository types the 'type' column to be introduced.
 */
CREATE TABLE repositories
(

    id         VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    repository VARCHAR     NOT NULL UNIQUE
);

/*
    The organization definition. Organization is the owner of the project.
 */
CREATE TABLE organizations
(

    id    VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title VARCHAR     NOT NULL UNIQUE
);

/*
    The team definition. Organization is the owner of the team.
 */
CREATE TABLE teams
(

    id    VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title VARCHAR     NOT NULL UNIQUE
);

/*
    Permission definitions.
 */
CREATE TABLE permissions
(

    id    VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title VARCHAR     NOT NULL UNIQUE
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
    UNIQUE (organization_id, project_id) ON CONFLICT ABORT
);

/*
    Team belongs to the organization. Multiple teams can belong to one organization.
 */
CREATE TABLE team_organization_mappings
(

    id              VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    team_id         VARCHAR(36) NOT NULL,
    organization_id VARCHAR(36) NOT NULL,
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
    UNIQUE (repository_id, project_id) ON CONFLICT ABORT
);

/*
    Image (asset) can belong to multiple projects.
    The image used in the context of the project is the project's avatar.
*/
CREATE TABLE asset_project_mappings
(

    id         VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    asset_id   VARCHAR(36) NOT NULL,
    project_id VARCHAR(36) NOT NULL,
    UNIQUE (asset_id, project_id) ON CONFLICT ABORT /* TODO: Create the conflict(s) unit test(s). */
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
    username VARCHAR(36) NOT NULL UNIQUE
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
    UNIQUE (user_id, organization_id) ON CONFLICT ABORT
);

/*
    User belongs to the organization's teams:
*/
CREATE TABLE user_team_mappings
(

    id      VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    user_id VARCHAR(36) NOT NULL,
    team_id VARCHAR(36) NOT NULL,
    UNIQUE (user_id, team_id) ON CONFLICT ABORT
);

/*
    User has the permissions.
    Each permission has be associated to the proper context.
    For example:

        organization.project
        organization.team
*/
CREATE TABLE permission_user_mappings
(

    id            VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    permission_id VARCHAR(36) NOT NULL,
    user_id       VARCHAR(36) NOT NULL,
    context       VARCHAR     NOT NULL,
    UNIQUE (user_id, permission_id, context) ON CONFLICT ABORT
);

/*
    TODO: Map permissions to teams.
*/