-- migrations/003_add_upload_to_field.down.sql

-- Rollback: Remove upload_to column and its index
DROP INDEX IF EXISTS idx_files_upload_to;
ALTER TABLE files DROP COLUMN IF EXISTS upload_to;
