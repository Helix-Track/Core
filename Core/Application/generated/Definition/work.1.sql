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
    deleted         BOOLEAN NOT NULL );
CREATE TABLE chats_yandex_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE chats_slack_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE chats_telegram_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE chats_google_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
CREATE TABLE chats_whatsapp_mappings
(
    id       TEXT    NOT NULL PRIMARY KEY UNIQUE,
    chat_id  TEXT    NOT NULL UNIQUE,
    property TEXT    NOT NULL,
    value    TEXT,
    created  INTEGER NOT NULL,
    modified INTEGER NOT NULL,
    deleted  BOOLEAN NOT NULL );
