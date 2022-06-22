/*
    Version: 1
*/

/*
    Notes:

    - Identifiers in the system are UUID strings (VARCHAR with the size of 36).
    - Mapping tables are used for binding entities and defining relationships.
        Mapping tables are used as well to append properties to the entities.
*/

/*
    TODOs:

      Main:

    - TODO: Indexes
    - TODO: Limit the varchar lengths

      Features:

*/

/*
    Cleaning up:
        TODO: Comment out before the production.
            Shell scripts should execute each SQl with TO-DO check. If TO-DO found -> fail.
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

    id          VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title       VARCHAR     NOT NULL,
    project_id  VARCHAR(36) NOT NULL,
    document_id VARCHAR(36),
    created     INTEGER     NOT NULL,
    modified    INTEGER     NOT NULL,
    deleted     BOOLEAN     NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

/*
    Each document is associated with its content.
    The content field can contain the raw content or the 'identifier' of the content asset of some kind.
 */
CREATE TABLE content_document_mappings
(

    id          VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    document_id VARCHAR(36) NOT NULL UNIQUE,
    content     VARCHAR,
    created     INTEGER     NOT NULL,
    modified    INTEGER     NOT NULL,
    deleted     BOOLEAN     NOT NULL CHECK (deleted IN (0, 1))
);