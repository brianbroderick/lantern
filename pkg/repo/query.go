package repo

import (
	"encoding/json"

	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/google/uuid"
)

type Query struct {
	UID           uuid.UUID       `json:"uid,omitempty"`            // unique sha of the query
	DatabaseUID   uuid.UUID       `json:"database_uid,omitempty"`   // the dataset the query belongs to
	Command       token.TokenType `json:"command,omitempty"`        // the type of query
	TotalCount    int64           `json:"total_count,omitempty"`    // the number of times the query was executed
	TotalDuration int64           `json:"total_duration,omitempty"` // the total duration of all executions of the query in microseconds
	MaskedQuery   string          `json:"masked_query,omitempty"`   // the query with parameters masked
	OriginalQuery string          `json:"original_query,omitempty"` // the original query
}

func (q *Query) MarshalJSON() ([]byte, error) {
	type Alias Query
	return json.Marshal(&struct {
		Command string `json:"command,omitempty"`
		*Alias
	}{
		Command: q.Command.String(),
		Alias:   (*Alias)(q),
	})
}

func (q *Query) UnmarshalJSON(data []byte) error {
	type Alias Query
	aux := &struct {
		Command string `json:"command,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(q),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	q.Command = token.Lookup(aux.Command)
	return nil
}
