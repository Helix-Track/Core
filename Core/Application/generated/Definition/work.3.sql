CREATE TABLE times
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    amount      INTEGER NOT NULL,
    unit_id     TEXT    NOT NULL,
    title       TEXT,
    description TEXT,
    ticket_id   TEXT    NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE units
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
