CREATE TABLE IF NOT EXISTS tables (
   uid UUID PRIMARY KEY NOT NULL,
   database_uid UUID, -- foreign key to databases table. 
   schema_name TEXT NOT NULL DEFAULT 'public',
   table_name TEXT NOT NULL  
);

CREATE INDEX IF NOT EXISTS idx_tables_database_uid ON tables (database_uid);
CREATE INDEX IF NOT EXISTS idx_tables_table_name ON tables (table_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tables_schema_name_table_name ON tables (schema_name, table_name);
