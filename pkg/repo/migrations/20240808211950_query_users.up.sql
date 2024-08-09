CREATE TABLE IF NOT EXISTS query_users (
   uid UUID PRIMARY KEY NOT NULL,
   query_uid UUID NOT NULL, -- foreign key to queries table. 
   user_name TEXT NOT NULL,
   total_count BIGINT NOT NULL DEFAULT 0,
   total_duration_us BIGINT NOT NULL DEFAULT 0 -- in microseconds
);