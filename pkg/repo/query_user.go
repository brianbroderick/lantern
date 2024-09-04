package repo

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type QueryUser struct {
	UID              uuid.UUID `json:"uid,omitempty"`
	QueriesByHourUID uuid.UUID `json:"queries_by_hour,omitempty"`
	UserName         string    `json:"user_name,omitempty"`
	TotalCount       int64     `json:"total_count,omitempty"`
	TotalDurationUs  int64     `json:"total_duration_us,omitempty"`
}

func (q *Queries) UpsertQueryUsers() {
	if len(q.Queries) == 0 {
		return
	}

	rows := q.insValuesQueryUsers()
	query := fmt.Sprintf(q.insQueryUsers(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (q *Queries) insQueryUsers() string {
	return `INSERT INTO query_users (uid, queries_by_hour_uid, user_name, total_count, total_duration_us) 
	VALUES %s
	ON CONFLICT (uid) DO UPDATE 
	SET queries_by_hour_uid = EXCLUDED.queries_by_hour_uid, 
		user_name = EXCLUDED.user_name, 
		total_count = EXCLUDED.total_count, 
		total_duration_us = EXCLUDED.total_duration_us;`
}

func (q *Queries) insValuesQueryUsers() []string {
	var rows []string

	for _, query := range q.Queries {
		for _, queryByHour := range query.QueryByHours {
			if len(queryByHour.Users) == 0 {
				continue
			}

			for _, user := range queryByHour.Users {
				rows = append(rows,
					fmt.Sprintf("('%s', '%s', '%s', %d, %d)",
						user.UID, user.QueriesByHourUID, user.UserName, user.TotalCount, user.TotalDurationUs))
			}
		}
	}

	return rows
}
