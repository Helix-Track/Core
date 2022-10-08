CREATE TABLE document
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    project_id  TEXT    NOT NULL,
    document_id TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
CREATE TABLE content_document_mapping
(
    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id TEXT    NOT NULL UNIQUE,
    content     TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL );
