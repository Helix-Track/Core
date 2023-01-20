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
    deleted     BOOLEAN NOT NULL);
CREATE TABLE time_unit
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL UNIQUE,
    description TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL);
