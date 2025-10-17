/*
    HelixTrack Chat Service - Database Schema V2

    A comprehensive internal chat system with real-time messaging,
    multi-entity support, and advanced features.

    Version: 2.0
    Database: PostgreSQL with SQL Cipher encryption
*/

/*
    Notes:
    - UUIDs for all identifiers
    - Soft delete pattern (deleted boolean + deleted_at timestamp)
    - Entity associations: user, team, project, ticket, account, attachment, etc.
    - Real-time events via WebSocket
    - Message threading and replies
    - Typing indicators and read receipts
    - Reactions and attachments
    - Full-text search support
*/

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS message_reaction CASCADE;
DROP TABLE IF EXISTS message_attachment CASCADE;
DROP TABLE IF EXISTS message_read_receipt CASCADE;
DROP TABLE IF EXISTS typing_indicator CASCADE;
DROP TABLE IF EXISTS message CASCADE;
DROP TABLE IF EXISTS chat_participant CASCADE;
DROP TABLE IF EXISTS chat_room CASCADE;
DROP TABLE IF EXISTS user_presence CASCADE;

-- Drop indexes
DROP INDEX IF EXISTS idx_chat_room_entity;
DROP INDEX IF EXISTS idx_chat_room_type;
DROP INDEX IF EXISTS idx_chat_room_created;
DROP INDEX IF EXISTS idx_chat_room_deleted;
DROP INDEX IF EXISTS idx_participant_chat;
DROP INDEX IF EXISTS idx_participant_user;
DROP INDEX IF EXISTS idx_participant_role;
DROP INDEX IF EXISTS idx_message_chat;
DROP INDEX IF EXISTS idx_message_sender;
DROP INDEX IF EXISTS idx_message_parent;
DROP INDEX IF EXISTS idx_message_created;
DROP INDEX IF EXISTS idx_message_type;
DROP INDEX IF EXISTS idx_message_deleted;
DROP INDEX IF EXISTS idx_message_fts;
DROP INDEX IF EXISTS idx_typing_chat;
DROP INDEX IF EXISTS idx_typing_user;
DROP INDEX IF EXISTS idx_read_receipt_message;
DROP INDEX IF EXISTS idx_read_receipt_user;
DROP INDEX IF EXISTS idx_attachment_message;
DROP INDEX IF EXISTS idx_reaction_message;
DROP INDEX IF EXISTS idx_reaction_user;
DROP INDEX IF EXISTS idx_presence_user;

/*
    User presence tracking for online/offline/away status
*/
CREATE TABLE user_presence
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL,
    status          VARCHAR(20) NOT NULL, -- online, offline, away, busy, dnd
    status_message  TEXT,
    last_seen       BIGINT      NOT NULL, -- Unix timestamp
    created_at      BIGINT      NOT NULL,
    updated_at      BIGINT      NOT NULL,

    CONSTRAINT chk_presence_status CHECK (status IN ('online', 'offline', 'away', 'busy', 'dnd'))
);

CREATE UNIQUE INDEX idx_presence_user ON user_presence (user_id);

/*
    Chat rooms - can be associated with any entity

    Entity types:
    - direct: 1-on-1 user chat
    - group: group chat with multiple users
    - team: team-based chat
    - project: project-specific chat
    - ticket: ticket discussion
    - account: account-level chat
    - organization: organization chat
    - attachment: discussion about specific attachment
    - custom: custom entity type
*/
CREATE TABLE chat_room
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255),
    description     TEXT,
    type            VARCHAR(20) NOT NULL, -- direct, group, team, project, ticket, account, organization, attachment, custom
    entity_type     VARCHAR(50), -- type of associated entity
    entity_id       UUID,        -- ID of associated entity
    created_by      UUID        NOT NULL,
    is_private      BOOLEAN     NOT NULL DEFAULT false,
    is_archived     BOOLEAN     NOT NULL DEFAULT false,
    created_at      BIGINT      NOT NULL,
    updated_at      BIGINT      NOT NULL,
    deleted         BOOLEAN     NOT NULL DEFAULT false,
    deleted_at      BIGINT,

    CONSTRAINT chk_room_type CHECK (type IN ('direct', 'group', 'team', 'project', 'ticket', 'account', 'organization', 'attachment', 'custom'))
);

