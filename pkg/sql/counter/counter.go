package counter

import (
	"fmt"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
)

// The counter package parses SQL queries and returns a count of the number of times each query is executed.

type Queries struct {
	Queries map[string]Query // the key is the sha of the query
}

type Query struct {
	Sha           string // unique sha of the query
	Query         string // the original query
	TotalCount    int64  // the number of times the query was executed
	TotalDuration int64  // the total duration of all executions of the query in milliseconds
}

// NewQueries creates a new Queries struct
func NewQueries() *Queries {
	return &Queries{
		Queries: make(map[string]Query),
	}
}

// AddQuery adds a query to the Queries struct
func (q *Queries) AddQuery(query string, duration int64) {
	sha := ShaQuery(query)

	if _, ok := q.Queries[sha]; !ok {
		q.Queries[sha] = Query{
			Sha:           sha,
			Query:         query,
			TotalCount:    1,
			TotalDuration: duration,
		}
	} else {
		q.Queries[sha] = Query{
			Sha:           sha,
			Query:         query,
			TotalCount:    q.Queries[sha].TotalCount + 1,
			TotalDuration: q.Queries[sha].TotalDuration + duration,
		}
	}
}

// ProcessQuery processes a query and returns a bool whether or not the query was parsed successfully
func (q *Queries) ProcessQuery(query string, duration int64) bool {
	l := lexer.New(query)
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

		q.AddQuery(output, duration)
	}

	return true
}

func (q *Queries) Stats() []string {
	stats := make([]string, 0)

	stats = append(stats, fmt.Sprintf("Number of unique queries: %d", len(q.Queries)))

	return stats
}
