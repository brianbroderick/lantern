CREATE TABLE IF NOT EXISTS query_stats (
  uid UUID PRIMARY KEY NOT NULL,  
  query_uid UUID, -- foreign key to queries table. Allow NULLs
  queried_date DATE NOT NULL DEFAULT NOW(), -- eventually build a date dimension table
  queried_hour INT NOT NULL DEFAULT 0,
  total_count BIGINT NOT NULL DEFAULT 0,
  total_duration_us BIGINT NOT NULL DEFAULT 0, -- in microseconds
  total_queries_in_transaction BIGINT NOT NULL DEFAULT 0, -- total number of queries in a single statement
  average_duration_us BIGINT NOT NULL DEFAULT 0, -- in microseconds
  average_queries_in_transaction NUMERIC(10,3) NOT NULL DEFAULT 0 -- average number of queries in a single statement  
);

COMMENT ON COLUMN query_stats.total_duration_us IS 'us denotes that the duration is in microseconds';

CREATE UNIQUE INDEX IF NOT EXISTS idx_query_stats_unique ON query_stats (query_uid, queried_date, queried_hour);
-- Include this when we have reports for specific hours
-- CREATE INDEX IF NOT EXISTS idx_query_stats_queried_hour ON query_stats (queried_hour);
CREATE INDEX IF NOT EXISTS idx_query_stats_queried_datetime ON query_stats (queried_date, queried_hour);
CREATE INDEX IF NOT EXISTS idx_query_stats_total_count ON query_stats (total_count);
CREATE INDEX IF NOT EXISTS idx_query_stats_total_duration ON query_stats (total_duration_us);