CREATE INDEX idx_chat_room_entity ON chat_room (entity_type, entity_id);
CREATE INDEX idx_chat_room_type ON chat_room (type);
CREATE INDEX idx_chat_room_created ON chat_room (created_at);
CREATE INDEX idx_chat_room_deleted ON chat_room (deleted);

/*
    Chat participants - users in chat rooms

    Roles: owner, admin, moderator, member, guest
*/
CREATE TABLE chat_participant
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id    UUID        NOT NULL REFERENCES chat_room (id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'member',
    is_muted        BOOLEAN     NOT NULL DEFAULT false,
    joined_at       BIGINT      NOT NULL,
    left_at         BIGINT,
    created_at      BIGINT      NOT NULL,
    updated_at      BIGINT      NOT NULL,
    deleted         BOOLEAN     NOT NULL DEFAULT false,
    deleted_at      BIGINT,

    CONSTRAINT chk_participant_role CHECK (role IN ('owner', 'admin', 'moderator', 'member', 'guest')),
    CONSTRAINT uq_chat_participant UNIQUE (chat_room_id, user_id)
);

CREATE INDEX idx_participant_chat ON chat_participant (chat_room_id);
CREATE INDEX idx_participant_user ON chat_participant (user_id);
CREATE INDEX idx_participant_role ON chat_participant (role);

/*
    Messages - the core chat content

    Message types:
    - text: regular text message
    - reply: reply to another message
    - quote: quoted message
    - system: system message (user joined, left, etc.)
    - file: file/attachment message
    - code: code snippet
    - poll: poll message
*/
CREATE TABLE message
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id    UUID        NOT NULL REFERENCES chat_room (id) ON DELETE CASCADE,
    sender_id       UUID        NOT NULL,
    parent_id       UUID REFERENCES message (id) ON DELETE SET NULL, -- for threads/replies
    quoted_message_id UUID REFERENCES message (id) ON DELETE SET NULL, -- for quotes
    type            VARCHAR(20) NOT NULL DEFAULT 'text',
    content         TEXT        NOT NULL,
    content_format  VARCHAR(20) DEFAULT 'plain', -- plain, markdown, html
    metadata        JSONB, -- additional metadata (mentions, links, etc.)
    is_edited       BOOLEAN     NOT NULL DEFAULT false,
    edited_at       BIGINT,
    is_pinned       BOOLEAN     NOT NULL DEFAULT false,
    pinned_at       BIGINT,
    pinned_by       UUID,
    created_at      BIGINT      NOT NULL,
    updated_at      BIGINT      NOT NULL,
    deleted         BOOLEAN     NOT NULL DEFAULT false,
    deleted_at      BIGINT,

    CONSTRAINT chk_message_type CHECK (type IN ('text', 'reply', 'quote', 'system', 'file', 'code', 'poll')),
    CONSTRAINT chk_content_format CHECK (content_format IN ('plain', 'markdown', 'html'))
);

CREATE INDEX idx_message_chat ON message (chat_room_id, created_at DESC);
CREATE INDEX idx_message_sender ON message (sender_id);
CREATE INDEX idx_message_parent ON message (parent_id);
CREATE INDEX idx_message_created ON message (created_at);
CREATE INDEX idx_message_type ON message (type);
CREATE INDEX idx_message_deleted ON message (deleted);

-- Full-text search index for message content
CREATE INDEX idx_message_fts ON message USING GIN (to_tsvector('english', content));

