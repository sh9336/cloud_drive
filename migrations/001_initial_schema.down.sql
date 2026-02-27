-- migrations/001_initial_schema.down.sql

DROP TRIGGER IF EXISTS update_files_updated_at ON files;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TRIGGER IF EXISTS update_admins_updated_at ON admins;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_actor;
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user;
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
DROP INDEX IF EXISTS idx_files_deleted_at;
DROP INDEX IF EXISTS idx_files_created_at;
DROP INDEX IF EXISTS idx_files_upload_status;
DROP INDEX IF EXISTS idx_files_tenant_id;
DROP INDEX IF EXISTS idx_tenants_created_by;
DROP INDEX IF EXISTS idx_tenants_is_active;
DROP INDEX IF EXISTS idx_tenants_email;

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS admins;

DROP EXTENSION IF EXISTS "uuid-ossp";