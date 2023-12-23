package counter

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
)

// The counter package parses SQL queries and returns a count of the number of times each query is executed.

type Queries struct {
	Queries map[string]*Query // the key is the sha of the query
}

type Query struct {
	Sha           string // unique sha of the query
	OriginalQuery string // the original query
	MaskedQuery   string // the query with parameters masked
	TotalCount    int64  // the number of times the query was executed
	TotalDuration int64  // the total duration of all executions of the query in microseconds
}

// NewQueries creates a new Queries struct
func NewQueries() *Queries {
	return &Queries{
		Queries: make(map[string]*Query),
	}
}

// ProcessQuery processes a query and returns a bool whether or not the query was parsed successfully
func (q *Queries) ProcessQuery(input string, duration int64) bool {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, msg := range p.Errors() {
			logit.Append("counter_error", msg)
		}
		return false
	}

	for _, stmt := range program.Statements {
		output := stmt.String(true) // maskParams = true, i.e. replace all values with $1, $2, etc.

		q.addQuery(input, output, duration)
	}

	return true
}

// addQuery adds a query to the Queries struct
func (q *Queries) addQuery(input, output string, duration int64) {
	sha := ShaQuery(output)

	if _, ok := q.Queries[sha]; !ok {
		q.Queries[sha] = &Query{
			Sha:           sha,
			OriginalQuery: input,
			MaskedQuery:   output,
			TotalCount:    1,
			TotalDuration: duration,
		}
	} else {
		q.Queries[sha].TotalCount++
		q.Queries[sha].TotalDuration += duration
	}
}

func (q *Queries) Stats() []string {
	stats := make([]string, 0)

	stats = append(stats, fmt.Sprintf("Number of unique queries: %d", len(q.Queries)))

	return stats
}
