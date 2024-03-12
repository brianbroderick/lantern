package repo

import (
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/google/uuid"
)

type Databases struct {
	Databases map[string]uuid.UUID `json:"databases,omitempty"` // the key is the sha of the database
}

func (d *Databases) addDatabase(database string) {
	if _, ok := d.Databases[database]; !ok {
		d.Databases[database] = UuidV5(database)
	}
}

func NewDatabases() *Databases {
	return &Databases{
		Databases: make(map[string]uuid.UUID),
	}
}

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
func (q *Queries) ProcessQuery(database, input string, duration int64) bool {
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

		q.addQuery(database, input, output, duration, stmt.Command())
	}

	return true
}

// addQuery adds a query to the Queries struct
func (q *Queries) addQuery(database, input, output string, duration int64, command token.TokenType) {
	uid := UuidV5(output)
	uidStr := uid.String()

	if _, ok := q.Queries[uidStr]; !ok {
		q.Queries[uidStr] = &Query{
			UID:           uid,
			DatabaseUID:   UuidV5(database),
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
	return `INSERT INTO queries (uid, command, total_count, total_duration, masked_query, original_query) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (q *Queries) insValues() []string {
	var rows []string

	for uid, query := range q.Queries {
		masked := strings.ReplaceAll(query.MaskedQuery, "'", "''")
		original := strings.ReplaceAll(query.OriginalQuery, "'", "''")

		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', '%d', '%d', '%s', '%s')",
				uid, query.DatabaseUID, query.Command.String(), query.TotalCount, query.TotalDuration, masked, original))

	}
	return rows
}
