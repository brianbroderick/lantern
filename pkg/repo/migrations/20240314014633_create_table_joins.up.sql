CREATE TABLE IF NOT EXISTS table_joins (
   uid UUID PRIMARY KEY NOT NULL, -- the uid is calculated from the hash of other columns
   queries_uid UUID NOT NULL, -- foreign key to queries table. 
   tables_uid_a UUID NOT NULL, -- foreign key to tables table. 
   tables_uid_b UUID NOT NULL, -- foreign key to queries table. 
   join_condition TEXT NOT NULL, -- INNER, LEFT, RIGHT, OUTER, etc.
   on_condition TEXT NOT NULL -- for now, we'll just store the entire ON condition as a string.
);

-- There may be many of these if the same tables are joined in multiple places.
CREATE INDEX IF NOT EXISTS idx_table_joins_uid ON table_joins (tables_uid_a, tables_uid_b);
CREATE INDEX IF NOT EXISTS idx_table_joins_queries_uid ON table_joins (queries_uid);
