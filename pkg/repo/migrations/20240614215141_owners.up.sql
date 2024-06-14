CREATE TABLE IF NOT EXISTS owners (
   uid UUID PRIMARY KEY NOT NULL,
   name TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_owners_name ON owners (name);