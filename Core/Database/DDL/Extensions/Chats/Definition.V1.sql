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

DROP TABLE IF EXISTS chat;
DROP TABLE IF EXISTS chat_yandex_mapping;
DROP TABLE IF EXISTS chat_google_mapping;
DROP TABLE IF EXISTS chat_slack_mapping;
DROP TABLE IF EXISTS chat_telegram_mapping;
DROP TABLE IF EXISTS chat_whatsapp_mapping;

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
CREATE TABLE chat
(

    id              TEXT    NOT NULL PRIMARY KEY UNIQUE,
    title           TEXT    NOT NULL,
    organization_id TEXT UNIQUE,
    team_id         TEXT UNIQUE,
    project_id      TEXT UNIQUE,
    ticket_id       TEXT UNIQUE,
    created         INTEGER NOT NULL,
    modified        INTEGER NOT NULL,
    deleted         BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX get_by_title ON chat (title);
CREATE INDEX get_by_team_id ON chat (team_id);
CREATE INDEX get_by_ticket_id ON chat (ticket_id);
CREATE INDEX get_by_project_id ON chat (project_id);
CREATE INDEX get_by_organization_id ON chat (organization_id);
CREATE INDEX get_by_deleted ON chat (deleted);
CREATE INDEX get_by_created ON chat (created);
CREATE INDEX get_by_modified ON chat (modified);
CREATE INDEX get_by_created_and_modified ON chat (created, modified);

/*
    Chats can be provided by the Yandex Messenger.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chat_yandex_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX get_yandex_chat_mappings_by_chat_id ON chat_yandex_mapping (chat_id);


/*
    Chats can be provided by the Slack.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chat_slack_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX get_slack_chat_mappings_by_chat_id ON chat_slack_mapping (chat_id);

/*
    Chats can be provided by the Telegram.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chat_telegram_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX get_telegram_chat_mappings_by_chat_id ON chat_telegram_mapping (chat_id);


/*
    Chats can be provided by the Google.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chat_google_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX get_google_chat_mappings_by_chat_id ON chat_google_mapping (chat_id);

/*
    Chats can be provided by the WhatsApp.
    The table contains all the meta-data associated with it.
*/
CREATE TABLE chat_whatsapp_mapping
(

    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL CHECK (deleted IN (0, 1))
);

CREATE INDEX get_whatsapp_chat_mappings_by_chat_id ON chat_whatsapp_mapping (chat_id);