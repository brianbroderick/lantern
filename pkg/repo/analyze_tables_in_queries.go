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
	return `INSERT INTO tables_in_queries (uid, table_uid, query_uid, schema_name, table_name) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (q *Queries) insValuesTablesInQueries() []string {
	var rows []string

	for uid, query := range q.TablesInQueries {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', '%s', '%s')",
				uid, query.TableUID, query.QueryUID, query.Schema, query.Name))
	}
	return rows
}
