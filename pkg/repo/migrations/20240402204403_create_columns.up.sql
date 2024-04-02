CREATE TABLE IF NOT EXISTS columns (
   uid UUID PRIMARY KEY NOT NULL,
   table_uid UUID, -- foreign key to databases table. 
   schema_name TEXT NOT NULL DEFAULT 'public',
   table_name TEXT NOT NULL,
   column_name TEXT NOT NULL  
);

CREATE INDEX IF NOT EXISTS idx_columns_table_uid ON columns (table_uid);