/*
    Typing indicators - real-time typing status
*/
CREATE TABLE typing_indicator
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id    UUID    NOT NULL REFERENCES chat_room (id) ON DELETE CASCADE,
    user_id         UUID    NOT NULL,
    is_typing       BOOLEAN NOT NULL DEFAULT true,
    started_at      BIGINT  NOT NULL,
    expires_at      BIGINT  NOT NULL, -- auto-expire after 5 seconds

    CONSTRAINT uq_typing_indicator UNIQUE (chat_room_id, user_id)
);

CREATE INDEX idx_typing_chat ON typing_indicator (chat_room_id);
CREATE INDEX idx_typing_user ON typing_indicator (user_id);

/*
    Read receipts - message read status
*/
CREATE TABLE message_read_receipt
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID   NOT NULL REFERENCES message (id) ON DELETE CASCADE,
    user_id         UUID   NOT NULL,
    read_at         BIGINT NOT NULL,
    created_at      BIGINT NOT NULL,

    CONSTRAINT uq_read_receipt UNIQUE (message_id, user_id)
);

CREATE INDEX idx_read_receipt_message ON message_read_receipt (message_id);
CREATE INDEX idx_read_receipt_user ON message_read_receipt (user_id);

/*
    Message attachments - files attached to messages
*/
CREATE TABLE message_attachment
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID         NOT NULL REFERENCES message (id) ON DELETE CASCADE,
    file_name       VARCHAR(255) NOT NULL,
    file_type       VARCHAR(100),
    file_size       BIGINT       NOT NULL,
    file_url        TEXT         NOT NULL,
    thumbnail_url   TEXT,
    metadata        JSONB, -- dimensions, duration, etc.
    uploaded_by     UUID         NOT NULL,
    created_at      BIGINT       NOT NULL,
    deleted         BOOLEAN      NOT NULL DEFAULT false,
    deleted_at      BIGINT
);

CREATE INDEX idx_attachment_message ON message_attachment (message_id);

/*
    Message reactions - emoji reactions to messages
*/
CREATE TABLE message_reaction
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID        NOT NULL REFERENCES message (id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL,
    emoji           VARCHAR(50) NOT NULL, -- emoji unicode or :emoji_name:
    created_at      BIGINT      NOT NULL,

    CONSTRAINT uq_message_reaction UNIQUE (message_id, user_id, emoji)
);

CREATE INDEX idx_reaction_message ON message_reaction (message_id);
CREATE INDEX idx_reaction_user ON message_reaction (user_id);

/*
    External chat integrations (from V1, kept for compatibility)
*/
CREATE TABLE chat_external_integration
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id    UUID        NOT NULL REFERENCES chat_room (id) ON DELETE CASCADE,
    provider        VARCHAR(50) NOT NULL, -- slack, telegram, yandex, google, whatsapp
    external_id     VARCHAR(255) NOT NULL,
    config          JSONB, -- provider-specific configuration
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    created_at      BIGINT      NOT NULL,
    updated_at      BIGINT      NOT NULL,
    deleted         BOOLEAN     NOT NULL DEFAULT false,
    deleted_at      BIGINT,

    CONSTRAINT chk_provider CHECK (provider IN ('slack', 'telegram', 'yandex', 'google', 'whatsapp', 'custom'))
);

CREATE INDEX idx_integration_chat ON chat_external_integration (chat_room_id);
CREATE INDEX idx_integration_provider ON chat_external_integration (provider);

-- Comments for documentation
COMMENT ON TABLE chat_room IS 'Chat rooms with multi-entity support';
COMMENT ON TABLE chat_participant IS 'Users participating in chat rooms';
COMMENT ON TABLE message IS 'Chat messages with threading and reply support';
COMMENT ON TABLE typing_indicator IS 'Real-time typing indicators';
COMMENT ON TABLE message_read_receipt IS 'Message read status tracking';
COMMENT ON TABLE message_attachment IS 'File attachments for messages';
COMMENT ON TABLE message_reaction IS 'Emoji reactions to messages';
COMMENT ON TABLE user_presence IS 'User online/offline presence tracking';
COMMENT ON TABLE chat_external_integration IS 'Integration with external chat providers';
