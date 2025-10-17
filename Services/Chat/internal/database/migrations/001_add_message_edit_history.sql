-- Migration: Add Message Edit History Table
-- Purpose: Track complete history of message edits for transparency and audit
-- Author: Claude Code
-- Date: 2025-10-17

-- Create message_edit_history table
CREATE TABLE IF NOT EXISTS message_edit_history (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id              UUID NOT NULL,
    editor_id               UUID NOT NULL,
    previous_content        TEXT NOT NULL,
    previous_content_format VARCHAR(20) NOT NULL DEFAULT 'plain',
    previous_metadata       JSONB,
    edit_number             INTEGER NOT NULL DEFAULT 1,
    edited_at               BIGINT NOT NULL,
    created_at              BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())::BIGINT,

    -- Foreign key constraints
    CONSTRAINT fk_message_edit_history_message
        FOREIGN KEY (message_id)
        REFERENCES message(id)
        ON DELETE CASCADE,

    -- Indexes for performance
    CONSTRAINT unique_message_edit_number
        UNIQUE(message_id, edit_number)
);

-- Create indexes for efficient queries
CREATE INDEX idx_message_edit_history_message_id
    ON message_edit_history(message_id);

CREATE INDEX idx_message_edit_history_editor_id
    ON message_edit_history(editor_id);

CREATE INDEX idx_message_edit_history_edited_at
    ON message_edit_history(edited_at DESC);

-- Composite index for common query pattern (get edit history for a message)
CREATE INDEX idx_message_edit_history_message_edit
    ON message_edit_history(message_id, edit_number DESC);

-- Comments for documentation
COMMENT ON TABLE message_edit_history IS 'Complete history of all message edits for transparency and audit trail';
COMMENT ON COLUMN message_edit_history.id IS 'Unique identifier for the edit history record';
COMMENT ON COLUMN message_edit_history.message_id IS 'Reference to the message that was edited';
COMMENT ON COLUMN message_edit_history.editor_id IS 'User who made the edit (UUID from Core service)';
COMMENT ON COLUMN message_edit_history.previous_content IS 'Content before the edit was made';
COMMENT ON COLUMN message_edit_history.previous_content_format IS 'Format of the previous content (plain, markdown, html)';
COMMENT ON COLUMN message_edit_history.previous_metadata IS 'Previous metadata in JSONB format';
COMMENT ON COLUMN message_edit_history.edit_number IS 'Sequential number of this edit (1 = first edit, 2 = second edit, etc.)';
COMMENT ON COLUMN message_edit_history.edited_at IS 'Unix timestamp when the edit occurred';
COMMENT ON COLUMN message_edit_history.created_at IS 'Unix timestamp when this history record was created';

-- Grant permissions (adjust based on your database user setup)
-- GRANT SELECT, INSERT ON message_edit_history TO chat_service;
