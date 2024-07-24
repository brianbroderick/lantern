package extractor

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/google/uuid"
)

// Defaults:
// Schema: public
// Table: UNKNOWN

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
	Command   token.TokenType `json:"command"`
	Clause    token.TokenType `json:"clause"`
}

type TablesInQueries struct {
	UID      uuid.UUID       `json:"uid"`
	TableUID uuid.UUID       `json:"table_uid"`
	QueryUID uuid.UUID       `json:"query_uid"`
	Command  token.TokenType `json:"command"`
	Schema   string          `json:"schema_name"`
	Name     string          `json:"table_name"`
}

type Tables struct {
	UID    uuid.UUID `json:"uid"`
	Schema string    `json:"schema_name"`
	Name   string    `json:"table_name"`
}

// May have to store the reverse join as well
type TableJoinsInQueries struct {
	UID           uuid.UUID `json:"uid"`
	QueryUID      uuid.UUID `json:"query_uid"`
	TableUIDa     uuid.UUID `json:"table_uid_a"`
	TableUIDb     uuid.UUID `json:"table_uid_b"`
	JoinCondition string    `json:"join_condition"` // LEFT, RIGHT, INNER, OUTER, etc
	OnCondition   string    `json:"on_condition"`   // Right now, this is the String() of the expression
	SchemaA       string    `json:"schema_a"`
	TableA        string    `json:"table_a"`
	SchemaB       string    `json:"schema_b"`
	TableB        string    `json:"table_b"`
}

// This passes around a 3 element slice. []string{schema, table, fully_qualified_table_name}
func (d *Extractor) AddJoinInQuery(columnA, columnB *ast.Identifier, on_condition string, env *object.Environment) *TableJoinsInQueries {
	alphabetical := func(a, b []string) ([]string, []string) {
		if a[2] < b[2] {
			return a, b
		}
		return b, a
	}

	tableA := []string{"public", "UNKNOWN", "public.UNKNOWN"}
	tableB := []string{"public", "UNKNOWN", "public.UNKNOWN"}

	switch len(columnA.Value) {
	case 1:
		// TODO: add support in resolver to add tables when not specified
		fmt.Println("AddJoin: columns do not have tables associated with them")
	case 2:
		// assume the schema is public if it's not specified. This is a safe assumption for now.
		// TODO: add support for search_path
		tableA[1] = columnA.Value[0].(*ast.SimpleIdentifier).Value
	case 3:
		tableA[0] = columnA.Value[0].(*ast.SimpleIdentifier).Value
		tableA[1] = columnA.Value[1].(*ast.SimpleIdentifier).Value
	}
	tableA[2] = fmt.Sprintf("%s.%s", tableA[0], tableA[1])

	switch len(columnB.Value) {
	case 1:
		// TODO: add support in resolver to add tables when not specified
		fmt.Println("AddJoin: columns do not have tables associated with them")
	case 2:
		// TODO: add support for search_path
		tableB[1] = columnB.Value[0].(*ast.SimpleIdentifier).Value
	case 3:
		tableB[0] = columnB.Value[0].(*ast.SimpleIdentifier).Value
		tableB[1] = columnB.Value[1].(*ast.SimpleIdentifier).Value
	}
	tableB[2] = fmt.Sprintf("%s.%s", tableB[0], tableB[1])

	a, b := alphabetical(tableA, tableB)

	uniq := UuidV5(fmt.Sprintf("%s|%s", a[2], b[2]))
	uniqStr := uniq.String()

	if _, ok := d.TableJoinsInQueries[uniqStr]; !ok {

		var joinType string
		obj, ok := env.Get("join_type")
		if !ok {
			joinType = "UNKNOWN"
		} else {
			str, ok := obj.(*object.String)
			if !ok {
				joinType = "UNKNOWN"
			} else {
				joinType = str.Value
			}
		}

		d.TableJoinsInQueries[uniqStr] = &TableJoinsInQueries{
			UID:           uniq,
			TableUIDa:     UuidV5(a[2]),
			TableUIDb:     UuidV5(b[2]),
			OnCondition:   on_condition,
			JoinCondition: joinType,
			SchemaA:       a[0],
			TableA:        a[1],
			SchemaB:       b[0],
			TableB:        b[1],
		}
	}

	return d.TableJoinsInQueries[uniqStr]
}

