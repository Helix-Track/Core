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
DROP TABLE IF EXISTS chats;

/*
    Chats.

    TODO: Tbd.
 */
CREATE TABLE chats
(

    id          VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    title       VARCHAR     NOT NULL,
    created     INTEGER     NOT NULL,
    modified    INTEGER     NOT NULL,
    deleted     BOOLEAN     NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);
