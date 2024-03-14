package extractor

import (
	"fmt"

	"github.com/google/uuid"
)

type Table struct {
	UID         uuid.UUID `json:"uid"`
	DatabaseUID uuid.UUID `json:"database_uid"`
	Schema      string    `json:"schema_name"`
	Name        string    `json:"table_name"`
}

type TableInQuery struct {
	UID        uuid.UUID `json:"uid"`
	TablesUID  uuid.UUID `json:"tables_uid"`
	QueriesUID uuid.UUID `json:"queries_uid"`
}

type TableJoin struct {
	UID           uuid.UUID `json:"uid"`
	QueriesUID    uuid.UUID `json:"queries_uid"`
	TablesUIDa    uuid.UUID `json:"tables_uid_a"`
	TablesUIDb    uuid.UUID `json:"tables_uid_b"`
	JoinCondition string    `json:"join_condition"`
	OnCondition   string    `json:"on_condition"`
}

func (d *Extractor) AddTable(database_uid uuid.UUID, schema, table string) *Table {
	fqtn := fmt.Sprintf("%s.%s", schema, table)

	if _, ok := d.Tables[fqtn]; !ok {
		uid := UuidV5(fqtn)

		d.Tables[fqtn] = &Table{
			UID:         uid,
			DatabaseUID: database_uid,
			Schema:      schema,
			Name:        table,
		}

	}

	return d.Tables[table]
}

func (d *Extractor) AddTableInQuery(table_uid, query_uid uuid.UUID) *TableInQuery {
	uniq := UuidV5(fmt.Sprintf("%s.%s", table_uid, query_uid))
	uniqStr := uniq.String()

	if _, ok := d.TablesinQueries[uniqStr]; !ok {

		d.TablesinQueries[uniqStr] = &TableInQuery{
			UID:        uniq,
			TablesUID:  table_uid,
			QueriesUID: query_uid,
		}
	}

	return d.TablesinQueries[table_uid.String()]
}

func (d *Extractor) AddTableJoin(table_uid_a, table_uid_b uuid.UUID, join_condition, on_condition string) *TableJoin {
	uniq := UuidV5(fmt.Sprintf("%s.%s.%s.%s", table_uid_a, table_uid_b, join_condition, on_condition))
	uniqStr := uniq.String()

	if _, ok := d.TableJoins[uniqStr]; !ok {

		d.TableJoins[uniqStr] = &TableJoin{
			UID:           uniq,
			TablesUIDa:    table_uid_a,
			TablesUIDb:    table_uid_b,
			JoinCondition: join_condition,
			OnCondition:   on_condition,
		}
	}

	return d.TableJoins[uniqStr]
}
