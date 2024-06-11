package repo

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Tables struct {
	Tables map[string]*Table `json:"tables,omitempty"`
}

func NewTables() *Tables {
	return &Tables{
		Tables: make(map[string]*Table),
	}
}

func (t *Tables) Add(databaseUID uuid.UUID, schema, name string, isPhyscial bool) *Table {
	if _, ok := t.Tables[fmt.Sprintf("%s.%s", schema, name)]; !ok {
		t.Tables[name] = NewTable(databaseUID, schema, name, isPhyscial)
	}

	return t.Tables[name]
}

func (t *Tables) CountInDB() int {
	db := Conn()
	defer db.Close()

	var count int
	row := db.QueryRow("SELECT COUNT(1) FROM tables")
	row.Scan(&count)

	return count
}

func (t *Tables) Upsert() {
	if len(t.Tables) == 0 {
		return
	}

	rows := t.insValues()
	query := fmt.Sprintf(t.ins(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (t *Tables) ins() string {
	return `INSERT INTO tables (uid, schema_name, table_name, is_physical) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (t *Tables) insValues() []string {
	var rows []string

	for uid, table := range t.Tables {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', %t)",
				uid, table.Schema, table.Name, table.IsPhysicalTable))
	}
	return rows
}
