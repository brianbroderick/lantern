package repo

import (
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
)

// addTablesInQueries adds the tables in a query to the Queries struct
func (q *Queries) addTablesInQueries(qu *Query, ext *extractor.Extractor) {

	for _, table := range ext.TablesInQueries {
		uid := UuidV5(fmt.Sprintf("%s|%s", qu.UID, table.TableUID))
		uidStr := uid.String()
		if _, ok := q.TablesInQueries[uidStr]; !ok {
			q.TablesInQueries[uidStr] = &extractor.TablesInQueries{
				UID:      uid,
				TableUID: table.TableUID,
				QueryUID: qu.UID,
				Schema:   table.Schema,
				Name:     table.Name,
				Command:  table.Command,
			}
		}

		tableUIDStr := table.TableUID.String()
		if _, ok := q.Tables[tableUIDStr]; !ok {
			q.Tables[tableUIDStr] = &extractor.Tables{
				UID:    table.TableUID,
				Schema: table.Schema,
				Name:   table.Name,
			}
		}
	}
}

func (q *Queries) UpsertTablesInQueries() {
	if len(q.TablesInQueries) == 0 {
		return
	}

	rows := q.insValuesTablesInQueries()
	query := fmt.Sprintf(q.insTablesInQueries(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (q *Queries) insTablesInQueries() string {
	return `INSERT INTO tables_in_queries (uid, table_uid, query_uid, schema_name, table_name, command) 
	VALUES %s 
	ON CONFLICT (uid) DO UPDATE 
	SET table_uid = EXCLUDED.table_uid, 
		query_uid = EXCLUDED.query_uid, 
		schema_name = EXCLUDED.schema_name, 
		table_name = EXCLUDED.table_name,
		command = EXCLUDED.command;`
}

func (q *Queries) insValuesTablesInQueries() []string {
	var rows []string

	for uid, query := range q.TablesInQueries {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s')",
				uid, query.TableUID, query.QueryUID, query.Schema, query.Name, query.Command))
	}
	return rows
}
