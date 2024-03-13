CREATE TABLE IF NOT EXISTS tables (
   uid UUID PRIMARY KEY NOT NULL,
   database_uid UUID, -- foreign key to databases table. 
   table_name TEXT NOT NULL,
   total_count BIGINT NOT NULL,
   total_duration BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tables_database_uid ON tables (database_uid);
CREATE INDEX IF NOT EXISTS idx_tables_table_name ON tables (table_name);
CREATE INDEX IF NOT EXISTS idx_tables_total_count ON tables (total_count);
CREATE INDEX IF NOT EXISTS idx_tables_total_duration ON tables (total_duration);

