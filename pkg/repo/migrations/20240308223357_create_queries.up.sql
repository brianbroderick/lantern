CREATE TABLE IF NOT EXISTS queries (
   uid UUID PRIMARY KEY NOT NULL,  
   database_uid UUID, -- foreign key to databases table. 
   command TEXT NOT NULL,
   total_count BIGINT NOT NULL,
   total_duration BIGINT NOT NULL,   
   masked_query TEXT NOT NULL,
   original_query TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_queries_database_uid ON queries (database_uid);
CREATE INDEX IF NOT EXISTS idx_queries_command ON queries (command);
CREATE INDEX IF NOT EXISTS idx_queries_total_count ON queries (total_count);
CREATE INDEX IF NOT EXISTS idx_queries_total_duration ON queries (total_duration);