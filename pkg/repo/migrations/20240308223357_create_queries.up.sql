CREATE TABLE IF NOT EXISTS queries (
   uid UUID PRIMARY KEY NOT NULL,  
   database_uid UUID, -- foreign key to databases table. Allow NULLs
   source_uid UUID, -- foreign key to sources table. Allow NULLs
   command TEXT NOT NULL,
   total_count BIGINT NOT NULL,
   total_duration_us BIGINT NOT NULL, -- in microseconds
   total_queries_in_transaction BIGINT NOT NULL,
   average_duration_us BIGINT NOT NULL, -- in microseconds
   average_queries_in_transaction NUMERIC(10,3) NOT NULL,
   masked_query TEXT NOT NULL,
   unmasked_query TEXT NOT NULL,
   source_query TEXT NOT NULL
);

COMMENT ON COLUMN queries.total_duration_us IS 'us denotes that the duration is in microseconds';

CREATE INDEX IF NOT EXISTS idx_queries_database_uid ON queries (database_uid);
CREATE INDEX IF NOT EXISTS idx_queries_source_uid ON queries (source_uid);
CREATE INDEX IF NOT EXISTS idx_queries_command ON queries (command);
CREATE INDEX IF NOT EXISTS idx_queries_total_count ON queries (total_count);
CREATE INDEX IF NOT EXISTS idx_queries_total_duration ON queries (total_duration_us);