-- HelixTrack Chat Service - Initial Database Schema
-- Purpose: Complete database schema for chat microservice
-- Author: Claude Code
-- Date: 2025-10-17
-- PostgreSQL 12+ with UUID support

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ===========================
-- Table 1: user_presence
-- ===========================
CREATE TABLE IF NOT EXISTS user_presence (
    user_id         UUID PRIMARY KEY,
    status          VARCHAR(20) NOT NULL DEFAULT 'offline',
    status_message  VARCHAR(255),
    last_seen       BIGINT NOT NULL,
    updated_at      BIGINT NOT NULL,

    CONSTRAINT chk_presence_status CHECK (status IN ('online', 'offline', 'away', 'busy', 'dnd'))
);

CREATE INDEX idx_user_presence_status ON user_presence(status);
CREATE INDEX idx_user_presence_last_seen ON user_presence(last_seen DESC);

-- ===========================
-- Table 2: chat_room
-- ===========================
CREATE TABLE IF NOT EXISTS chat_room (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    type            VARCHAR(50) NOT NULL,
    entity_type     VARCHAR(50),
    entity_id       UUID,
    created_by      UUID NOT NULL,
    is_private      BOOLEAN NOT NULL DEFAULT false,
    is_archived     BOOLEAN NOT NULL DEFAULT false,
    created_at      BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at      BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    deleted         BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT chk_chat_room_type CHECK (type IN ('direct', 'group', 'channel', 'team', 'project', 'ticket', 'custom'))
);

CREATE INDEX idx_chat_room_type ON chat_room(type);
CREATE INDEX idx_chat_room_entity ON chat_room(entity_type, entity_id);
CREATE INDEX idx_chat_room_created_by ON chat_room(created_by);
CREATE INDEX idx_chat_room_is_archived ON chat_room(is_archived);
CREATE INDEX idx_chat_room_created_at ON chat_room(created_at DESC);

-- ===========================
-- Table 3: chat_participant
-- ===========================
CREATE TABLE IF NOT EXISTS chat_participant (
    chat_room_id    UUID NOT NULL,
    user_id         UUID NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at       BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    is_muted        BOOLEAN NOT NULL DEFAULT false,
    last_read_at    BIGINT,

    PRIMARY KEY (chat_room_id, user_id),
    CONSTRAINT fk_chat_participant_room
        FOREIGN KEY (chat_room_id)
        REFERENCES chat_room(id)
        ON DELETE CASCADE,
    CONSTRAINT chk_participant_role CHECK (role IN ('owner', 'admin', 'moderator', 'member'))
);

CREATE INDEX idx_chat_participant_user_id ON chat_participant(user_id);
CREATE INDEX idx_chat_participant_role ON chat_participant(chat_room_id, role);

