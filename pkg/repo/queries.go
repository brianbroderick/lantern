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
	Queries             map[string]*Query                         `json:"queries,omitempty"`
	FunctionsInQueries  map[string]*extractor.FunctionsInQueries  `json:"functions_in_queries,omitempty"`
	ColumnsInQueries    map[string]*extractor.ColumnsInQueries    `json:"columns_in_queries,omitempty"`
	TablesInQueries     map[string]*extractor.TablesInQueries     `json:"tables_in_queries,omitempty"`
	TableJoinsInQueries map[string]*extractor.TableJoinsInQueries `json:"table_joins_in_queries,omitempty"`
	Tables              map[string]*extractor.Tables              `json:"tables,omitempty"`
}

// NewQueries creates a new Queries struct
func NewQueries() *Queries {
	return &Queries{
		Queries:             make(map[string]*Query),
		FunctionsInQueries:  make(map[string]*extractor.FunctionsInQueries),
		ColumnsInQueries:    make(map[string]*extractor.ColumnsInQueries),
		TablesInQueries:     make(map[string]*extractor.TablesInQueries),
		TableJoinsInQueries: make(map[string]*extractor.TableJoinsInQueries),
		Tables:              make(map[string]*extractor.Tables),
	}
}

func (q *Queries) Process() bool {
	for _, query := range q.Queries {
		w := QueryWorker{
			Masked:      query.MaskedQuery,
			Unmasked:    query.UnmaskedQuery,
			MustExtract: true,
		}

		query.Process(w, q)
	}

	q.ExtractStats()
	q.UpsertTablesInQueries()
	q.UpsertColumnsInQueries()
	q.UpsertTableJoinsInQueries()
	q.UpsertTables() // must run after UpsertTablesInQueries to populate the tables map

	return true
}

// Analyze processes a query and returns a bool whether or not the query was parsed successfully
func (q *Queries) Analyze(w QueryWorker) bool {
	l := lexer.New(w.Input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		sqlLen := len(w.Input)
		truncated := ""
		if sqlLen > 1048576 {
			truncated = "... [truncated]"
			sqlLen = 1048576
		}

		logit.Append("queries-process-error", fmt.Sprintf("Next Errors from: %s%s", w.Input[0:sqlLen], truncated))

		for _, msg := range p.Errors() {
			logit.Append("queries-process-error", msg)
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
		database := w.Databases.AddDatabase(w.Database, "")
		q.Queries[uidStr] = &Query{
			UID:           uid,
			DatabaseUID:   database.UID,
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

func (q *Queries) CountInDB() int {
	db := Conn()
	defer db.Close()

	var count int
	row := db.QueryRow("SELECT COUNT(1) FROM queries")
	row.Scan(&count)

	return count
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

func (q *Queries) ExtractStats() {
	fmt.Printf("Queries Len: %d\n", len(q.Queries))
	fmt.Printf("TablesInQueries Len: %d\n", len(q.TablesInQueries))
	fmt.Printf("ColumnsInQueries Len: %d\n", len(q.ColumnsInQueries))
	fmt.Printf("TableJoinsInQueries Len: %d\n", len(q.TableJoinsInQueries))
}
