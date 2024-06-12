package repo

import (
	"fmt"
	"strings"
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
	return `INSERT INTO tables (uid, schema_name, table_name, table_type) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (t *Tables) insValues() []string {
	var rows []string

	for uid, table := range t.Tables {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', %s)",
				uid, table.Schema, table.Name, table.TableType))
	}
	return rows
}
