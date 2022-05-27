/*
    Version: 1
*/

/*
    Notes:
    - Identifiers in the system are UUID strings (VARCHAR).
*/

/*
    Cleaning up:
        TODO: Comment out before the production.
*/
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS scopes;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS repositories;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users_yandex_mappings;
DROP TABLE IF EXISTS scope_organization_mappings;
DROP TABLE IF EXISTS team_organization_mappings;
DROP TABLE IF EXISTS repository_scope_mappings;
DROP TABLE IF EXISTS image_scope_mappings;

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

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE
);

/*
    The basic project definition.
 */
CREATE TABLE scopes
(

    id    VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title VARCHAR     NOT NULL
);

/*
    Images, defined by identifier and the resource url.
 */
CREATE TABLE images
(

    id    VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    image VARCHAR     NOT NULL UNIQUE
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
    The organization definition. Organization is the owner of the scope.
 */
CREATE TABLE organizations
(

    id            VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    organizations VARCHAR     NOT NULL UNIQUE
);

/*
    The team definition. Organization is the owner of the scope.
 */
CREATE TABLE teams
(

    id   VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    team VARCHAR     NOT NULL UNIQUE
);

/*
    Mappings:
 */

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
    Scope belongs to the organization. Multiple scopes can belong to one organization.
 */
CREATE TABLE scope_organization_mappings
(

    id              VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    organization_id VARCHAR(36) NOT NULL,
    scope_id        VARCHAR(36) NOT NULL,
    UNIQUE (organization_id, scope_id) ON CONFLICT ABORT
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
     Repository belongs to scope. Multiple repositories can belong to multiple scopes.
     So, two projects can actually have the same repository.
 */
CREATE TABLE repository_scope_mappings
(

    id            VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    repository_id VARCHAR(36) NOT NULL,
    scope_id      VARCHAR(36) NOT NULL,
    UNIQUE (repository_id, scope_id) ON CONFLICT ABORT
);

/*
    Image can belong to multiple scopes.
    The image used in the context of the scope is the scope's avatar.
*/
CREATE TABLE image_scope_mappings
(

    id       VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    image_id VARCHAR(36) NOT NULL,
    scope_id VARCHAR(36) NOT NULL,
    UNIQUE (image_id, scope_id) ON CONFLICT ABORT /* TODO: Create the conflict unit test. */
);
