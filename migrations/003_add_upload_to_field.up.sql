-- migrations/003_add_upload_to_field.up.sql

-- Add upload_to column to files table to track which folder the file was uploaded to
ALTER TABLE files ADD COLUMN upload_to VARCHAR(255) DEFAULT 'uploads' NOT NULL;

-- Create index for queries by upload_to
CREATE INDEX idx_files_upload_to ON files(tenant_id, upload_to);
