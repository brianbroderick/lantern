CREATE TABLE IF NOT EXISTS tables (
   uid UUID PRIMARY KEY NOT NULL default uuid_generate_v1mc(),
   table_name TEXT NOT NULL
);