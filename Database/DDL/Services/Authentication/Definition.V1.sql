/*
    Version: 1
*/

/*
    Notes:

    - The main project board: https://github.com/orgs/red-elf/projects/2/views/1
    - Identifiers in the system are UUID strings.
    - Mapping tables are used for binding entities and defining relationships.
        Mapping tables are used as well to append properties to the entities.
    - Additional tables are defined to provide the meta-data to entities of the system.
    - To follow the order of entities definition in the system follow the 'DROP TABLE' directives.
*/

DROP TABLE IF EXISTS users;

DROP INDEX IF EXISTS users_get_by_username;
DROP INDEX IF EXISTS users_get_by_username_and_password;
DROP INDEX IF EXISTS users_get_by_token;
DROP INDEX IF EXISTS users_get_by_username_and_token;
DROP INDEX IF EXISTS users_get_by_created;
DROP INDEX IF EXISTS users_get_by_modified;
DROP INDEX IF EXISTS users_get_by_deleted;
DROP INDEX IF EXISTS users_get_by_created_and_modified;


/*
     System's users.
     User is identified by the unique identifier (id).
     Since there may be different types of users, different kinds of data
     can be mapped (associated) with the user ID.
     For that purpose there are other mappings to the user ID such as Yandex OAuth2 mappings for example.
*/
CREATE TABLE users
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    username TEXT    NOT NULL UNIQUE,
    password TEXT    NOT NULL,
    token    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL
);

CREATE INDEX users_get_by_username ON users (username);
CREATE INDEX users_get_by_token ON users (token);
CREATE INDEX users_get_by_username_and_token ON users (username, token);
CREATE INDEX users_get_by_username_and_password ON users (username, password);
CREATE INDEX users_get_by_deleted ON users (deleted);
CREATE INDEX users_get_by_created ON users (created);
CREATE INDEX users_get_by_modified ON users (modified);
CREATE INDEX users_get_by_created_and_modified ON users (created, modified);
