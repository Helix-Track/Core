CREATE TABLE user
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    username TEXT    NOT NULL UNIQUE,
    password TEXT    NOT NULL,
    token    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
