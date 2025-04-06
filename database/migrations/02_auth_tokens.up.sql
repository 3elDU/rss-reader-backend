CREATE TABLE auth_tokens (
  id INTEGER PRIMARY KEY ASC,
  token TEXT NOT NULL,
  created_at TEXT NOT NULL,
  valid_until TEXT
);