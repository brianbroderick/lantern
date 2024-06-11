package repo

import (
	"time"

	"github.com/google/uuid"
)

type Table struct {
	UID                 uuid.UUID `json:"uid,omitempty"`                    // unique UUID of the table
	DatabaseUID         uuid.UUID `json:"database_uid,omitempty"`           // the UUID of the database
	Schema              string    `json:"schema_name,omitempty"`            // the schema of the table
	Name                string    `json:"table_name,omitempty"`             // the name of the table
	TableRows           int64     `json:"table_rows,omitempty"`             // the number of rows in the table
	TableColumns        int64     `json:"table_columns,omitempty"`          // the number of columns in the table
	TableIndexCount     int64     `json:"table_index_count,omitempty"`      // the number of indexes in the table
	TableIndexSizeBytes int64     `json:"table_index_size_bytes,omitempty"` // the size of the indexes in the table
	TableDataSizeBytes  int64     `json:"table_data_size_bytes,omitempty"`  // the size of the data in the table
	IsPhysicalTable     bool      `json:"is_physical_table,omitempty"`      // is this a physical table
	CreatedAt           time.Time `json:"created_at,omitempty"`             // when the table was created
	UpdatedAt           time.Time `json:"updated_at,omitempty"`             // when the table was last updated
}

func NewTable(databaseUID uuid.UUID, schema, name string, isPhysical bool) *Table {
	return &Table{
		UID:             UuidV5(schema + "." + name),
		DatabaseUID:     databaseUID,
		Schema:          schema,
		Name:            name,
		IsPhysicalTable: isPhysical,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
