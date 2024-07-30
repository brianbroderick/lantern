package main

import (
	"fmt"

	"github.com/brianbroderick/lantern/internal/postgresql/logs"
)

func main() {
	// TODO: make the file name dynamic
	f := "postgresql.log.2024-07-10-1748.cp"
	log := logs.LoadLogFile(f)

	fmt.Println("Processing log file, length:", len(log))

	logs.AggregateLogs(log, "queries.json", "databases.json")

	logs.UpsertQueries()
	logs.UpsertDatabases()
	logs.ExtractAndUpsertQueryMetadata()
}
