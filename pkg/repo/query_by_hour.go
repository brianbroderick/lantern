package repo

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type QueryByHour struct {
	UID                       uuid.UUID             `json:"uid,omitempty"`                          // unique sha of the query plus the time
	QueryUID                  uuid.UUID             `json:"query_uid,omitempty"`                    // unique sha of the query
	QueriedDate               string                `json:"queried_date,omitempty"`                 // the date the query was executed
	QueriedHour               int                   `json:"queried_hour,omitempty"`                 // the hour the query was executed
	TotalCount                int64                 `json:"total_count,omitempty"`                  // the number of times the query was executed
	TotalDurationUs           int64                 `json:"total_duration_us,omitempty"`            // the total duration of all executions of the query in microseconds
	TotalQueriesInTransaction int64                 `json:"total_queries_in_transaction,omitempty"` // the sum total number of queries each time this query was executed in a transaction
	Users                     map[string]*QueryUser `json:"users,omitempty"`                        // the users who executed the query
}

func (q *Queries) UpsertQueryByHours() {
	if len(q.Queries) == 0 {
		return
	}

	rows := q.insValuesQueryByHours()
	query := fmt.Sprintf(q.insQueryByHours(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (q *Queries) insQueryByHours() string {
	return `INSERT INTO queries_by_hours (uid, query_uid, queried_date, queried_hour, total_count, total_duration_us, total_queries_in_transaction) 
	VALUES %s
	ON CONFLICT (uid) DO UPDATE 
	SET query_uid = EXCLUDED.query_uid, 
		queried_date = EXCLUDED.queried_date, 
		queried_hour = EXCLUDED.queried_hour, 
		total_count = EXCLUDED.total_count, 
		total_duration_us = EXCLUDED.total_duration_us, 
		total_queries_in_transaction = EXCLUDED.total_queries_in_transaction;`
}

func (q *Queries) insValuesQueryByHours() []string {
	var rows []string

	for _, query := range q.Queries {
		for _, queryByHour := range query.QueryByHours {
			rows = append(rows,
				fmt.Sprintf("('%s', '%s', '%s', %d, %d, %d, %d)",
					queryByHour.UID, queryByHour.QueryUID, queryByHour.QueriedDate, queryByHour.QueriedHour, queryByHour.TotalCount, queryByHour.TotalDurationUs, queryByHour.TotalQueriesInTransaction))
		}
	}

	return rows
}
