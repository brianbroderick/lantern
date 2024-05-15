package repo

import (
	"fmt"
	"strings"
)

// tables are added in addTablesInQueries() function

func (q *Queries) UpsertTables() {
	if len(q.Tables) == 0 {
		return
	}

	rows := q.insValuesTables()
	query := fmt.Sprintf(q.insTables(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (q *Queries) insTables() string {
	return `INSERT INTO tables (uid, schema_name, table_name) 
	VALUES %s
	ON CONFLICT (uid) DO UPDATE 
	SET schema_name = EXCLUDED.schema_name, table_name = EXCLUDED.table_name;`
}

func (q *Queries) insValuesTables() []string {
	var rows []string

	for uid, table := range q.Tables {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s')",
				uid, table.Schema, table.Name))
	}
	return rows
}
