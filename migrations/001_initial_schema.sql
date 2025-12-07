-- Snipo Initial Schema
-- Version: 1

-- Snippets table - core entity
CREATE TABLE IF NOT EXISTS snippets (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(8)))),
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    content TEXT NOT NULL,
    language TEXT DEFAULT 'plaintext',
    is_favorite INTEGER DEFAULT 0,
    is_public INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    s3_key TEXT DEFAULT NULL,
    checksum TEXT DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    color TEXT DEFAULT '#6366f1',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Snippet-Tag junction table
CREATE TABLE IF NOT EXISTS snippet_tags (
    snippet_id TEXT NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (snippet_id, tag_id),
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Folders/Collections for organization
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parent_id INTEGER DEFAULT NULL,
    icon TEXT DEFAULT 'folder',
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Snippet-Folder relationship
CREATE TABLE IF NOT EXISTS snippet_folders (
    snippet_id TEXT NOT NULL,
    folder_id INTEGER NOT NULL,
    PRIMARY KEY (snippet_id, folder_id),
    FOREIGN KEY (snippet_id) REFERENCES snippets(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Application settings (single row)
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    app_name TEXT DEFAULT 'snipo',
    custom_css TEXT DEFAULT '',
    theme TEXT DEFAULT 'auto',
    default_language TEXT DEFAULT 'plaintext',
    s3_enabled INTEGER DEFAULT 0,
    s3_endpoint TEXT DEFAULT '',
    s3_bucket TEXT DEFAULT '',
    s3_region TEXT DEFAULT 'us-east-1',
    backup_encryption_enabled INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- API tokens for external access
CREATE TABLE IF NOT EXISTS api_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    token_hash TEXT UNIQUE NOT NULL,
    permissions TEXT DEFAULT 'read',
    last_used_at DATETIME DEFAULT NULL,
    expires_at DATETIME DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table for web authentication
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    token_hash TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_snippets_language ON snippets(language);
CREATE INDEX IF NOT EXISTS idx_snippets_favorite ON snippets(is_favorite);
CREATE INDEX IF NOT EXISTS idx_snippets_public ON snippets(is_public);
CREATE INDEX IF NOT EXISTS idx_snippets_created ON snippets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_snippets_updated ON snippets(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Full-text search (external content FTS5 table)
CREATE VIRTUAL TABLE IF NOT EXISTS snippets_fts USING fts5(
    snippet_id,
    title,
    description,
    content,
    content='snippets',
    content_rowid='rowid'
);

-- Triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS snippets_ai AFTER INSERT ON snippets BEGIN
    INSERT INTO snippets_fts(rowid, snippet_id, title, description, content)
    VALUES (NEW.rowid, NEW.id, NEW.title, NEW.description, NEW.content);
END;

CREATE TRIGGER IF NOT EXISTS snippets_ad AFTER DELETE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, snippet_id, title, description, content)
    VALUES('delete', OLD.rowid, OLD.id, OLD.title, OLD.description, OLD.content);
END;

CREATE TRIGGER IF NOT EXISTS snippets_au AFTER UPDATE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, snippet_id, title, description, content)
    VALUES('delete', OLD.rowid, OLD.id, OLD.title, OLD.description, OLD.content);
    INSERT INTO snippets_fts(rowid, snippet_id, title, description, content)
    VALUES (NEW.rowid, NEW.id, NEW.title, NEW.description, NEW.content);
END;

-- Insert default settings row
INSERT OR IGNORE INTO settings (id) VALUES (1);
