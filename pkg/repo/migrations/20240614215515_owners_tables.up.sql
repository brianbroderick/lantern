CREATE TABLE IF NOT EXISTS owners_tables (
   uid UUID PRIMARY KEY NOT NULL,
   owner_uid UUID, -- foreign key to owners table.
   table_uid UUID  -- foreign key to databases table.
);

CREATE INDEX IF NOT EXISTS idx_owners_tables_join ON owners_tables (owner_uid, table_uid);