-- migrations/002_sync_tokens.up.sql

CREATE TABLE sync_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    
    -- Permissions
    can_read BOOLEAN DEFAULT TRUE,
    can_write BOOLEAN DEFAULT TRUE,
    can_delete BOOLEAN DEFAULT FALSE,
    
    -- Lifecycle
    expires_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    revoked_at TIMESTAMP,
    revoked_by UUID REFERENCES admins(id),
    revoked_reason TEXT,
    
    -- Metadata
    created_by UUID REFERENCES admins(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Stats (for monitoring)
    total_requests BIGINT DEFAULT 0,
    total_bytes_uploaded BIGINT DEFAULT 0,
    total_bytes_downloaded BIGINT DEFAULT 0
);

-- Indexes for performance
CREATE INDEX idx_sync_tokens_tenant_id ON sync_tokens(tenant_id);
CREATE INDEX idx_sync_tokens_token_hash ON sync_tokens(token_hash);
CREATE INDEX idx_sync_tokens_is_active ON sync_tokens(is_active);
CREATE INDEX idx_sync_tokens_expires_at ON sync_tokens(expires_at);
CREATE INDEX idx_sync_tokens_created_by ON sync_tokens(created_by);

-- Trigger to update updated_at
CREATE TRIGGER update_sync_tokens_updated_at BEFORE UPDATE ON sync_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();