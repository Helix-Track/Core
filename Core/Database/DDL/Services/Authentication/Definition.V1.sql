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

DROP TABLE IF EXISTS user;

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
CREATE TABLE user
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX users_get_by_created ON user (created);
CREATE INDEX users_get_by_modified ON user (modified);
CREATE INDEX users_get_by_deleted ON user (deleted);
CREATE INDEX users_get_by_created_and_modified ON user (created, modified);
