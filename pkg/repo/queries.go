package repo

import (
	"fmt"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/google/uuid"
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

// Process processes a query and returns a bool whether or not the query was parsed successfully
func (q *Queries) Process(w QueryWorker) bool {
	l := lexer.New(w.Input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, msg := range p.Errors() {
			logit.Append("queries-process-error", fmt.Sprintf("%s | Input: %s", msg, p.Input()))
		}
		return false
	}

	for _, stmt := range program.Statements {
		env := object.NewEnvironment()
		r := extractor.NewExtractor(&stmt, w.MustExtract)
		r.Extract(*r.Ast, env)

		w.Masked = stmt.String(true)    // maskParams = true, i.e. replace all values with ?
		w.Unmasked = stmt.String(false) // maskParams = false, i.e. leave params alone
		w.Command = stmt.Command()

		q.addQuery(w)
	}

	return true
}

// addQuery adds a query to the Queries struct
func (q *Queries) addQuery(w QueryWorker) {
	uid := UuidV5(w.Masked)
	uidStr := uid.String()

	var sourceUID uuid.UUID

	if w.Source != nil {
		sourceUID = w.Source.UID
	} else if w.SourceUID != uuid.Nil {
		sourceUID = w.SourceUID
	}

	if _, ok := q.Queries[uidStr]; !ok {
		q.Queries[uidStr] = &Query{
			UID:           uid,
			DatabaseUID:   w.Databases.addDatabase(w.Database),
			SourceUID:     sourceUID,
			SourceQuery:   w.Input,
			MaskedQuery:   w.Masked,
			UnmaskedQuery: w.Unmasked,
			TotalCount:    1,
			TotalDuration: w.Duration,
			Command:       w.Command,
		}
	} else {
		q.Queries[uidStr].TotalCount++
		q.Queries[uidStr].TotalDuration += w.Duration
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
