-- 003_embeddings.sql: Embedding column + HNSW index for semantic search

-- The pgvector extension was already created in 001_initial.sql

-- Add embedding column to events table partitions
-- (Adding to the parent table propagates to partitions)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'events' AND column_name = 'embedding'
    ) THEN
        ALTER TABLE events ADD COLUMN embedding vector(384);
    END IF;
END $$;

-- HNSW index on the embedding column for fast approximate nearest-neighbor search
-- Using cosine distance (appropriate for normalized text embeddings)
CREATE INDEX IF NOT EXISTS idx_events_embedding_hnsw
    ON events USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- AI conversation history
CREATE TABLE IF NOT EXISTS ai_conversations (
    conversation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        REFERENCES users(user_id) ON DELETE CASCADE,
    title           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_messages (
    message_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID        NOT NULL REFERENCES ai_conversations(conversation_id) ON DELETE CASCADE,
    role            TEXT        NOT NULL,
    content         TEXT        NOT NULL,
    tool_calls      JSONB,
    tool_call_id    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation
    ON ai_messages (conversation_id, created_at);
