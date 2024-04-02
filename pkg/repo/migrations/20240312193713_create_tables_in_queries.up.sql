CREATE TABLE IF NOT EXISTS tables_in_queries (
   uid UUID PRIMARY KEY NOT NULL,
   query_uid UUID NOT NULL, -- foreign key to queries table. 
   table_uid UUID NOT NULL, -- foreign key to tables table.    
   schema_name TEXT NOT NULL DEFAULT 'public',
   table_name TEXT NOT NULL  
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tables_in_queries_tables_uid ON tables_in_queries (query_uid, table_uid);