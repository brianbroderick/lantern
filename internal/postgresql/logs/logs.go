package logs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/brianbroderick/lantern/internal/postgresql/ast"
	"github.com/brianbroderick/lantern/internal/postgresql/lexer"
	"github.com/brianbroderick/lantern/internal/postgresql/parser"
	"github.com/brianbroderick/lantern/pkg/repo"
	"github.com/brianbroderick/lantern/pkg/sql/projectpath"
)

func Logs() {
	databases := repo.NewDatabases()
	statements := repo.NewQueries()

	f := "postgresql.log.2024-07-10-1748.cp"
	// f := "postgresql-2024-07-09_000000.log"
	// f := "postgresql-simple.log"
	path := filepath.Join(projectpath.Root, "logs", f)

	file, err := readPayload(path)
	if HasErr("processFile", err) {
		return
	}
	l := lexer.New(string(file))
	p := parser.New(l)
	program := p.ParseProgram()

	fmt.Println("Number of statements from file", len(program.Statements))

	for i, stmt := range program.Statements {
		query := stmt.(*ast.LogStatement)

		if query.Query == "" || query.DurationLit == "" {
			continue
		}

		w := repo.QueryWorker{
			Databases:   databases,
			Database:    query.Database,
			Input:       query.Query,
			Duration:    convertTime(query.DurationLit, query.DurationMeasure),
			MustExtract: false, // We're passing in false into mustExtract because that'll happen at a later step
		}

		fmt.Println("Query:", i, query.Query)
		fmt.Println("")

		statements.Analyze(w)

		// if i%10000 == 0 {
		fmt.Printf("Parsed %d statements\n\n", i)
		// }
	}

}

func HasErr(msg string, err error) bool {
	if err != nil {
		fmt.Printf("Message: %s\nHasErr: %s\n\n", msg, err.Error())
		return true
	}
	return false
}

func readPayload(file string) ([]byte, error) {
	if len(file) == 0 {
		return []byte{}, errors.New("file is empty")
	}

	data, err := os.ReadFile(string(file))
	if HasErr("readPayload", err) {
		return []byte{}, err
	}
	return data, nil
}

// convertTime converts a string like "0.0001 sec" to an int64 of microseconds
func convertTime(time, measure string) int64 {
	if time == "" {
		return -1
	}

	ms, err := strconv.ParseFloat(time, 64)
	if err != nil {
		HasErr("convertTime", err)
	}
	switch measure {
	case "ms":
		return int64(ms * 1000)
	case "sec":
		return int64(ms * 1000000)
	default:
		return 0
	}
}
