package counter

import (
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type Queries struct {
	Queries map[string]*Query `json:"queries,omitempty"` // the key is the sha of the query
}

type Query struct {
	Sha           string          `json:"sha,omitempty"`            // unique sha of the query
	Command       token.TokenType `json:"command,omitempty"`        // the type of query
	OriginalQuery string          `json:"original_query,omitempty"` // the original query
	MaskedQuery   string          `json:"masked_query,omitempty"`   // the query with parameters masked
	TotalCount    int64           `json:"total_count,omitempty"`    // the number of times the query was executed
	TotalDuration int64           `json:"total_duration,omitempty"` // the total duration of all executions of the query in microseconds
}

type QueryStats struct {
	ByCount              []*Query                  `json:"by_count,omitempty"`
	ByDuration           []*Query                  `json:"by_duration,omitempty"`
	SumCommandByCount    map[token.TokenType]int64 `json:"sum_command_by_count,omitempty"`
	SumCommandByDuration map[token.TokenType]int64 `json:"sum_command_by_duration,omitempty"`
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

func MarshalJSON(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if HasErr("marshallJSON", err) {
		return ""
	}
	return string(b)
}

func UnmarshalJSON(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	HasErr("unmarshallJSON", err)
}

func HasErr(msg string, err error) bool {
	if err != nil {
		fmt.Printf("Message: %s\nHasErr: %s\n\n", msg, err.Error())
		return true
	}
	return false
}
