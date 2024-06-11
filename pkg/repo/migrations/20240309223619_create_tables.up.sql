CREATE TABLE IF NOT EXISTS tables (
   uid UUID PRIMARY KEY NOT NULL,
   database_uid UUID, -- foreign key to databases table. 
   schema_name TEXT NOT NULL DEFAULT 'public',
   table_name TEXT NOT NULL,   
   table_rows BIGINT NOT NULL DEFAULT 0,
   table_columns BIGINT NOT NULL DEFAULT 0,
   table_index_count BIGINT NOT NULL DEFAULT 0,
   table_index_size_bytes BIGINT NOT NULL DEFAULT 0,
   table_data_size_bytes BIGINT NOT NULL DEFAULT 0,
   is_physical_table BOOLEAN NOT NULL DEFAULT FALSE, -- if false, this may be a view or temporary table   
   -- is_partitioned BOOLEAN NOT NULL DEFAULT FALSE, -- if true, this table is partitioned
   -- partition_key TEXT, -- if is_partitioned is true, this is the column name used for partitioning
   -- partition_type TEXT, -- if is_partitioned is true, this is the type of partitioning
   -- partition_expression TEXT, -- if is_partitioned is true, this is the expression used for partitioning
   -- partition_count BIGINT NOT NULL DEFAULT 0, -- if is_partitioned is true, this is the number of partitions
   -- partition_size_bytes BIGINT NOT NULL DEFAULT 0, -- if is_partitioned is true, this is the size of each partition
   -- partition_index_count BIGINT NOT NULL DEFAULT 0, -- if is_partitioned is true, this is the number of indexes per partition
   -- partition_index_size_bytes BIGINT NOT NULL DEFAULT 0, -- if is_partitioned is true, this is the size of indexes per partition
   -- partition_data_size_bytes BIGINT NOT NULL DEFAULT 0, -- if is_partitioned is true, this is the size of data per partition
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()   
);

CREATE INDEX IF NOT EXISTS idx_tables_database_uid ON tables (database_uid);
CREATE INDEX IF NOT EXISTS idx_tables_table_name ON tables (table_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tables_schema_name_table_name ON tables (schema_name, table_name);
