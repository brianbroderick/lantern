package repo

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Tables struct {
	Tables map[string]*Table `json:"tables,omitempty"`
}

func NewTables() *Tables {
	return &Tables{
		Tables: make(map[string]*Table),
	}
}

func (t *Tables) Add(tab *Table) *Table {
	fqtn := fmt.Sprintf("%s.%s", tab.Schema, tab.Name)

	if _, ok := t.Tables[fqtn]; !ok {
		t.Tables[fqtn] = tab
	}

	return t.Tables[fqtn]
}

func (t *Tables) CountInDB(db *sql.DB) int {
	var count int
	row := db.QueryRow("SELECT COUNT(1) FROM tables")
	row.Scan(&count)

	return count
}

func (t *Tables) Upsert(db *sql.DB) {
	if len(t.Tables) == 0 {
		return
	}

	rows := t.insValues()
	query := fmt.Sprintf(t.ins(), strings.Join(rows, ",\n"))

	ExecuteQuery(db, query)
}

func (t *Tables) ins() string {
	return `INSERT INTO tables (uid, database_uid, schema_name, table_name, estimated_row_count, data_size_bytes, table_type, updated_at)
	VALUES %s 
	ON CONFLICT (uid) DO UPDATE 
	SET uid = EXCLUDED.uid, 
	    database_uid = EXCLUDED.database_uid, 
			schema_name = EXCLUDED.schema_name, 
			table_name = EXCLUDED.table_name, 
			estimated_row_count = EXCLUDED.estimated_row_count, 
			data_size_bytes = EXCLUDED.data_size_bytes, 
			table_type = EXCLUDED.table_type, 
			updated_at = EXCLUDED.updated_at;`
}

func (t *Tables) insValues() []string {
	var rows []string

	for _, table := range t.Tables {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', '%s', '%d', '%d', '%s', '%s')",
				table.UID, table.DatabaseUID, table.Schema, table.Name, table.EstimatedRowCount, table.DataSizeBytes, table.TableType, table.UpdatedAt.Format(time.DateTime)))
	}
	return rows
}
