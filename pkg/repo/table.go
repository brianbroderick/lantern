package repo

import (
	"encoding/json"
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

func (t *Table) MarshalJSON() ([]byte, error) {
	type Alias Table

	var (
		upd string
		cre string
	)

	if !t.UpdatedAt.IsZero() {
		upd = t.UpdatedAt.Format(time.DateTime)
	}

	if !t.CreatedAt.IsZero() {
		cre = t.CreatedAt.Format(time.DateTime)
	}

	s := &struct {
		*Alias
		UpdatedAt string `json:"updated_at,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
	}{
		Alias:     (*Alias)(t),
		UpdatedAt: upd,
		CreatedAt: cre,
	}

	b, err := json.MarshalIndent(s, "", "  ")
	if HasErr("marshallJSON", err) {
		return []byte{}, err
	}
	return b, err
}

func (t *Table) UnmarshalJSON(data []byte) error {
	type Alias Table

	s := &struct {
		*Alias
		UpdatedAt string `json:"updated_at,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s.UpdatedAt != "" {
		t.UpdatedAt, _ = time.Parse(time.DateTime, s.UpdatedAt)
	}

	if s.CreatedAt != "" {
		t.CreatedAt, _ = time.Parse(time.DateTime, s.CreatedAt)
	}

	return nil
}
