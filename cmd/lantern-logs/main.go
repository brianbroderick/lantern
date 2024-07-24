package main

import "github.com/brianbroderick/lantern/internal/postgresql/logs"

func main() {
	logs.AggregateLogs()
	logs.UpsertQueries()
	logs.UpsertDatabases()
	logs.ExtractAndUpsertQueryMetadata()
}
