CREATE TABLE IF NOT EXISTS create_statements_in_queries (
   uid UUID PRIMARY KEY NOT NULL,
   query_uid UUID NOT NULL, -- foreign key to queries table. 
   create_statement_uid UUID NOT NULL -- foreign key to create_statements table.       
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_create_statements_in_queries_join ON create_statements_in_queries (query_uid, create_statement_uid);