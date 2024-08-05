CREATE TABLE IF NOT EXISTS create_statements (
  uid UUID PRIMARY KEY NOT NULL,  
  scope TEXT NOT NULL DEFAULT '',                   -- GLOBAL or LOCAL
  is_unique BOOLEAN NOT NULL DEFAULT false,         -- UNIQUE
  used_concurrently BOOLEAN NOT NULL DEFAULT false, -- CONCURRENTLY phrase was used
  is_temp BOOLEAN NOT NULL DEFAULT false,           -- TEMP or TEMPORARY (same thing)
  is_unlogged BOOLEAN NOT NULL DEFAULT false,       -- UNLOGGED
  object_type TEXT NOT NULL DEFAULT '',             -- TABLE, INDEX, VIEW, etc.
  if_not_exists BOOLEAN NOT NULL DEFAULT false,     -- IF NOT EXISTS phrase was used
  name TEXT NOT NULL DEFAULT '',                    -- the name of the object
  on_commit TEXT NOT NULL DEFAULT '',               -- PRESERVE ROWS, DELETE ROWS, DROP
  operator TEXT NOT NULL DEFAULT '',                -- AS (for CREATE TABLE AS), ON for CREATE INDEX ON, etc.
  expression TEXT NOT NULL DEFAULT '',              -- the expression to create the object
  where_clause TEXT NOT NULL DEFAULT ''             -- the WHERE clause for the object
);

CREATE INDEX IF NOT EXISTS idx_create_statements_name ON create_statements (name);
COMMENT ON COLUMN create_statements.scope IS 'GLOBAL or LOCAL';
COMMENT ON COLUMN create_statements.is_unique IS 'UNIQUE';
COMMENT ON COLUMN create_statements.used_concurrently IS 'CONCURRENTLY phrase was used';
COMMENT ON COLUMN create_statements.is_temp IS 'TEMP or TEMPORARY (same thing)';
COMMENT ON COLUMN create_statements.is_unlogged IS 'UNLOGGED';
COMMENT ON COLUMN create_statements.object_type IS 'TABLE, INDEX, VIEW, etc.';
COMMENT ON COLUMN create_statements.if_not_exists IS 'IF NOT EXISTS phrase was used';
COMMENT ON COLUMN create_statements.name IS 'the name of the object';
COMMENT ON COLUMN create_statements.on_commit IS 'PRESERVE ROWS, DELETE ROWS, DROP';
COMMENT ON COLUMN create_statements.operator IS 'AS (for CREATE TABLE AS), ON for CREATE INDEX ON, etc.';
COMMENT ON COLUMN create_statements.expression IS 'the expression to create the object';
COMMENT ON COLUMN create_statements.where_clause IS 'the WHERE clause for the object';
