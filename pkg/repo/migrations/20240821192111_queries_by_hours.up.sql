CREATE TABLE IF NOT EXISTS queries_by_hours (
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

COMMENT ON COLUMN queries_by_hours.total_duration_us IS 'us denotes that the duration is in microseconds';

CREATE UNIQUE INDEX IF NOT EXISTS idx_queries_by_hours_unique ON queries_by_hours (query_uid, queried_date, queried_hour);
-- Include this when we have reports for specific hours
-- CREATE INDEX IF NOT EXISTS idx_queries_by_hours_queried_hour ON queries_by_hours (queried_hour);
CREATE INDEX IF NOT EXISTS idx_queries_by_hours_queried_datetime ON queries_by_hours (queried_date, queried_hour);
CREATE INDEX IF NOT EXISTS idx_queries_by_hours_total_count ON queries_by_hours (total_count);
CREATE INDEX IF NOT EXISTS idx_queries_by_hours_total_duration ON queries_by_hours (total_duration_us);