// AddColumnsInQueries adds a column to the extractor. If the column already exists, it returns the existing column.
// This will potentially add calculated columns as it doesn't yet map to an existing table, just whatever is in the identifier.
// In a later step, columns will be mapped to real tables.
func (d *Extractor) AddColumnsInQueries(i *ast.Identifier) *ColumnsInQueries {
	var (
		schema string
		table  string
		column string
	)
	// it's public by default though this can be overridden.
	// TODO: PG allows you to set the search path, which can change the default schema
	// though this is rare enough that we don't yet support it.
	schema = "public"
	// TODO: add support for adding tables in the resolver when not specified.
	table = "UNKNOWN"

	switch len(i.Value) {
	case 1:
		switch i.Value[0].(type) {
		case *ast.SimpleIdentifier:
			column = i.Value[0].(*ast.SimpleIdentifier).Value
		case *ast.WildcardLiteral:
			column = "*"
		}
	case 2:
		table = i.Value[0].(*ast.SimpleIdentifier).Value
		switch i.Value[1].(type) {
		case *ast.SimpleIdentifier:
			column = i.Value[1].(*ast.SimpleIdentifier).Value
		case *ast.WildcardLiteral:
			column = "*"
		}
	case 3:
		schema = i.Value[0].(*ast.SimpleIdentifier).Value
		table = i.Value[1].(*ast.SimpleIdentifier).Value

		switch i.Value[2].(type) {
		case *ast.SimpleIdentifier:
			column = i.Value[2].(*ast.SimpleIdentifier).Value
		case *ast.WildcardLiteral:
			column = "*"
		}
	}

	fqcn := fmt.Sprintf("%s|%s.%s.%s", i.Clause().String(), schema, table, column) // fqcn is the fully qualified column name with clause

	if _, ok := d.ColumnsInQueries[fqcn]; !ok {
		uid := UuidV5(fqcn)

		d.ColumnsInQueries[fqcn] = &ColumnsInQueries{
			UID:       uid,
			Command:   i.Command(),
			Schema:    schema,
			Table:     table,
			TableUID:  UuidV5(fmt.Sprintf("%s.%s", schema, table)),
			Name:      column,
			ColumnUID: UuidV5(fmt.Sprintf("%s.%s.%s", schema, table, column)), // don't include the clause in the column UID
			Clause:    i.Clause(),
		}
	}

	return d.ColumnsInQueries[fqcn]
}

// AddTablesInQueries adds a table to the extractor. If the table already exists, it returns the existing table.
// It doesn't know the QueryUID yet, so this is a generic table that will be mapped to a query later.
func (d *Extractor) AddTablesInQueries(i *ast.Identifier) *TablesInQueries {
	schema := "public"
	table := "UNKNOWN"

	switch len(i.Value) {
	case 1:
		table = i.Value[0].(*ast.SimpleIdentifier).Value
	case 2:
		schema = i.Value[0].(*ast.SimpleIdentifier).Value
		table = i.Value[1].(*ast.SimpleIdentifier).Value
	}

	fqtn := fmt.Sprintf("%s.%s", schema, table) // fqtn is the fully qualified table name

	if _, ok := d.TablesInQueries[fqtn]; !ok {
		tableUid := UuidV5(fqtn)

		d.TablesInQueries[fqtn] = &TablesInQueries{
			TableUID: tableUid,
			Command:  i.Command(),
			Schema:   schema,
			Name:     table,
		}
	}

	if _, ok := d.Tables[fqtn]; !ok {
		d.Tables[fqtn] = &Tables{
			UID:    UuidV5(fqtn),
			Schema: schema,
			Name:   table,
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
