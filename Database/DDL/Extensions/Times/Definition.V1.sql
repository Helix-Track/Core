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
DROP TABLE IF EXISTS times;

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
CREATE TABLE times
(

    id          VARCHAR(36) NOT NULL PRIMARY KEY UNIQUE,
    created     INTEGER     NOT NULL,
    modified    INTEGER     NOT NULL,
    amount      INTEGER     NOT NULL,

    unit        VARCHAR CHECK (
            unit IN (
                     'Minute', 'Hour', 'Day', 'Week', 'Month'
            )
        )                   NOT NULL DEFAULT 'Hour',

    title       VARCHAR,
    description VARCHAR,
    ticket_id   VARCHAR(36) NOT NULL,
    deleted     BOOLEAN     NOT NULL CHECK (deleted IN (0, 1))
);

