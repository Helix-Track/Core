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

DROP TABLE IF EXISTS time_tracking;
DROP TABLE IF EXISTS time_unit;

DROP INDEX IF EXISTS get_by_title;
DROP INDEX IF EXISTS get_by_description;
DROP INDEX IF EXISTS get_by_title_and_description;
DROP INDEX IF EXISTS get_by_created;
DROP INDEX IF EXISTS get_by_modified;
DROP INDEX IF EXISTS get_by_created_and_modified;
DROP INDEX IF EXISTS get_by_deleted;
DROP INDEX IF EXISTS get_by_ticket_id;
DROP INDEX IF EXISTS get_by_ticket_id_and_title;
DROP INDEX IF EXISTS units_get_by_title;
DROP INDEX IF EXISTS units_get_by_created;
DROP INDEX IF EXISTS units_get_by_deleted;
DROP INDEX IF EXISTS units_get_by_modified;
DROP INDEX IF EXISTS units_get_by_created_and_modified;

/*
    Time tracking.
    Time is tracked against the tickets.
    One entry is associated with the parent ticket and it contains the information:
        - How much time
        - Unit (time unit)
        - The title for the performed work (optional)
        - The description for the performed work (optional)
        - The identifier of the work ticket.
*/
CREATE TABLE time_tracking
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    amount      INTEGER NOT NULL,
    unit_id     TEXT    NOT NULL,
    title       TEXT,
    description TEXT,
    ticket_id   TEXT    NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX get_by_title ON time_tracking (title);
CREATE INDEX get_by_description ON time_tracking (description);
CREATE INDEX get_by_title_and_description ON time_tracking (title, description);
CREATE INDEX get_by_ticket_id ON time_tracking (ticket_id);
CREATE INDEX get_by_ticket_id_and_title ON time_tracking (ticket_id, title);
CREATE INDEX get_by_deleted ON time_tracking (deleted);
CREATE INDEX get_by_created ON time_tracking (created);
CREATE INDEX get_by_modified ON time_tracking (modified);
CREATE INDEX get_by_created_and_modified ON time_tracking (created, modified);

/*
    'Minute', 'Hour', 'Day', 'Week', 'Month'
*/
CREATE TABLE time_unit
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX units_get_by_title ON time_unit (title);
CREATE INDEX units_get_by_created ON time_unit (created);
CREATE INDEX units_get_by_deleted ON time_unit (deleted);
CREATE INDEX units_get_by_modified ON time_unit (modified);
CREATE INDEX units_get_by_created_and_modified ON time_unit (created, modified);
