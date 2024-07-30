package main

import (
	"fmt"

	"github.com/brianbroderick/lantern/internal/postgresql/logs"
)

func main() {
	// TODO: make the file name dynamic
	f := "postgresql.log.2024-07-10-1748.cp"

	fmt.Println("Processing log file", f)

	log := logs.LoadLogFile(f)
	logs.AggregateLogs(log, "queries.json", "databases.json")

	logs.UpsertQueries()
	logs.UpsertDatabases()
	logs.ExtractAndUpsertQueryMetadata()
}
