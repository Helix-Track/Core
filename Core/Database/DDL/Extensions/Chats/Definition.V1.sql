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

DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS chats_yandex_mappings;
DROP TABLE IF EXISTS chats_google_mappings;
DROP TABLE IF EXISTS chats_slack_mappings;
DROP TABLE IF EXISTS chats_telegram_mappings;
DROP TABLE IF EXISTS chats_whatsapp_mappings;

DROP INDEX IF EXISTS get_by_title;
DROP INDEX IF EXISTS get_by_team_id;
DROP INDEX IF EXISTS get_by_ticket_id;
DROP INDEX IF EXISTS get_by_project_id;
DROP INDEX IF EXISTS get_by_organization_id;
DROP INDEX IF EXISTS get_by_deleted;
DROP INDEX IF EXISTS get_by_created;
DROP INDEX IF EXISTS get_by_modified;
DROP INDEX IF EXISTS get_by_created_and_modified;

DROP INDEX IF EXISTS get_yandex_chat_mappings_by_chat_id;
DROP INDEX IF EXISTS get_slack_chat_mappings_by_chat_id;
DROP INDEX IF EXISTS get_telegram_chat_mappings_by_chat_id;
DROP INDEX IF EXISTS get_google_chat_mappings_by_chat_id;
DROP INDEX IF EXISTS get_whatsapp_chat_mappings_by_chat_id;

/*
    Chat support for the projects.

    Chat room can be connected with:
        - Organization
        - Team
        - Project
        - Ticket.

    Each of these entities can have up to one chat room.
    Chat rooms can be provided by the various vendors:
        - Slack
        - Yandex
        - Google
        - Telegram
        - WhatsApp, etc.
*/
CREATE TABLE chats
(

    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title           TEXT    NOT NULL,
    organization_id TEXT UNIQUE,
    team_id         TEXT UNIQUE,
    project_id      TEXT UNIQUE,
    ticket_id       TEXT UNIQUE,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX get_by_title ON chats (title);
CREATE INDEX get_by_team_id ON chats (team_id);
CREATE INDEX get_by_ticket_id ON chats (ticket_id);
CREATE INDEX get_by_project_id ON chats (project_id);
CREATE INDEX get_by_organization_id ON chats (organization_id);
CREATE INDEX get_by_deleted ON chats (deleted);
CREATE INDEX get_by_created ON chats (created);
CREATE INDEX get_by_modified ON chats (modified);
CREATE INDEX get_by_created_and_modified ON chats (created, modified);

/*
    Chats can be provided by the Yandex Messenger.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chats_yandex_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX get_yandex_chat_mappings_by_chat_id ON chats_yandex_mappings (chat_id);


/*
    Chats can be provided by the Slack.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chats_slack_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX get_slack_chat_mappings_by_chat_id ON chats_slack_mappings (chat_id);

/*
    Chats can be provided by the Telegram.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chats_telegram_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX get_telegram_chat_mappings_by_chat_id ON chats_telegram_mappings (chat_id);


/*
    Chats can be provided by the Google.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chats_google_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX get_google_chat_mappings_by_chat_id ON chats_google_mappings (chat_id);

/*
    Chats can be provided by the WhatsApp.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chats_whatsapp_mappings
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1)) DEFAULT 0
);

CREATE INDEX get_whatsapp_chat_mappings_by_chat_id ON chats_whatsapp_mappings (chat_id);