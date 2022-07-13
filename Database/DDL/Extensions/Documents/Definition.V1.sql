/*
    Version: 1
*/

/*
    Notes:

    - TODOs: https://github.com/orgs/red-elf/projects/2/views/1
    - Identifiers in the system are UUID strings.
    - Mapping tables are used for binding entities and defining relationships.
        Mapping tables are used as well to append properties to the entities.
*/

DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS content_document_mappings;

/*
    Documents.
    Users can create the project documentation.
    Each document (the root) belongs to the project. It can also belong to the the parent document.
*/
CREATE TABLE documents
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title       TEXT    NOT NULL,
    project_id  TEXT    NOT NULL,
    document_id TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

/*
    Each document is associated with its content.
    The content field can contain the raw content or the 'identifier' of the content asset of some kind.
    Other content type extensions can create additional document mappings tables.
*/
CREATE TABLE content_document_mappings
(

    id          TEXT    NOT NULL PRIMARY KEY UNIQUE,
    document_id TEXT    NOT NULL UNIQUE,
    content     TEXT,
    created     INTEGER NOT NULL,
    modified    INTEGER NOT NULL,
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);