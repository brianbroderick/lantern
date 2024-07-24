CREATE TABLE IF NOT EXISTS tables_in_queries (
   uid UUID PRIMARY KEY NOT NULL,
   query_uid UUID NOT NULL, -- foreign key to queries table. 
   table_uid UUID NOT NULL, -- foreign key to tables table.    
   command TEXT NOT NULL, -- the command that was executed on the table i.e. SELECT, INSERT, UPDATE, DELETE
   schema_name TEXT NOT NULL DEFAULT 'public',
   table_name TEXT NOT NULL  
);

COMMENT ON COLUMN tables_in_queries.command IS 'the command that was executed on the table i.e. SELECT, INSERT, UPDATE, DELETE';

CREATE UNIQUE INDEX IF NOT EXISTS idx_tables_in_queries_join ON tables_in_queries (query_uid, table_uid, command);