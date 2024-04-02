CREATE TABLE IF NOT EXISTS columns_in_queries (
   uid UUID PRIMARY KEY NOT NULL,
   query_uid UUID NOT NULL, -- foreign key to queries table. 
   table_uid UUID NOT NULL, -- foreign key to tables table.
   column_uid UUID NOT NULL, -- foreign key to columns table. 
   schema_name TEXT NOT NULL DEFAULT 'public',
   table_name TEXT NOT NULL, 
   column_name TEXT NOT NULL,
   clause TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_columns_in_queries_column_uid ON columns_in_queries (query_uid, table_uid, column_uid);