-- ===========================
-- Table 4: message
-- ===========================
CREATE TABLE IF NOT EXISTS message (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id        UUID NOT NULL,
    sender_id           UUID NOT NULL,
    type                VARCHAR(20) NOT NULL DEFAULT 'text',
    content             TEXT NOT NULL,
    content_format      VARCHAR(20) NOT NULL DEFAULT 'plain',
    metadata            JSONB,
    parent_id           UUID,
    is_edited           BOOLEAN NOT NULL DEFAULT false,
    is_pinned           BOOLEAN NOT NULL DEFAULT false,
    pinned_by           UUID,
    created_at          BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at          BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    deleted             BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT fk_message_room
        FOREIGN KEY (chat_room_id)
        REFERENCES chat_room(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_message_parent
        FOREIGN KEY (parent_id)
        REFERENCES message(id)
        ON DELETE SET NULL,
    CONSTRAINT chk_message_type CHECK (type IN ('text', 'quote', 'reply', 'system', 'file', 'image')),
    CONSTRAINT chk_content_format CHECK (content_format IN ('plain', 'markdown', 'html'))
);

CREATE INDEX idx_message_chat_room ON message(chat_room_id, created_at DESC);
CREATE INDEX idx_message_sender ON message(sender_id);
CREATE INDEX idx_message_parent ON message(parent_id);
CREATE INDEX idx_message_is_pinned ON message(chat_room_id, is_pinned) WHERE is_pinned = true;
CREATE INDEX idx_message_created_at ON message(created_at DESC);

-- Full-text search index for message content
CREATE INDEX idx_message_content_fts ON message USING gin(to_tsvector('english', content));

-- ===========================
-- Table 5: typing_indicator
-- ===========================
CREATE TABLE IF NOT EXISTS typing_indicator (
    chat_room_id    UUID NOT NULL,
    user_id         UUID NOT NULL,
    started_at      BIGINT NOT NULL,
    expires_at      BIGINT NOT NULL,

    PRIMARY KEY (chat_room_id, user_id),
    CONSTRAINT fk_typing_room
        FOREIGN KEY (chat_room_id)
        REFERENCES chat_room(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_typing_expires ON typing_indicator(expires_at);

-- ===========================
-- Table 6: message_read_receipt
-- ===========================
CREATE TABLE IF NOT EXISTS message_read_receipt (
    message_id      UUID NOT NULL,
    user_id         UUID NOT NULL,
    chat_room_id    UUID NOT NULL,
    read_at         BIGINT NOT NULL,

    PRIMARY KEY (message_id, user_id),
    CONSTRAINT fk_read_receipt_message
        FOREIGN KEY (message_id)
        REFERENCES message(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_read_receipt_room
        FOREIGN KEY (chat_room_id)
        REFERENCES chat_room(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_read_receipt_user ON message_read_receipt(user_id, chat_room_id);
CREATE INDEX idx_read_receipt_room ON message_read_receipt(chat_room_id);

-- ===========================
-- Table 7: message_attachment
-- ===========================
CREATE TABLE IF NOT EXISTS message_attachment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      UUID NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_size       BIGINT NOT NULL,
    file_type       VARCHAR(100) NOT NULL,
    file_url        TEXT NOT NULL,
    thumbnail_url   TEXT,
    uploaded_by     UUID NOT NULL,
    uploaded_at     BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,

    CONSTRAINT fk_attachment_message
        FOREIGN KEY (message_id)
        REFERENCES message(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_attachment_message ON message_attachment(message_id);
CREATE INDEX idx_attachment_uploaded_by ON message_attachment(uploaded_by);

-- ===========================
-- Table 8: message_reaction
-- ===========================
CREATE TABLE IF NOT EXISTS message_reaction (
    message_id      UUID NOT NULL,
    user_id         UUID NOT NULL,
    emoji           VARCHAR(50) NOT NULL,
    created_at      BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,

    PRIMARY KEY (message_id, user_id, emoji),
    CONSTRAINT fk_reaction_message
        FOREIGN KEY (message_id)
        REFERENCES message(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_reaction_message ON message_reaction(message_id);
CREATE INDEX idx_reaction_user ON message_reaction(user_id);

-- ===========================
-- Table 9: chat_external_integration
-- ===========================
CREATE TABLE IF NOT EXISTS chat_external_integration (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_room_id        UUID NOT NULL,
    provider            VARCHAR(50) NOT NULL,
    provider_room_id    VARCHAR(255) NOT NULL,
    webhook_url         TEXT,
    api_key             TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    last_sync_at        BIGINT,
    created_at          BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,
    updated_at          BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,

    CONSTRAINT fk_integration_room
        FOREIGN KEY (chat_room_id)
        REFERENCES chat_room(id)
        ON DELETE CASCADE,
    CONSTRAINT chk_provider CHECK (provider IN ('slack', 'telegram', 'whatsapp', 'yandex', 'google_chat'))
);

CREATE INDEX idx_integration_room ON chat_external_integration(chat_room_id);
CREATE INDEX idx_integration_provider ON chat_external_integration(provider);

-- ===========================
-- Comments for Documentation
-- ===========================

COMMENT ON TABLE user_presence IS 'Tracks user online/offline status and last seen timestamps';
COMMENT ON TABLE chat_room IS 'Chat rooms with support for multiple entity types (direct, group, team, project, ticket)';
COMMENT ON TABLE chat_participant IS 'Users participating in chat rooms with role-based permissions';
COMMENT ON TABLE message IS 'Messages with threading support (replies), pinning, and edit tracking';
COMMENT ON TABLE typing_indicator IS 'Real-time typing indicators with expiration';
COMMENT ON TABLE message_read_receipt IS 'Message read receipts for tracking who has read what';
COMMENT ON TABLE message_attachment IS 'File attachments linked to messages';
COMMENT ON TABLE message_reaction IS 'Emoji reactions to messages';
COMMENT ON TABLE chat_external_integration IS 'Integration with external chat providers (Slack, Telegram, etc.)';

-- ===========================
-- Grant Permissions
-- ===========================
-- Adjust based on your PostgreSQL user setup
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO chat_service;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO chat_service;

-- ===========================
-- Schema Version Tracking
-- ===========================
CREATE TABLE IF NOT EXISTS schema_version (
    version         INTEGER PRIMARY KEY,
    description     TEXT NOT NULL,
    applied_at      BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT
);

INSERT INTO schema_version (version, description) VALUES (0, 'Initial schema with 9 core tables');

-- End of initial schema
