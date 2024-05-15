package repo

import (
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
)

// addTableJoinsInQueries adds the tables in a query to the Queries struct
func (q *Queries) addTableJoinsInQueries(qu *Query, ext *extractor.Extractor) {

	for _, table := range ext.TableJoinsInQueries {
		uid := UuidV5(fmt.Sprintf("%s|%s|%s|%s", qu.UID, table.TableUIDa, table.TableUIDb, table.OnCondition))
		uidStr := uid.String()
		if _, ok := q.TableJoinsInQueries[uidStr]; !ok {
			q.TableJoinsInQueries[uidStr] = &extractor.TableJoinsInQueries{
				UID:           uid,
				QueryUID:      qu.UID,
				TableUIDa:     table.TableUIDa,
				TableUIDb:     table.TableUIDb,
				JoinCondition: table.JoinCondition,
				OnCondition:   table.OnCondition,
				SchemaA:       table.SchemaA,
				TableA:        table.TableA,
				SchemaB:       table.SchemaB,
				TableB:        table.TableB,
			}
		}
	}
}

func (q *Queries) UpsertTableJoinsInQueries() {
	if len(q.TableJoinsInQueries) == 0 {
		return
	}

	rows := q.insValuesTableJoinsInQueries()
	query := fmt.Sprintf(q.insTableJoinsInQueries(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (q *Queries) insTableJoinsInQueries() string {
	return `INSERT INTO table_joins_in_queries (uid, query_uid, table_uid_a, table_uid_b, join_condition, on_condition, schema_a, table_a, schema_b, table_b)
	VALUES %s 
	ON CONFLICT (uid) DO UPDATE 
	SET query_uid = EXCLUDED.query_uid, table_uid_a = EXCLUDED.table_uid_a, table_uid_b = EXCLUDED.table_uid_b, 
	  join_condition = EXCLUDED.join_condition, schema_a = EXCLUDED.schema_a, table_a = EXCLUDED.table_a, schema_b = EXCLUDED.schema_b, table_b = EXCLUDED.table_b;`
}

func (q *Queries) insValuesTableJoinsInQueries() []string {
	var rows []string

	for uid, query := range q.TableJoinsInQueries {
		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')",
				uid, query.QueryUID, query.TableUIDa, query.TableUIDb, query.JoinCondition, query.OnCondition, query.SchemaA, query.TableA, query.SchemaB, query.TableB))
	}
	return rows
}
