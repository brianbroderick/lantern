package repo

import (
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
)

// 14584 in load
// addColumnsInQueries adds the columns in a query the Queries struct
func (q *Queries) addColumnsInQueries(qu *Query, ext *extractor.Extractor) {
	for _, column := range ext.ColumnsInQueries {
		uid := UuidV5(fmt.Sprintf("%s|%s|%s.%s.%s", qu.UID, column.Clause.String(), column.Schema, column.Table, column.Name))
		uidStr := uid.String()
		if _, ok := q.ColumnsInQueries[uidStr]; !ok {
			q.ColumnsInQueries[uidStr] = &extractor.ColumnsInQueries{
				UID:       uid,
				TableUID:  column.TableUID,
				ColumnUID: column.ColumnUID,
				QueryUID:  qu.UID,
				Schema:    column.Schema,
				Table:     column.Table,
				Name:      column.Name,
				Clause:    column.Clause,
			}
		}
	}
}

func (q *Queries) UpsertColumnsInQueries() {
	if len(q.ColumnsInQueries) == 0 {
		return
	}

	rows := q.insValuesColumnsInQueries()
	query := fmt.Sprintf(q.insColumnsInQueries(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (q *Queries) insColumnsInQueries() string {
	return `INSERT INTO columns_in_queries (uid, query_uid, table_uid, column_uid, schema_name, table_name, column_name, clause)
	VALUES %s 
	ON CONFLICT (uid) DO UPDATE 
	SET query_uid = EXCLUDED.query_uid, table_uid = EXCLUDED.table_uid, column_uid = EXCLUDED.column_uid, 
	    schema_name = EXCLUDED.schema_name, table_name = EXCLUDED.table_name, column_name = EXCLUDED.column_name, clause = EXCLUDED.clause;`
}

func (q *Queries) insValuesColumnsInQueries() []string {
	var rows []string

	for uid, query := range q.ColumnsInQueries {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')",
				uid, query.QueryUID, query.TableUID, query.ColumnUID, query.Schema, query.Table, query.Name, query.Clause))
	}
	return rows
}
