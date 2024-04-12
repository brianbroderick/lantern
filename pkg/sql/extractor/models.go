package extractor

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/google/uuid"
)

// Relationships between Objects and Queries
type FunctionsInQueries struct {
	UID         uuid.UUID `json:"uid"`
	FunctionUID uuid.UUID `json:"function_uid"`
	QueryUID    uuid.UUID `json:"query_uid"`
	Name        string    `json:"function_name"`
}

type ColumnsInQueries struct {
	UID       uuid.UUID       `json:"uid"`
	ColumnUID uuid.UUID       `json:"column_uid"`
	TableUID  uuid.UUID       `json:"table_uid"`
	QueryUID  uuid.UUID       `json:"query_uid"`
	Schema    string          `json:"schema_name"`
	Table     string          `json:"table_name"`
	Name      string          `json:"column_name"`
	Clause    token.TokenType `json:"clause"`
}

type TablesInQueries struct {
	UID      uuid.UUID `json:"uid"`
	TableUID uuid.UUID `json:"table_uid"`
	QueryUID uuid.UUID `json:"query_uid"`
	Schema   string    `json:"schema_name"`
	Name     string    `json:"table_name"`
}

// May have to store the reverse join as well
type TableJoinsInQueries struct {
	UID           uuid.UUID `json:"uid"`
	QueryUID      uuid.UUID `json:"query_uid"`
	TableUIDa     uuid.UUID `json:"table_uid_a"`
	TableUIDb     uuid.UUID `json:"table_uid_b"`
	JoinCondition string    `json:"join_condition"` // LEFT, RIGHT, INNER, OUTER, etc
	OnCondition   string    `json:"on_condition"`   // Right now, this is the String() of the expression
	TableA        string    `json:"table_a"`
	TableB        string    `json:"table_b"`
}

// TODO: check backwards in case someone joins the opposite way
// Maybe alphabetical order of the table names?
func (d *Extractor) AddJoinInQuery(columnA, columnB *ast.Identifier, on_condition string) *TableJoinsInQueries {
	alphabetical := func(a, b string) (string, string) {
		if a < b {
			return a, b
		}
		return b, a
	}

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

	a, b := alphabetical(tableA, tableB)

	uniq := UuidV5(fmt.Sprintf("%s|%s", a, b))
	uniqStr := uniq.String()

	if _, ok := d.TableJoinsInQueries[uniqStr]; !ok {

		d.TableJoinsInQueries[uniqStr] = &TableJoinsInQueries{
			UID:         uniq,
			TableUIDa:   UuidV5(a),
			TableUIDb:   UuidV5(b),
			OnCondition: on_condition,
			TableA:      a,
			TableB:      b,
		}
	}

	return d.TableJoinsInQueries[uniqStr]
}

// AddColumnsInQueries adds a column to the extractor. If the column already exists, it returns the existing column.
// This will potentially add calculated columns as it doesn't yet map to an existing table, just whatever is in the identifier.
// In a later step, columns will be mapped to real tables.
func (d *Extractor) AddColumnsInQueries(ident *ast.Identifier, clause token.TokenType) *ColumnsInQueries {
	var (
		schema string
		table  string
		column string
	)
	switch len(ident.Value) {
	case 1:
		switch ident.Value[0].(type) {
		case *ast.SimpleIdentifier:
			column = ident.Value[0].(*ast.SimpleIdentifier).Value
		case *ast.WildcardLiteral:
			column = "*"
		}
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

		switch ident.Value[2].(type) {
		case *ast.SimpleIdentifier:
			column = ident.Value[2].(*ast.SimpleIdentifier).Value
		case *ast.WildcardLiteral:
			column = "*"
		}
	}

	fqcn := fmt.Sprintf("%s|%s", clause.String(), ident.String(false)) // fqcn is the fully qualified column name with clause

	if _, ok := d.ColumnsInQueries[fqcn]; !ok {
		uid := UuidV5(fqcn)

		d.ColumnsInQueries[fqcn] = &ColumnsInQueries{
			UID:    uid,
			Schema: schema,
			Table:  table,
			Name:   column,
			Clause: clause,
		}
	}

	return d.ColumnsInQueries[fqcn]
}

// AddTablesInQueries adds a table to the extractor. If the table already exists, it returns the existing table.
// It doesn't know the QueryUID yet, so this is a generic table that will be mapped to a query later.
func (d *Extractor) AddTablesInQueries(ident *ast.Identifier) *TablesInQueries {
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

	if _, ok := d.TablesInQueries[fqtn]; !ok {
		tableUid := UuidV5(fqtn)

		d.TablesInQueries[fqtn] = &TablesInQueries{
			TableUID: tableUid,
			Schema:   schema,
			Name:     table,
		}
	}

	return d.TablesInQueries[fqtn]
}

func (d *Extractor) AddFunctionsInQueries(ident *ast.Identifier) *FunctionsInQueries {
	fqn := ident.String(false) // fqn is the fully qualified function name

	if _, ok := d.FunctionsInQueries[fqn]; !ok {
		uid := UuidV5(fqn)

		d.FunctionsInQueries[fqn] = &FunctionsInQueries{
			UID:  uid,
			Name: fqn,
		}
	}

	return d.FunctionsInQueries[fqn]
}
