package repo

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/extractor"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/google/uuid"
)

type Queries struct {
	Source                    string                                    `json:"source,omitempty"`
	Queries                   map[string]*Query                         `json:"queries,omitempty"`
	FunctionsInQueries        map[string]*extractor.FunctionsInQueries  `json:"functions_in_queries,omitempty"`
	ColumnsInQueries          map[string]*extractor.ColumnsInQueries    `json:"columns_in_queries,omitempty"`
	TablesInQueries           map[string]*extractor.TablesInQueries     `json:"tables_in_queries,omitempty"`
	TableJoinsInQueries       map[string]*extractor.TableJoinsInQueries `json:"table_joins_in_queries,omitempty"`
	Tables                    map[string]*extractor.Tables              `json:"tables,omitempty"`
	CreateStatementsInQueries map[string]*CreateStatementsInQueries     `json:"create_statements_in_queries,omitempty"`
	CreateStatements          map[string]*CreateStatement               `json:"create_statements,omitempty"`

	Errors map[string]int `json:"errors,omitempty"`
}

// NewQueries creates a new Queries struct
func NewQueries(source string) *Queries {
	return &Queries{
		Source:                    source,
		Queries:                   make(map[string]*Query),
		FunctionsInQueries:        make(map[string]*extractor.FunctionsInQueries),
		ColumnsInQueries:          make(map[string]*extractor.ColumnsInQueries),
		TablesInQueries:           make(map[string]*extractor.TablesInQueries),
		TableJoinsInQueries:       make(map[string]*extractor.TableJoinsInQueries),
		Tables:                    make(map[string]*extractor.Tables),
		CreateStatementsInQueries: make(map[string]*CreateStatementsInQueries),
		CreateStatements:          make(map[string]*CreateStatement),

		Errors: make(map[string]int),
	}
}

func (q *Queries) Process() bool {
	for _, query := range q.Queries {
		w := QueryWorker{
			MustExtract: true,
		}

		query.Process(w, q)
	}

	q.ExtractStats()
	q.UpsertQueryUsers()
	q.UpsertTablesInQueries()
	q.UpsertColumnsInQueries()
	q.UpsertTableJoinsInQueries()
	q.UpsertTables() // must run after UpsertTablesInQueries to populate the tables map
	q.UpsertCreateStatements()
	q.UpsertCreateStatementsInQueries()

	return true
}

