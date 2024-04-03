package repo

import (
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

type Queries struct {
	Queries map[string]*Query `json:"queries,omitempty"` // the key is the sha of the query
}

// NewQueries creates a new Queries struct
func NewQueries() *Queries {
	return &Queries{
		Queries: make(map[string]*Query),
	}
}

// ProcessQuery processes a query and returns a bool whether or not the query was parsed successfully
func (q *Queries) ProcessQuery(databases *Databases, source *Source, database, input string, duration int64, mustExtract bool) bool {
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
		env := object.NewEnvironment()
		r := extractor.NewExtractor(&stmt, mustExtract)
		r.Extract(*r.Ast, env)

		output := stmt.String(true) // maskParams = true, i.e. replace all values with ?

		q.addQuery(databases, database, source, input, output, duration, stmt.Command())
	}

	return true
}

// addQuery adds a query to the Queries struct
func (q *Queries) addQuery(databases *Databases, database string, source *Source, input, output string, duration int64, command token.TokenType) {
	uid := UuidV5(output)
	uidStr := uid.String()

	if _, ok := q.Queries[uidStr]; !ok {
		q.Queries[uidStr] = &Query{
			UID:           uid,
			DatabaseUID:   databases.addDatabase(database),
			SourceUID:     source.UID,
			OriginalQuery: input,
			MaskedQuery:   output,
			TotalCount:    1,
			TotalDuration: duration,
			Command:       command,
		}
	} else {
		q.Queries[uidStr].TotalCount++
		q.Queries[uidStr].TotalDuration += duration
	}
}

func (q *Queries) Upsert() {
	if len(q.Queries) == 0 {
		return
	}

	rows := q.insValues()
	query := fmt.Sprintf(q.ins(), strings.Join(rows, ",\n"))

	db := Conn()
	defer db.Close()
	ExecuteQuery(db, query)
}

func (q *Queries) ins() string {
	return `INSERT INTO queries (uid, database_uid, source_uid, command, total_count, total_duration, masked_query, original_query) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (q *Queries) insValues() []string {
	var rows []string

	for uid, query := range q.Queries {
		masked := strings.ReplaceAll(query.MaskedQuery, "'", "''")
		original := strings.ReplaceAll(query.OriginalQuery, "'", "''")

		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', '%s', '%d', '%d', '%s', '%s')",
				uid, query.DatabaseUID, query.SourceUID, query.Command.String(), query.TotalCount, query.TotalDuration, masked, original))

	}
	return rows
}
