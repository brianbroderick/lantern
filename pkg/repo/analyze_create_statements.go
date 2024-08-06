package repo

import (
	"fmt"
	"strings"
)

func (q *Queries) UpsertCreateStatements() {
	if len(q.CreateStatements) == 0 {
		return
	}

	rows := q.insValuesCreateStatements()
	query := fmt.Sprintf(q.insCreateStatements(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (c *Queries) insCreateStatements() string {
	return `INSERT INTO create_statements (
	uid, scope, is_unique, used_concurrently, is_temp, 
	is_unlogged, object_type, if_not_exists, name, on_commit, operator, where_clause)
	VALUES %s
	ON CONFLICT (uid) DO UPDATE
	SET scope = EXCLUDED.scope, 
	    is_unique = EXCLUDED.is_unique, 
	    used_concurrently = EXCLUDED.used_concurrently, 
	    is_temp = EXCLUDED.is_temp, 
	    is_unlogged = EXCLUDED.is_unlogged, 
	    object_type = EXCLUDED.object_type, 
	    if_not_exists = EXCLUDED.if_not_exists, 
	    name = EXCLUDED.name, 
	    on_commit = EXCLUDED.on_commit, 
	    operator = EXCLUDED.operator, 
			where_clause = EXCLUDED.where_clause;`
}

func (c *Queries) insValuesCreateStatements() []string {
	var rows []string

	for _, cs := range c.CreateStatements {
		whereClause := strings.ReplaceAll(cs.WhereClause, "'", "''")

		// rows = append(rows, fmt.Sprintf("('%s', '%s', '%t', '%t', '%t', '%t', '%s', '%t', '%s', '%s', '%s', '%s', '%s')",
		//	cs.UID.String(), cs.Scope, cs.IsUnique, cs.UsedConcurrently, cs.IsTemp, cs.IsUnlogged, cs.ObjectType, cs.IfNotExists, cs.Name, cs.OnCommit, cs.Operator, cs.Expression, cs.WhereClause))
		rows = append(rows, fmt.Sprintf("('%s', '%s', '%t', '%t', '%t', '%t', '%s', '%t', '%s', '%s', '%s', '%s')",
			cs.UID.String(), cs.Scope, cs.IsUnique, cs.UsedConcurrently, cs.IsTemp, cs.IsUnlogged, cs.ObjectType, cs.IfNotExists, cs.Name, cs.OnCommit, cs.Operator, whereClause))
	}

	return rows
}
