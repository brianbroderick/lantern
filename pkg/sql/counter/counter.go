package counter

// import (
// 	"github.com/brianbroderick/lantern/pkg/sql/lexer"
// 	"github.com/brianbroderick/lantern/pkg/sql/logit"
// 	"github.com/brianbroderick/lantern/pkg/sql/parser"
// 	"github.com/brianbroderick/lantern/pkg/sql/token"
// )

// // The counter package parses SQL queries and returns a count of the number of times each query is executed.

// // NewQueries creates a new Queries struct
// func NewQueries() *Queries {
// 	return &Queries{
// 		Queries: make(map[string]*Query),
// 	}
// }

// // ProcessQuery processes a query and returns a bool whether or not the query was parsed successfully
// func (q *Queries) ProcessQuery(input string, duration int64) bool {
// 	l := lexer.New(input)
// 	p := parser.New(l)
// 	program := p.ParseProgram()

// 	if len(p.Errors()) > 0 {
// 		for _, msg := range p.Errors() {
// 			logit.Append("counter_error", msg)
// 		}
// 		return false
// 	}

// 	for _, stmt := range program.Statements {
// 		output := stmt.String(true) // maskParams = true, i.e. replace all values with $1, $2, etc.

// 		q.addQuery(input, output, duration, stmt.Command())
// 	}

// 	return true
// }

// // addQuery adds a query to the Queries struct
// func (q *Queries) addQuery(input, output string, duration int64, command token.TokenType) {
// 	sha := ShaQuery(output)

// 	if _, ok := q.Queries[sha]; !ok {
// 		q.Queries[sha] = &Query{
// 			Sha:           sha,
// 			OriginalQuery: input,
// 			MaskedQuery:   output,
// 			TotalCount:    1,
// 			TotalDuration: duration,
// 			Command:       command,
// 		}
// 	} else {
// 		q.Queries[sha].TotalCount++
// 		q.Queries[sha].TotalDuration += duration
// 	}
// }
