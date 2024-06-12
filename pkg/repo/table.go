package repo

import (
	"time"

	"github.com/google/uuid"
)

type Table struct {
	UID               uuid.UUID `json:"uid,omitempty"`                 // unique UUID of the table
	DatabaseUID       uuid.UUID `json:"database_uid,omitempty"`        // the UUID of the database
	Schema            string    `json:"schema_name,omitempty"`         // the schema of the table
	Name              string    `json:"table_name,omitempty"`          // the name of the table
	Description       string    `json:"table_description,omitempty"`   // a description of the table
	EstimatedRowCount int64     `json:"estimated_row_count,omitempty"` // the estimated number of rows in the table
	ColumnCount       int64     `json:"column_count,omitempty"`        // the number of columns in the table
	IndexCount        int64     `json:"index_count,omitempty"`         // the number of indexes on the table
	IndexSizeBytes    int64     `json:"index_size_bytes,omitempty"`    // the size of the indexes on the table
	DataSizeBytes     int64     `json:"data_size_bytes,omitempty"`     // the size of the data in the table
	TableType         string    `json:"table_type,omitempty"`          // the type of table (e.g. view, table, materialized view)
	CreatedAt         time.Time `json:"created_at,omitempty"`          // when the table was created
	UpdatedAt         time.Time `json:"updated_at,omitempty"`          // when the table was last updated
}

func (t *Table) SetUID() {
	t.UID = UuidV5(t.Schema + "." + t.Name)
}
