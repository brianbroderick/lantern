package repo

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type CreateStatementsInQueries struct {
	UID                uuid.UUID `json:"uid,omitempty"`
	CreateStatementUID uuid.UUID `json:"create_statement_uid,omitempty"`
	QueryUID           uuid.UUID `json:"query_uid,omitempty"`
}

func (q *Queries) UpsertCreateStatementsInQueries() {
	if len(q.CreateStatementsInQueries) == 0 {
		return
	}

	rows := q.insValuesCreateStatementsInQueries()
	query := fmt.Sprintf(q.insCreateStatementsInQueries(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (q *Queries) insCreateStatementsInQueries() string {
	return `INSERT INTO create_statements_in_queries (uid, create_statement_uid, query_uid) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (q *Queries) insValuesCreateStatementsInQueries() []string {
	var rows []string

	for _, cs := range q.CreateStatementsInQueries {
		rows = append(rows, fmt.Sprintf("('%s', '%s', '%s')",
			cs.UID, cs.CreateStatementUID, cs.QueryUID))
	}

	return rows
}
