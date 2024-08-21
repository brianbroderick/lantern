CREATE TABLE IF NOT EXISTS processed_files (
  uid UUID PRIMARY KEY NOT NULL,
  file_name TEXT NOT NULL,
  processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_processed_files_file_name ON processed_files (file_name);