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

type ProcessQuery struct {
	Databases   *Databases
	Source      *Source
	Database    string
	Input       string
	Duration    int64
	MustExtract bool
	Masked      string
	Unmasked    string
	Command     token.TokenType
}

// NewQueries creates a new Queries struct
func NewQueries() *Queries {
	return &Queries{
		Queries: make(map[string]*Query),
	}
}

// ProcessQuery processes a query and returns a bool whether or not the query was parsed successfully
func (q *Queries) ProcessQuery(pq ProcessQuery) bool {
	l := lexer.New(pq.Input)
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
		r := extractor.NewExtractor(&stmt, pq.MustExtract)
		r.Extract(*r.Ast, env)

		pq.Masked = stmt.String(true)    // maskParams = true, i.e. replace all values with ?
		pq.Unmasked = stmt.String(false) // maskParams = false, i.e. leave params alone
		pq.Command = stmt.Command()

		q.addQuery(pq)
	}

	return true
}

// addQuery adds a query to the Queries struct
func (q *Queries) addQuery(pq ProcessQuery) {
	uid := UuidV5(pq.Masked)
	uidStr := uid.String()

	if _, ok := q.Queries[uidStr]; !ok {
		q.Queries[uidStr] = &Query{
			UID:           uid,
			DatabaseUID:   pq.Databases.addDatabase(pq.Database),
			SourceUID:     pq.Source.UID,
			SourceQuery:   pq.Input,
			MaskedQuery:   pq.Masked,
			UnmaskedQuery: pq.Unmasked,
			TotalCount:    1,
			TotalDuration: pq.Duration,
			Command:       pq.Command,
		}
	} else {
		q.Queries[uidStr].TotalCount++
		q.Queries[uidStr].TotalDuration += pq.Duration
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
	return `INSERT INTO queries (uid, database_uid, source_uid, command, total_count, total_duration, masked_query, unmasked_query, source_query) 
	VALUES %s 
	ON CONFLICT (uid) DO NOTHING;`
}

func (q *Queries) insValues() []string {
	var rows []string

	for uid, query := range q.Queries {
		masked := strings.ReplaceAll(query.MaskedQuery, "'", "''")
		unmasked := strings.ReplaceAll(query.UnmaskedQuery, "'", "''")
		original := strings.ReplaceAll(query.SourceQuery, "'", "''")

		rows = append(rows,
			fmt.Sprintf("('%s', '%s', '%s', '%s', '%d', '%d', '%s', '%s', '%s')",
				uid, query.DatabaseUID, query.SourceUID, query.Command.String(), query.TotalCount, query.TotalDuration, masked, unmasked, original))

	}
	return rows
}
