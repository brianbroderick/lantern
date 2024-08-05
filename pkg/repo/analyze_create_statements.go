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
	is_unlogged, object_type, if_not_exists, name, on_commit, operator)
	VALUES %s
	ON CONFLICT (uid) DO UPDATE
	SET scope = excluded.scope, 
	    is_unique = excluded.is_unique, 
	    used_concurrently = excluded.used_concurrently, 
	    is_temp = excluded.is_temp, 
	    is_unlogged = excluded.is_unlogged, 
	    object_type = excluded.object_type, 
	    if_not_exists = excluded.if_not_exists, 
	    name = excluded.name, 
	    on_commit = excluded.on_commit, 
	    operator = excluded.operator;`
}

func (c *Queries) insValuesCreateStatements() []string {
	var rows []string

	for _, cs := range c.CreateStatements {
		// rows = append(rows, fmt.Sprintf("('%s', '%s', '%t', '%t', '%t', '%t', '%s', '%t', '%s', '%s', '%s', '%s', '%s')",
		//	cs.UID.String(), cs.Scope, cs.IsUnique, cs.UsedConcurrently, cs.IsTemp, cs.IsUnlogged, cs.ObjectType, cs.IfNotExists, cs.Name, cs.OnCommit, cs.Operator, cs.Expression, cs.WhereClause))
		rows = append(rows, fmt.Sprintf("('%s', '%s', '%t', '%t', '%t', '%t', '%s', '%t', '%s', '%s', '%s')",
			cs.UID.String(), cs.Scope, cs.IsUnique, cs.UsedConcurrently, cs.IsTemp, cs.IsUnlogged, cs.ObjectType, cs.IfNotExists, cs.Name, cs.OnCommit, cs.Operator))
	}

	return rows
}
