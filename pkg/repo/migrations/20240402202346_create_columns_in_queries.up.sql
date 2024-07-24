CREATE TABLE IF NOT EXISTS columns_in_queries (
   uid UUID PRIMARY KEY NOT NULL,
   query_uid UUID NOT NULL, -- foreign key to queries table. 
   table_uid UUID, -- foreign key to tables table. The extractor doesn't match blank tables to the table list, therefore, this field can be NULL.
   column_uid UUID NOT NULL, -- foreign key to columns table.    
   schema_name TEXT NOT NULL DEFAULT 'public',
   table_name TEXT NOT NULL, 
   column_name TEXT NOT NULL,
   command TEXT NOT NULL, -- the command that was executed on the column i.e. SELECT, INSERT, UPDATE, DELETE
   clause TEXT NOT NULL
);

COMMENT ON COLUMN columns_in_queries.command IS 'the command that was executed on the column i.e. SELECT, INSERT, UPDATE, DELETE';
CREATE INDEX IF NOT EXISTS idx_columns_in_queries_column_uid ON columns_in_queries (query_uid, table_uid, column_uid);

