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
	"github.com/brianbroderick/lantern/pkg/sql/logit"
	"github.com/brianbroderick/lantern/pkg/sql/projectpath"
)

func Logs() {
	logit.Clear("queries-process-error")

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

	success := 0
	analyzed := 0

loop:
	for _, stmt := range program.Statements {
		query := stmt.(*ast.LogStatement)

		switch query.PreparedStep {
		case "statement", "execute":
		default:
			continue loop
		}

		// if query.Query == "" || query.DurationLit == "" {
		// 	continue
		// }

		w := repo.QueryWorker{
			Databases:   databases,
			Database:    query.Database,
			Input:       query.Query,
			Duration:    convertTime(query.DurationLit, query.DurationMeasure),
			MustExtract: false, // We're passing in false into mustExtract because that'll happen at a later step
		}

		analyzed++
		if statements.Analyze(w) {
			success++
		}

		if analyzed%10000 == 0 {
			fmt.Printf("Parsed %d sucessfully of %d statements\n", success, analyzed)
		}
	}

	fmt.Printf("Analyzed %d of %d statements\n", analyzed, len(program.Statements))
	fmt.Printf("Parsed %d sucessfully of %d statements\n", success, analyzed)

	fmt.Printf("Number of statements: %d\n", len(statements.Queries))
	json := repo.MarshalJSON(statements)
	writeFile(filepath.Join(projectpath.Root, "processed", "queries.json"), []byte(json))

	fmt.Printf("Number of databases: %d\n", len(databases.Databases))
	dbJSON := repo.MarshalJSON(databases)
	writeFile(filepath.Join(projectpath.Root, "processed", "databases.json"), []byte(dbJSON))
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

func writeFile(file string, data []byte) error {
	if len(file) == 0 {
		return errors.New("file is empty")
	}

	err := os.WriteFile(string(file), data, 0644)
	if HasErr("writeFile", err) {
		return err
	}
	return nil
}
