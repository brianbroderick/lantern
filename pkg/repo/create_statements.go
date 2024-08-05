package repo

import (
	"database/sql"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
)

func NewCreateStatements() *CreateStatements {
	return &CreateStatements{
		CreateStatements: make(map[string]*CreateStatement),
	}
}

type CreateStatements struct {
	CreateStatements map[string]*CreateStatement `json:"create_statements,omitempty"`
}

func (q *Queries) addCreateStatements(qu *Query, ext *extractor.Extractor) {
	for _, cs := range ext.CreateStatements {
		if _, ok := q.CreateStatements[cs.UID.String()]; !ok {
			q.CreateStatements[cs.UID.String()] = &CreateStatement{
				UID:              cs.UID,
				Scope:            cs.Scope,
				IsUnique:         cs.IsUnique,
				UsedConcurrently: cs.UsedConcurrently,
				IsTemp:           cs.IsTemp,
				IsUnlogged:       cs.IsUnlogged,
				ObjectType:       cs.ObjectType,
				IfNotExists:      cs.IfNotExists,
				Name:             cs.Name,
				OnCommit:         cs.OnCommit,
				Operator:         cs.Operator,
				Expression:       cs.Expression,
				WhereClause:      cs.WhereClause,
			}
		}

		uid := UuidV5(fmt.Sprintf("%s|%s", qu.UID, cs.UID))
		uidStr := uid.String()
		if _, ok := q.CreateStatementsInQueries[uidStr]; !ok {
			q.CreateStatementsInQueries[uidStr] = &CreateStatementsInQueries{
				UID:                uid,
				CreateStatementUID: cs.UID,
				QueryUID:           qu.UID,
			}
		}
	}
}

func (c *CreateStatements) SetAll(db *sql.DB) {
	sql := `SELECT uid, scope, is_unique, used_concurrently, is_temp, is_unlogged, object_type, if_not_exists, name, on_commit, operator, expression, where_clause FROM create_statements`
	rows, err := db.Query(sql)
	if HasErr("SetAll Query", err) {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cs CreateStatement
		err := rows.Scan(&cs.UID, &cs.Scope, &cs.IsUnique, &cs.UsedConcurrently, &cs.IsTemp, &cs.IsUnlogged, &cs.ObjectType, &cs.IfNotExists, &cs.Name, &cs.OnCommit, &cs.Operator, &cs.Expression, &cs.WhereClause)
		if HasErr("SetAll Scan", err) {
			return
		}

		c.Add(&cs)
	}
}

func (c *CreateStatements) Add(cs *CreateStatement) *CreateStatement {
	if _, ok := c.CreateStatements[cs.UID.String()]; !ok {
		c.CreateStatements[cs.UID.String()] = cs
	}

	return c.CreateStatements[cs.UID.String()]
}

func (c *CreateStatements) CountInDB(db *sql.DB) int64 {
	var count int64
	row := db.QueryRow("SELECT COUNT(1) FROM create_statements")
	row.Scan(&count)

	return count
}