// Analyze processes a query and returns a bool whether or not the query was parsed successfully
// This ends up calling addQuery which adds the query to the Queries struct
// Then the Queries struct is cached as a JSON file
func (q *Queries) Analyze(w QueryWorker) bool {
	l := lexer.New(w.Input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		if val, ok := q.Errors[p.Errors()[0]]; ok {
			q.Errors[p.Errors()[0]] = val + 1
		} else {
			q.Errors[p.Errors()[0]] = 1
		}

		sqlLen := len(w.Input)
		truncated := ""
		if sqlLen > 1048576 {
			truncated = "... [truncated]"
			sqlLen = 1048576
		}

		logit.Append("queries-process-error", fmt.Sprintf("Next Errors from: %s%s\n---\n", w.Input[0:sqlLen], truncated))

		for _, msg := range p.Errors() {
			logit.Append("queries-process-error", msg)
		}
		return false
	}

	// Determine how many queries are in the statement since we can send multiple queries in a single statement
	var queryCount int64
loop:
	for _, stmt := range program.Statements {
		cmd := stmt.Command()
		switch cmd {
		case token.SEMICOLON, token.SET, token.COMMIT, token.ROLLBACK: // add token.BEGIN when we're parsing this token
			continue loop
		}
		queryCount++
	}
	w.TransactionQueryCount = queryCount

	for _, stmt := range program.Statements {
		r := extractor.NewExtractor(&stmt, w.MustExtract)
		r.Execute(*r.Ast)

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

	var (
		durationUs            int64
		transactionQueryCount int64
	)

	switch w.Command {
	case token.SEMICOLON, token.SET, token.COMMIT, token.ROLLBACK: // add token.BEGIN when we're parsing this token
		durationUs = 0
		transactionQueryCount = 1
	default:
		durationUs = int64(math.Round(float64(w.DurationUs) / float64(w.TransactionQueryCount)))
		transactionQueryCount = w.TransactionQueryCount
	}

	ts := w.TimestampByHour.Format("2006-01-02 15:00:00")
	qbhUID := UuidV5(fmt.Sprintf("%s|%s", uidStr, ts))

	// All Queries
	if _, ok := q.Queries[uidStr]; !ok {
		database := w.Databases.AddDatabase(w.Database, "")

		users := make(map[string]*QueryUser)
		users[w.UserName] = &QueryUser{UID: UuidV5(fmt.Sprintf("%s|%s", w.UserName, uidStr)), QueriesByHourUID: qbhUID, UserName: w.UserName, TotalCount: 1, TotalDurationUs: durationUs}

		queryByHours := make(map[string]*QueryByHour)

		queryByHours[ts] = &QueryByHour{
			UID:                       qbhUID,
			QueryUID:                  uid,
			QueriedDate:               w.TimestampByHour.Format("2006-01-02"),
			QueriedHour:               w.TimestampByHour.Hour(),
			TotalCount:                1,
			TotalDurationUs:           durationUs,
			TotalQueriesInTransaction: transactionQueryCount,
			Users:                     users,
		}

		q.Queries[uidStr] = &Query{
			UID:           uid,
			DatabaseUID:   database.UID,
			SourceUID:     sourceUID,
			SourceQuery:   w.Input,
			MaskedQuery:   w.Masked,
			UnmaskedQuery: w.Unmasked,
			Command:       w.Command,
			QueryByHours:  queryByHours,
		}
	} else {
		q.Queries[uidStr].QueryByHours[ts].TotalCount++
		q.Queries[uidStr].QueryByHours[ts].TotalDurationUs += durationUs
		q.Queries[uidStr].QueryByHours[ts].TotalQueriesInTransaction += transactionQueryCount

		if _, ok := q.Queries[uidStr].QueryByHours[ts].Users[w.UserName]; !ok {
			q.Queries[uidStr].QueryByHours[ts].Users[w.UserName] = &QueryUser{UID: UuidV5(fmt.Sprintf("%s|%s", w.UserName, uidStr)), QueriesByHourUID: qbhUID, UserName: w.UserName, TotalCount: 1, TotalDurationUs: durationUs}
		} else {
			q.Queries[uidStr].QueryByHours[ts].Users[w.UserName].TotalCount++
			q.Queries[uidStr].QueryByHours[ts].Users[w.UserName].TotalDurationUs += durationUs
		}
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
	return `INSERT INTO queries (
	uid, database_uid, source_uid, command, 	
	masked_query, unmasked_query, source_query) 
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
			fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', '%s')",
				uid, query.DatabaseUID, query.SourceUID, query.Command.String(),
				// query.TotalCount, query.TotalDurationUs, query.TotalQueriesInTransaction,
				// int64(math.Round(float64(query.TotalDurationUs)/float64(query.TotalCount))), float64(query.TotalQueriesInTransaction)/float64(query.TotalCount),
				masked, unmasked, original))
	}
	return rows
}

func (q *Queries) ExtractStats() {
	fmt.Printf("Queries Len: %d\n", len(q.Queries))
	fmt.Printf("TablesInQueries Len: %d\n", len(q.TablesInQueries))
	fmt.Printf("ColumnsInQueries Len: %d\n", len(q.ColumnsInQueries))
	fmt.Printf("TableJoinsInQueries Len: %d\n", len(q.TableJoinsInQueries))
}

func (q *Queries) LogAggregateOfErrors() {
	errCnt := 0
	errs := make([]string, 0, len(q.Errors))

	for key := range q.Errors {
		errCnt += q.Errors[key]
		errs = append(errs, key)
	}

	sort.SliceStable(errs, func(i, j int) bool {
		return q.Errors[errs[i]] > q.Errors[errs[j]]
	})

	for _, key := range errs {
		msg := fmt.Sprintf("  %s: %d", key, q.Errors[key])
		logit.Append("queries-process-error-aggregate", msg)
	}
}
