CREATE TABLE IF NOT EXISTS tables_in_queries (
   uid UUID PRIMARY KEY NOT NULL,
   tables_uid UUID NOT NULL, -- foreign key to tables table. 
   queries_uid UUID NOT NULL -- foreign key to queries table. 
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tables_in_queries_tables_uid ON tables_in_queries (tables_uid, queries_uid);