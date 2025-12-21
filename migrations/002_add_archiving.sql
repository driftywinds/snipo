-- Snipo Migration: Add Archiving
-- Version: 2

-- Add is_archived column to snippets
ALTER TABLE snippets ADD COLUMN is_archived INTEGER DEFAULT 0;

-- Create index for archived snippets
CREATE INDEX IF NOT EXISTS idx_snippets_archived ON snippets(is_archived);

-- Add archive_enabled column to settings
ALTER TABLE settings ADD COLUMN archive_enabled INTEGER DEFAULT 0;
