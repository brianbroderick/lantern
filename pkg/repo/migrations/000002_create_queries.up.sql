CREATE TABLE IF NOT EXISTS queries (
   uid UUID PRIMARY KEY NOT NULL,
   sha TEXT NOT NULL,
   command TEXT NOT NULL,
   total_count BIGINT NOT NULL,
   total_duration BIGINT NOT NULL,   
   masked_query TEXT NOT NULL,
   original_query TEXT NOT NULL
);