CREATE TABLE IF NOT EXISTS table_joins_in_queries (
   uid UUID PRIMARY KEY NOT NULL, -- the uid is calculated from the hash of other columns
   query_uid UUID NOT NULL, -- foreign key to queries table. 
   table_uid_a UUID NOT NULL, -- foreign key to tables table. 
   table_uid_b UUID NOT NULL, -- foreign key to queries table. 
   join_condition TEXT NOT NULL, -- INNER, LEFT, RIGHT, OUTER, etc.
   on_condition TEXT NOT NULL, -- for now, we'll just store the entire ON condition as a string.
   table_a TEXT NOT NULL, -- the name for table_a
   table_b TEXT NOT NULL -- the name for table_b
);

-- There may be many of these if the same tables are joined in multiple places.
CREATE INDEX IF NOT EXISTS idx_table_joins_in_queries_uid ON table_joins_in_queries (table_uid_a, table_uid_b);
CREATE INDEX IF NOT EXISTS idx_table_joins_in_queries_queries_uid ON table_joins_in_queries (query_uid);
