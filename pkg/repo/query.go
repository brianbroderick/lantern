package repo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/google/uuid"
)

type Query struct {
	UID           uuid.UUID               `json:"uid,omitempty"`            // unique sha of the query
	DatabaseUID   uuid.UUID               `json:"database_uid,omitempty"`   // the dataset the query belongs to
	SourceUID     uuid.UUID               `json:"source_uid,omitempty"`     // the source the query belongs to
	QueryByHours  map[string]*QueryByHour `json:"query_by_hours,omitempty"` // query stats per hour
	Command       token.TokenType         `json:"command,omitempty"`        // the type of query
	MaskedQuery   string                  `json:"masked_query,omitempty"`   // the query with parameters masked
	UnmaskedQuery string                  `json:"unmasked_query,omitempty"` // the query with parameters unmasked
	SourceQuery   string                  `json:"source,omitempty"`         // the original query from the source

	// TimestampByHour           time.Time               `json:"timestamp_by_hour,omitempty"`            // the time the query was executed, rounded to the hour
	// TotalCount                int64                   `json:"total_count,omitempty"`                  // the number of times the query was executed
	// TotalDurationUs           int64                   `json:"total_duration_us,omitempty"`            // the total duration of all executions of the query in microseconds
	// TotalQueriesInTransaction int64                   `json:"total_queries_in_transaction,omitempty"` // the sum total number of queries each time this query was executed in a transaction
	// Users                     map[string]*QueryUser   `json:"users,omitempty"`                        // the users who executed the query
}

// This is an interim struct with additional Query meta data. The struct is passed around and built up instead
// of passing around a ton of individual variables.
// This is used both initially when compiling a list of queries and then individually when processing each query
type QueryWorker struct {
	TimestampByHour       time.Time
	Databases             *Databases
	Source                *Source
	SourceUID             uuid.UUID
	Database              string
	DatabaseUID           uuid.UUID
	UserName              string
	Input                 string // Original query. This may contain many queries
	TransactionQueryCount int64  // Number of queries in a transaction
	DurationUs            int64  // Duration of the query in microseconds
	MustExtract           bool
	Command               token.TokenType
	Masked                string // Masked query. This is the query with all values replaced with ?
	Unmasked              string // Unmasked query. This is the query with all values left alone
}

// Process processes a query and returns a bool whether or not the query was parsed successfully
func (q *Query) Process(w QueryWorker, qs *Queries) bool {
	l := lexer.New(q.UnmaskedQuery)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, msg := range p.Errors() {
			logit.Append("query-process-error", fmt.Sprintf("%s | Input: %s", msg, p.Input()))
		}
		return false
	}

	for _, stmt := range program.Statements {
		env := object.NewEnvironment()
		r := extractor.NewExtractor(&stmt, w.MustExtract)
		r.Extract(*r.Ast, env)

		qs.addTablesInQueries(q, r)
		qs.addColumnsInQueries(q, r)
		qs.addTableJoinsInQueries(q, r)
		qs.addCreateStatements(q, r)

	}

	return true
}

func (q *Query) MarshalJSON() ([]byte, error) {
	type Alias Query
	return json.Marshal(&struct {
		Command string `json:"command,omitempty"`
		// TimestampByHour string `json:"timestamp_by_hour,omitempty"`

		*Alias
	}{
		Command: q.Command.String(),
		// TimestampByHour: q.TimestampByHour.Format("2006-01-02 15:04:05"),
		Alias: (*Alias)(q),
	})
}

func (q *Query) UnmarshalJSON(data []byte) error {
	type Alias Query
	aux := &struct {
		Command string `json:"command,omitempty"`
		// TimestampByHour string `json:"timestamp_by_hour,omitempty"`

		*Alias
	}{
		Alias: (*Alias)(q),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	q.Command = token.Lookup(aux.Command)
	// t, err := time.Parse("2006-01-02 15:04:05", aux.TimestampByHour)
	// if err != nil {
	// 	return err
	// }
	// q.TimestampByHour = t
	return nil
}
