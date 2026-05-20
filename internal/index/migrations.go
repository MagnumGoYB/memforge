package index

const CurrentSchemaVersion = "1"

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS schema_version (version TEXT NOT NULL);`,
	`CREATE TABLE IF NOT EXISTS memories (
	  id TEXT PRIMARY KEY,
	  project_id TEXT NOT NULL,
	  kind TEXT NOT NULL,
	  title TEXT NOT NULL,
	  content TEXT NOT NULL,
	  tags_json TEXT NOT NULL DEFAULT '[]',
	  tags_flat TEXT NOT NULL DEFAULT '',
	  source TEXT,
	  confidence REAL NOT NULL DEFAULT 1.0,
	  usage_count INTEGER NOT NULL DEFAULT 0,
	  created_at TEXT NOT NULL,
	  updated_at TEXT NOT NULL
	);`,
	`CREATE INDEX IF NOT EXISTS memories_kind_idx ON memories(kind);`,
	`CREATE INDEX IF NOT EXISTS memories_updated_idx ON memories(updated_at);`,
	`CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
	  title,
	  content,
	  tags_flat,
	  content='memories',
	  content_rowid='rowid',
	  tokenize='unicode61 remove_diacritics 2'
	);`,
	`CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
	  INSERT INTO memories_fts(rowid, title, content, tags_flat)
	  VALUES (new.rowid, new.title, new.content, new.tags_flat);
	END;`,
	`CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
	  INSERT INTO memories_fts(memories_fts, rowid, title, content, tags_flat)
	  VALUES ('delete', old.rowid, old.title, old.content, old.tags_flat);
	END;`,
	`CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
	  INSERT INTO memories_fts(memories_fts, rowid, title, content, tags_flat)
	  VALUES ('delete', old.rowid, old.title, old.content, old.tags_flat);
	  INSERT INTO memories_fts(rowid, title, content, tags_flat)
	  VALUES (new.rowid, new.title, new.content, new.tags_flat);
	END;`,
}
