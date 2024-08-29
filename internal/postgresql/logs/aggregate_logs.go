package logs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/brianbroderick/lantern/internal/postgresql/ast"
	"github.com/brianbroderick/lantern/internal/postgresql/lexer"
	"github.com/brianbroderick/lantern/internal/postgresql/parser"
	"github.com/brianbroderick/lantern/internal/postgresql/projectpath"
	"github.com/brianbroderick/lantern/pkg/repo"
	"github.com/brianbroderick/lantern/pkg/sql/logit"
)

func LoadLogFile(f string) string {
	path := filepath.Join(projectpath.Root, "logs", f)

	file, err := readPayload(path)
	if HasErr("processFile", err) {
		return ""
	}
	return string(file)
}

func AggregateLogs(fileName, log, queriesFile, databasesFile string) (*repo.Databases, *repo.Queries) {
	logit.Clear("queries-process-error")

	databases := repo.NewDatabases(fileName)
	statements := repo.NewQueries(fileName)

	l := lexer.New(log)
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

		timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%s %s", query.Date, query.Time), loadTz(query.Timezone))
		if HasErr("time.Parse", err) {
			continue loop
		}

		w := repo.QueryWorker{
			TimestampByHour: time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), 0, 0, 0, loadTz("UTC")),
			Databases:       databases,
			Database:        query.Database,
			Input:           query.Query,
			UserName:        query.User,
			DurationUs:      convertTime(query.DurationLit, query.DurationMeasure),
			MustExtract:     false, // We're passing in false into mustExtract because that'll happen at a later step
		}

		analyzed++
		if statements.Analyze(w) {
			success++
		}

		if analyzed%100000 == 0 {
			fmt.Printf("Parsed %7d successfully of %7d statements: %f%%\n", success, analyzed, (float64(success)/float64(analyzed))*100)
		}
	}

	fmt.Printf("Analyzed %7d of %7d statements\n", analyzed, len(program.Statements))
	fmt.Printf("Parsed %7d successfully of %7d statements: %f%%\n", success, analyzed, (float64(success)/float64(analyzed))*100)

	fmt.Printf("Number of statements: %d\n", len(statements.Queries))
	json := repo.MarshalJSON(statements)
	writeFile(filepath.Join(projectpath.Root, "processed", queriesFile), []byte(json))

	fmt.Printf("Number of databases: %d\n", len(databases.Databases))
	dbJSON := repo.MarshalJSON(databases)
	writeFile(filepath.Join(projectpath.Root, "processed", databasesFile), []byte(dbJSON))

	statements.LogAggregateOfErrors()

	return databases, statements
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

func loadTz(short string) *time.Location {
	var tz *time.Location
	var err error

	switch short {
	case "UTC":
		return time.UTC
	case "PDT", "PST":
		tz, err = time.LoadLocation("America/Los_Angeles")
		if HasErr("time.LoadLocation", err) {
			return time.UTC
		}
	case "EDT", "EST":
		tz, err = time.LoadLocation("America/New_York")
		if HasErr("time.LoadLocation", err) {
			return time.UTC
		}
	case "CDT", "CST":
		tz, err = time.LoadLocation("America/Chicago")
		if HasErr("time.LoadLocation", err) {
			return time.UTC
		}
	case "MDT", "MST":
		tz, err = time.LoadLocation("America/Denver")
		if HasErr("time.LoadLocation", err) {
			return time.UTC
		}
	}

	return tz
}
