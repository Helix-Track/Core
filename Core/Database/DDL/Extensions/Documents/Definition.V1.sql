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

DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS content_document_mappings;

DROP INDEX IF EXISTS get_by_title;
DROP INDEX IF EXISTS get_by_project_id;
DROP INDEX IF EXISTS get_by_document_id;
DROP INDEX IF EXISTS get_by_deleted;
DROP INDEX IF EXISTS get_by_created;
DROP INDEX IF EXISTS get_by_modified;
DROP INDEX IF EXISTS get_by_created_and_modified;
DROP INDEX IF EXISTS content_document_mappings_get_by_document_id;

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
    deleted     BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX get_by_title ON documents (title);
CREATE INDEX get_by_project_id ON documents (project_id);
CREATE INDEX get_by_document_id ON documents (document_id);
CREATE INDEX get_by_deleted ON documents (deleted);
CREATE INDEX get_by_created ON documents (created);
CREATE INDEX get_by_modified ON documents (modified);
CREATE INDEX get_by_created_and_modified ON documents (created, modified);

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

CREATE INDEX content_document_mappings_get_by_document_id ON content_document_mappings (document_id);