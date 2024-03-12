CREATE TABLE IF NOT EXISTS queries (
   uid UUID PRIMARY KEY NOT NULL,  
   database_uid UUID, -- foreign key to databases table
   command TEXT NOT NULL,
   total_count BIGINT NOT NULL,
   total_duration BIGINT NOT NULL,   
   masked_query TEXT NOT NULL,
   original_query TEXT NOT NULL
);