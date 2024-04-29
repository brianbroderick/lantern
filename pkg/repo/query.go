package repo

import (
	"encoding/json"
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/google/uuid"
)

type Query struct {
	UID           uuid.UUID       `json:"uid,omitempty"`            // unique sha of the query
	DatabaseUID   uuid.UUID       `json:"database_uid,omitempty"`   // the dataset the query belongs to
	SourceUID     uuid.UUID       `json:"source_uid,omitempty"`     // the source the query belongs to
	Command       token.TokenType `json:"command,omitempty"`        // the type of query
	TotalCount    int64           `json:"total_count,omitempty"`    // the number of times the query was executed
	TotalDuration int64           `json:"total_duration,omitempty"` // the total duration of all executions of the query in microseconds
	MaskedQuery   string          `json:"masked_query,omitempty"`   // the query with parameters masked
	UnmaskedQuery string          `json:"unmasked_query,omitempty"` // the query with parameters unmasked
	SourceQuery   string          `json:"source,omitempty"`         // the original query from the source
}

// This is an interim struct with additional Query meta data. The struct is passed around and built up instead
// of passing around a ton of individual variables.
// This is used both initially when compiling a list of queries and then individually when processing each query
type QueryWorker struct {
	Databases   *Databases
	Source      *Source
	SourceUID   uuid.UUID
	Database    string
	DatabaseUID uuid.UUID
	Input       string // Original query. This may contain many queries
	Duration    int64
	MustExtract bool
	Command     token.TokenType
	Masked      string // Masked query. This is the query with all values replaced with ?
	Unmasked    string // Unmasked query. This is the query with all values left alone
}

// Process processes a query and returns a bool whether or not the query was parsed successfully
func (q *Query) Process(w QueryWorker, qs *Queries) bool {
	l := lexer.New(w.Unmasked)
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

		// w.Masked = stmt.String(true)    // maskParams = true, i.e. replace all values with ?
		// w.Unmasked = stmt.String(false) // maskParams = false, i.e. leave params alone
	}

	return true
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
