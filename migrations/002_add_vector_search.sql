-- Add vector search support for AI-powered review search
-- This migration adds pgvector extension and embedding columns

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Add embedding column to reviews table
-- Using 1536 dimensions for OpenAI text-embedding-3-small model
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS embedding vector(1536);

-- Add embedding status tracking
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS embedding_status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS embedding_updated_at TIMESTAMP WITH TIME ZONE;

-- Create index for vector similarity search
-- Using HNSW index for fast approximate nearest neighbor search
CREATE INDEX IF NOT EXISTS idx_reviews_embedding_hnsw 
ON reviews USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Create index for embedding status filtering
CREATE INDEX IF NOT EXISTS idx_reviews_embedding_status 
ON reviews(embedding_status) 
WHERE embedding_status != 'completed';

-- Add function to update embedding status
CREATE OR REPLACE FUNCTION update_embedding_status()
RETURNS TRIGGER AS $$
BEGIN
    NEW.embedding_updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add trigger for embedding status updates
CREATE TRIGGER update_reviews_embedding_status 
BEFORE UPDATE OF embedding, embedding_status ON reviews
FOR EACH ROW EXECUTE FUNCTION update_embedding_status();

-- Add comments for documentation
COMMENT ON COLUMN reviews.embedding IS 'Vector embedding for semantic search (1536 dimensions)';
COMMENT ON COLUMN reviews.embedding_status IS 'Status of embedding generation: pending, processing, completed, failed';
COMMENT ON COLUMN reviews.embedding_updated_at IS 'Timestamp when embedding was last updated';
