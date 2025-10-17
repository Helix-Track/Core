/*
    Migration Script: Chat Extension V1 -> V2

    This script migrates the chat extension from external integration only (V1)
    to a full internal chat system with messaging capabilities (V2).

    Changes:
    - Adds comprehensive internal chat tables
    - Preserves existing external integration data
    - Adds real-time messaging, typing indicators, read receipts
    - Adds message reactions and attachments
    - Adds user presence tracking
*/

-- Step 1: Create new V2 tables
-- (These don't conflict with V1 tables)

CREATE TABLE IF NOT EXISTS user_presence
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL,
    status          VARCHAR(20) NOT NULL,
    status_message  TEXT,
    last_seen       BIGINT      NOT NULL,
    created_at      BIGINT      NOT NULL,
    updated_at      BIGINT      NOT NULL,
    CONSTRAINT chk_presence_status CHECK (status IN ('online', 'offline', 'away', 'busy', 'dnd'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_presence_user ON user_presence (user_id);

CREATE TABLE IF NOT EXISTS chat_room
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255),
    description     TEXT,
    type            VARCHAR(20) NOT NULL,
    entity_type     VARCHAR(50),
    entity_id       UUID,
    created_by      UUID        NOT NULL,
    is_private      BOOLEAN     NOT NULL DEFAULT false,
    is_archived     BOOLEAN     NOT NULL DEFAULT false,
    created_at      BIGINT      NOT NULL,
    updated_at      BIGINT      NOT NULL,
    deleted         BOOLEAN     NOT NULL DEFAULT false,
    deleted_at      BIGINT,
    CONSTRAINT chk_room_type CHECK (type IN ('direct', 'group', 'team', 'project', 'ticket', 'account', 'organization', 'attachment', 'custom'))
);

CREATE INDEX IF NOT EXISTS idx_chat_room_entity ON chat_room (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_chat_room_type ON chat_room (type);
CREATE INDEX IF NOT EXISTS idx_chat_room_created ON chat_room (created_at);
CREATE INDEX IF NOT EXISTS idx_chat_room_deleted ON chat_room (deleted);

CREATE TABLE IF NOT EXISTS chat_participant
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

CREATE INDEX IF NOT EXISTS idx_participant_chat ON chat_participant (chat_room_id);
CREATE INDEX IF NOT EXISTS idx_participant_user ON chat_participant (user_id);
CREATE INDEX IF NOT EXISTS idx_participant_role ON chat_participant (role);

CREATE TABLE IF NOT EXISTS message
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id    UUID        NOT NULL REFERENCES chat_room (id) ON DELETE CASCADE,
    sender_id       UUID        NOT NULL,
    parent_id       UUID REFERENCES message (id) ON DELETE SET NULL,
    quoted_message_id UUID REFERENCES message (id) ON DELETE SET NULL,
    type            VARCHAR(20) NOT NULL DEFAULT 'text',
    content         TEXT        NOT NULL,
    content_format  VARCHAR(20) DEFAULT 'plain',
    metadata        JSONB,
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

CREATE INDEX IF NOT EXISTS idx_message_chat ON message (chat_room_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_message_sender ON message (sender_id);
CREATE INDEX IF NOT EXISTS idx_message_parent ON message (parent_id);
CREATE INDEX IF NOT EXISTS idx_message_created ON message (created_at);
CREATE INDEX IF NOT EXISTS idx_message_type ON message (type);
CREATE INDEX IF NOT EXISTS idx_message_deleted ON message (deleted);
CREATE INDEX IF NOT EXISTS idx_message_fts ON message USING GIN (to_tsvector('english', content));

CREATE TABLE IF NOT EXISTS typing_indicator
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id    UUID    NOT NULL REFERENCES chat_room (id) ON DELETE CASCADE,
    user_id         UUID    NOT NULL,
    is_typing       BOOLEAN NOT NULL DEFAULT true,
    started_at      BIGINT  NOT NULL,
    expires_at      BIGINT  NOT NULL,
    CONSTRAINT uq_typing_indicator UNIQUE (chat_room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_typing_chat ON typing_indicator (chat_room_id);
CREATE INDEX IF NOT EXISTS idx_typing_user ON typing_indicator (user_id);

CREATE TABLE IF NOT EXISTS message_read_receipt
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID   NOT NULL REFERENCES message (id) ON DELETE CASCADE,
    user_id         UUID   NOT NULL,
    read_at         BIGINT NOT NULL,
    created_at      BIGINT NOT NULL,
    CONSTRAINT uq_read_receipt UNIQUE (message_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_read_receipt_message ON message_read_receipt (message_id);
CREATE INDEX IF NOT EXISTS idx_read_receipt_user ON message_read_receipt (user_id);

CREATE TABLE IF NOT EXISTS message_attachment
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID         NOT NULL REFERENCES message (id) ON DELETE CASCADE,
    file_name       VARCHAR(255) NOT NULL,
    file_type       VARCHAR(100),
    file_size       BIGINT       NOT NULL,
    file_url        TEXT         NOT NULL,
    thumbnail_url   TEXT,
    metadata        JSONB,
    uploaded_by     UUID         NOT NULL,
    created_at      BIGINT       NOT NULL,
    deleted         BOOLEAN      NOT NULL DEFAULT false,
    deleted_at      BIGINT
);

CREATE INDEX IF NOT EXISTS idx_attachment_message ON message_attachment (message_id);

CREATE TABLE IF NOT EXISTS message_reaction
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID        NOT NULL REFERENCES message (id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL,
    emoji           VARCHAR(50) NOT NULL,
    created_at      BIGINT      NOT NULL,
    CONSTRAINT uq_message_reaction UNIQUE (message_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_reaction_message ON message_reaction (message_id);
CREATE INDEX IF NOT EXISTS idx_reaction_user ON message_reaction (user_id);

-- Step 2: Migrate existing V1 chat data to V2 structure
-- Create chat rooms for existing external chat mappings

INSERT INTO chat_room (id, name, type, entity_type, entity_id, created_by, created_at, updated_at, deleted)
SELECT
    CAST(c.id AS UUID),
    c.title,
    CASE
        WHEN c.team_id IS NOT NULL THEN 'team'
        WHEN c.project_id IS NOT NULL THEN 'project'
        WHEN c.ticket_id IS NOT NULL THEN 'ticket'
        WHEN c.organization_id IS NOT NULL THEN 'organization'
        ELSE 'custom'
    END as type,
    CASE
        WHEN c.team_id IS NOT NULL THEN 'team'
        WHEN c.project_id IS NOT NULL THEN 'project'
        WHEN c.ticket_id IS NOT NULL THEN 'ticket'
        WHEN c.organization_id IS NOT NULL THEN 'organization'
        ELSE NULL
    END as entity_type,
    COALESCE(
        CAST(c.team_id AS UUID),
        CAST(c.project_id AS UUID),
        CAST(c.ticket_id AS UUID),
        CAST(c.organization_id AS UUID)
    ) as entity_id,
    gen_random_uuid() as created_by, -- placeholder, update with actual user
    c.created,
    c.modified,
    c.deleted
FROM chat c
WHERE NOT EXISTS (
    SELECT 1 FROM chat_room cr WHERE cr.id = CAST(c.id AS UUID)
);

-- Step 3: Create external integration table and migrate V1 integrations

CREATE TABLE IF NOT EXISTS chat_external_integration
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id    UUID        NOT NULL REFERENCES chat_room (id) ON DELETE CASCADE,
    provider        VARCHAR(50) NOT NULL,
    external_id     VARCHAR(255) NOT NULL,
    config          JSONB,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    created_at      BIGINT      NOT NULL,
    updated_at      BIGINT      NOT NULL,
    deleted         BOOLEAN     NOT NULL DEFAULT false,
    deleted_at      BIGINT,
    CONSTRAINT chk_provider CHECK (provider IN ('slack', 'telegram', 'yandex', 'google', 'whatsapp', 'custom'))
);

CREATE INDEX IF NOT EXISTS idx_integration_chat ON chat_external_integration (chat_room_id);
CREATE INDEX IF NOT EXISTS idx_integration_provider ON chat_external_integration (provider);

-- Migrate Slack mappings
INSERT INTO chat_external_integration (chat_room_id, provider, external_id, config, created_at, updated_at, deleted, deleted_at)
SELECT
    CAST(chat_id AS UUID),
    'slack',
    id,
    jsonb_build_object('property', property, 'value', value),
    created,
    modified,
    deleted,
    CASE WHEN deleted THEN modified ELSE NULL END
FROM chat_slack_mapping
WHERE NOT EXISTS (
    SELECT 1 FROM chat_external_integration cei
    WHERE cei.chat_room_id = CAST(chat_slack_mapping.chat_id AS UUID)
    AND cei.provider = 'slack'
);

-- Migrate Telegram mappings
INSERT INTO chat_external_integration (chat_room_id, provider, external_id, config, created_at, updated_at, deleted, deleted_at)
SELECT
    CAST(chat_id AS UUID),
    'telegram',
    id,
    jsonb_build_object('property', property, 'value', value),
    created,
    modified,
    deleted,
    CASE WHEN deleted THEN modified ELSE NULL END
FROM chat_telegram_mapping
WHERE NOT EXISTS (
    SELECT 1 FROM chat_external_integration cei
    WHERE cei.chat_room_id = CAST(chat_telegram_mapping.chat_id AS UUID)
    AND cei.provider = 'telegram'
);

-- Migrate Yandex mappings
INSERT INTO chat_external_integration (chat_room_id, provider, external_id, config, created_at, updated_at, deleted, deleted_at)
SELECT
    CAST(chat_id AS UUID),
    'yandex',
    id,
    jsonb_build_object('property', property, 'value', value),
    created,
    modified,
    deleted,
    CASE WHEN deleted THEN modified ELSE NULL END
FROM chat_yandex_mapping
WHERE NOT EXISTS (
    SELECT 1 FROM chat_external_integration cei
    WHERE cei.chat_room_id = CAST(chat_yandex_mapping.chat_id AS UUID)
    AND cei.provider = 'yandex'
);

-- Migrate Google mappings
INSERT INTO chat_external_integration (chat_room_id, provider, external_id, config, created_at, updated_at, deleted, deleted_at)
SELECT
    CAST(chat_id AS UUID),
    'google',
    id,
    jsonb_build_object('property', property, 'value', value),
    created,
    modified,
    deleted,
    CASE WHEN deleted THEN modified ELSE NULL END
FROM chat_google_mapping
WHERE NOT EXISTS (
    SELECT 1 FROM chat_external_integration cei
    WHERE cei.chat_room_id = CAST(chat_google_mapping.chat_id AS UUID)
    AND cei.provider = 'google'
);

-- Migrate WhatsApp mappings
INSERT INTO chat_external_integration (chat_room_id, provider, external_id, config, created_at, updated_at, deleted, deleted_at)
SELECT
    CAST(chat_id AS UUID),
    'whatsapp',
    id,
    jsonb_build_object('property', property, 'value', value),
    created,
    modified,
    deleted,
    CASE WHEN deleted THEN modified ELSE NULL END
FROM chat_whatsapp_mapping
WHERE NOT EXISTS (
    SELECT 1 FROM chat_external_integration cei
    WHERE cei.chat_room_id = CAST(chat_whatsapp_mapping.chat_id AS UUID)
    AND cei.provider = 'whatsapp'
);

-- Step 4: V1 tables can remain for backwards compatibility
-- or be dropped if no longer needed:
-- DROP TABLE chat_whatsapp_mapping;
-- DROP TABLE chat_google_mapping;
-- DROP TABLE chat_telegram_mapping;
-- DROP TABLE chat_slack_mapping;
-- DROP TABLE chat_yandex_mapping;
-- DROP TABLE chat;

-- Migration complete
SELECT 'Chat Extension Migration V1 -> V2 completed successfully' as status;
