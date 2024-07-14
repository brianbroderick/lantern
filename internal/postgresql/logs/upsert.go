package logs

import (
	"os"
	"path/filepath"

	"github.com/brianbroderick/lantern/internal/postgresql/projectpath"
	"github.com/brianbroderick/lantern/pkg/repo"
)

func UpsertQueries() {
	data, err := os.ReadFile(filepath.Join(projectpath.Root, "processed", "queries.json"))
	if HasErr("Error reading file", err) {
		return
	}

	var statements repo.Queries
	repo.UnmarshalJSON(data, &statements)
	statements.Upsert()
}

func UpsertDatabases() {
	data, err := os.ReadFile(filepath.Join(projectpath.Root, "processed", "databases.json"))
	if HasErr("Error reading file", err) {
		return
	}
	var databases repo.Databases
	repo.UnmarshalJSON(data, &databases)
	db := repo.Conn()
	defer db.Close()
	databases.Upsert(db)
}

func ExtractAndUpsertQueryMetadata() {
	data, err := os.ReadFile(filepath.Join(projectpath.Root, "processed", "queries.json"))
	if HasErr("Error reading file", err) {
		return
	}

	queries := repo.NewQueries()

	jsonQueries := repo.Queries{}
	repo.UnmarshalJSON(data, &jsonQueries)
	queries.Queries = jsonQueries.Queries

	if len(queries.Queries) == 0 {
		return
	}

	queries.Process()
}
