-- migrations/002_sync_tokens.down.sql

DROP TRIGGER IF EXISTS update_sync_tokens_updated_at ON sync_tokens;
DROP INDEX IF EXISTS idx_sync_tokens_created_by;
DROP INDEX IF EXISTS idx_sync_tokens_expires_at;
DROP INDEX IF EXISTS idx_sync_tokens_is_active;
DROP INDEX IF EXISTS idx_sync_tokens_token_hash;
DROP INDEX IF EXISTS idx_sync_tokens_tenant_id;
DROP TABLE IF EXISTS sync_tokens;