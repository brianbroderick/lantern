package extractor

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/google/uuid"
)

type Column struct {
	UID      uuid.UUID `json:"uid"`
	TableUID uuid.UUID `json:"table_uid"`   // This will get populated, if it matches something in the Tables map
	Schema   string    `json:"schema_name"` // This won't persist to the DB. It's a placeholder before it's compared to physical tables
	Table    string    `json:"table_name"`  // This won't persist to the DB. It's a placeholder before it's compared to physical tables
	Name     string    `json:"column_name"`
}

// This represents all of the physical tables in the query.
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
	JoinCondition string    `json:"join_condition"` // LEFT, RIGHT, INNER, OUTER, etc
	OnCondition   string    `json:"on_condition"`   // Right now, this is the String() of the expression
	TableA        string    `json:"table_a"`        // This won't be in the DB, but is for debugging purposes to see the table name
	TableB        string    `json:"table_b"`        // This won't be in the DB, but is for debugging purposes to see the table name
}

// TODO: check backwards in case someone joins the opposite way
// Maybe alphabetical order of the table names?
func (d *Extractor) AddJoin(columnA, columnB *ast.Identifier, on_condition string) *TableJoin {
	uniq := UuidV5(fmt.Sprintf("%s|%s", columnA.String(false), columnB.String(false)))
	uniqStr := uniq.String()

	var (
		tableA string
		tableB string
	)
	switch len(columnA.Value) {
	case 1:
		fmt.Println("AddJoin: columns do not have tables associated with them")
	case 2:
		tableA = columnA.Value[0].(*ast.SimpleIdentifier).Value
	case 3:
		tableA = fmt.Sprintf("%s.%s", columnA.Value[0].(*ast.SimpleIdentifier).Value, columnA.Value[1].(*ast.SimpleIdentifier).Value)
	}

	switch len(columnB.Value) {
	case 1:
		fmt.Println("AddJoin: columns do not have tables associated with them")
	case 2:
		tableB = columnB.Value[0].(*ast.SimpleIdentifier).Value
	case 3:
		tableB = fmt.Sprintf("%s.%s", columnB.Value[0].(*ast.SimpleIdentifier).Value, columnB.Value[1].(*ast.SimpleIdentifier).Value)
	}

	if _, ok := d.TableJoins[uniqStr]; !ok {

		d.TableJoins[uniqStr] = &TableJoin{
			UID:         uniq,
			TablesUIDa:  UuidV5(columnA.String(false)),
			TablesUIDb:  UuidV5(columnB.String(false)),
			OnCondition: on_condition,
			TableA:      tableA,
			TableB:      tableB,
		}
	}

	return d.TableJoins[uniqStr]
}

// AddColumn adds a column to the extractor. If the column already exists, it returns the existing column.
// This will potentially add calculated columns as it doesn't yet map to an existing table, just whatever is in the identifier.
// In a later step, columns will be mapped to real tables.
func (d *Extractor) AddColumn(ident *ast.Identifier) *Column {
	var (
		schema string
		table  string
		column string
	)
	switch len(ident.Value) {
	case 1:
		column = ident.Value[0].(*ast.SimpleIdentifier).Value
	case 2:
		table = ident.Value[0].(*ast.SimpleIdentifier).Value
		switch ident.Value[1].(type) {
		case *ast.SimpleIdentifier:
			column = ident.Value[1].(*ast.SimpleIdentifier).Value
		case *ast.WildcardLiteral:
			column = "*"
		}
	case 3:
		schema = ident.Value[0].(*ast.SimpleIdentifier).Value
		table = ident.Value[1].(*ast.SimpleIdentifier).Value
		column = ident.Value[2].(*ast.SimpleIdentifier).Value
	}

	fqcn := ident.String(false) // fqcn is the fully qualified column name

	if _, ok := d.Columns[fqcn]; !ok {
		uid := UuidV5(fqcn)

		d.Columns[fqcn] = &Column{
			UID:    uid,
			Schema: schema,
			Table:  table,
			Name:   column,
		}
	}

	return d.Columns[fqcn]
}

func (d *Extractor) AddTable(ident *ast.Identifier) *Table {
	var (
		schema string
		table  string
	)

	switch len(ident.Value) {
	case 1:
		table = ident.Value[0].(*ast.SimpleIdentifier).Value
	case 2:
		schema = ident.Value[0].(*ast.SimpleIdentifier).Value
		table = ident.Value[1].(*ast.SimpleIdentifier).Value
	}

	fqtn := ident.String(false) // fqtn is the fully qualified table name

	if _, ok := d.Tables[fqtn]; !ok {
		uid := UuidV5(fqtn)

		d.Tables[fqtn] = &Table{
			UID:    uid,
			Schema: schema,
			Name:   table,
		}
	}

	return d.Tables[fqtn]
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

// func (d *Extractor) AddTableJoin(table_uid_a, table_uid_b uuid.UUID, join_condition, on_condition string) *TableJoin {
// 	uniq := UuidV5(fmt.Sprintf("%s.%s.%s.%s", table_uid_a, table_uid_b, join_condition, on_condition))
// 	uniqStr := uniq.String()

// 	if _, ok := d.TableJoins[uniqStr]; !ok {

// 		d.TableJoins[uniqStr] = &TableJoin{
// 			UID:           uniq,
// 			TablesUIDa:    table_uid_a,
// 			TablesUIDb:    table_uid_b,
// 			JoinCondition: join_condition,
// 			OnCondition:   on_condition,
// 		}
// 	}

// 	return d.TableJoins[uniqStr]
// }
