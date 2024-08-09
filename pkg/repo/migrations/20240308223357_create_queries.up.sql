CREATE TABLE IF NOT EXISTS queries (
  uid UUID PRIMARY KEY NOT NULL,  
  database_uid UUID, -- foreign key to databases table. Allow NULLs
  source_uid UUID, -- foreign key to sources table. Allow NULLs
  command TEXT NOT NULL,
  total_count BIGINT NOT NULL DEFAULT 0,
  total_duration_us BIGINT NOT NULL DEFAULT 0, -- in microseconds
  total_queries_in_transaction BIGINT NOT NULL DEFAULT 0, -- total number of queries in a single statement
  average_duration_us BIGINT NOT NULL DEFAULT 0, -- in microseconds
  average_queries_in_transaction NUMERIC(10,3) NOT NULL DEFAULT 0, -- average number of queries in a single statement
  masked_query TEXT NOT NULL DEFAULT '',   -- the query with params masked
  unmasked_query TEXT NOT NULL DEFAULT '', -- the query with params unmasked
  source_query TEXT NOT NULL DEFAULT ''   -- the original query as it was sent from the source
);

COMMENT ON COLUMN queries.total_duration_us IS 'us denotes that the duration is in microseconds';

CREATE INDEX IF NOT EXISTS idx_queries_database_uid ON queries (database_uid);
CREATE INDEX IF NOT EXISTS idx_queries_source_uid ON queries (source_uid);
CREATE INDEX IF NOT EXISTS idx_queries_command ON queries (command);
CREATE INDEX IF NOT EXISTS idx_queries_total_count ON queries (total_count);
CREATE INDEX IF NOT EXISTS idx_queries_total_duration ON queries (total_duration_us);