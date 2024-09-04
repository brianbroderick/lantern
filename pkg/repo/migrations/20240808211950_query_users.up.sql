CREATE TABLE IF NOT EXISTS query_users (
   uid UUID PRIMARY KEY NOT NULL,
   queries_by_hour_uid UUID NOT NULL, -- foreign key to queries_by_hour table. 
   user_name TEXT NOT NULL,
   total_count BIGINT NOT NULL DEFAULT 0,
   total_duration_us BIGINT NOT NULL DEFAULT 0 -- in microseconds
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_query_users_uniq ON query_users (queries_by_hour_uid, user_name);