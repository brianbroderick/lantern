CREATE TABLE IF NOT EXISTS databases (
   uid UUID PRIMARY KEY NOT NULL,   
   name TEXT NOT NULL,   
   template TEXT DEFAULT '' NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_databases_name ON databases (name);
CREATE INDEX IF NOT EXISTS idx_databases_template ON databases (template);