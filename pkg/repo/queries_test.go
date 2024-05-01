package repo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueriesIntegration(t *testing.T) {
	DBName = "lantern_test"

	sourceData := JsonSources()
	var sources Sources
	UnmarshalJSON([]byte(sourceData), &sources)
	assert.Equal(t, 1, len(sources.Sources))

	sources.Upsert()
	assert.Equal(t, 1, sources.CountInDB())

	databaseData := JsonDatabases()
	var databases Databases
	UnmarshalJSON([]byte(databaseData), &databases)
	assert.Equal(t, 2, len(databases.Databases))

	databases.Upsert()
	assert.Equal(t, 2, databases.CountInDB())

	// queries := NewQueries()
	queryData := JsonQueries()

	var statements Queries
	UnmarshalJSON([]byte(queryData), &statements)
	assert.Equal(t, 1, len(statements.Queries))

	statements.Upsert()
	assert.Equal(t, 1, statements.CountInDB())
}

func TestQueriesAnalyze(t *testing.T) {
	databases := NewDatabases()
	queries := NewQueries()
	t1 := time.Now()

	source := NewSource("testDB", "testDB")

	tests := []struct {
		input    string
		output   string
		duration int64
		uid      string
	}{
		{"select * from users where id = 42", "(SELECT * FROM users WHERE (id = ?));", 3, "a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"},
		{"select * from users where id = 74", "(SELECT * FROM users WHERE (id = ?));", 5, "a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"},
		{"select c.id from cars c where id = 19", "(SELECT cars.id FROM cars WHERE (id = ?));", 7, "8a8965e9-510a-502b-be8b-9aeb6e636616"},
	}

	for _, tt := range tests {
		w := QueryWorker{
			Databases:   databases,
			SourceUID:   source.UID,
			DatabaseUID: UuidFromString("fd68aa5c-a9c0-58db-a05f-13270c8c09dd"),
			Input:       tt.input,
			Duration:    tt.duration,
			MustExtract: false,
		}

		assert.True(t, queries.Analyze(w))
		assert.Equal(t, tt.output, queries.Queries[tt.uid].MaskedQuery)
	}

	t2 := time.Now()

	assert.Equal(t, 2, len(queries.Queries))

	if _, ok := queries.Queries["a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"]; !ok {
		t.Fatalf("Expected to find sha in queries")
	}

	assert.Equal(t, int64(2), queries.Queries["a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"].TotalCount)
	assert.Equal(t, int64(8), queries.Queries["a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"].TotalDuration)

	timeDiff := t2.Sub(t1)
	avg := timeDiff / time.Duration(len(tests))
	fmt.Printf("TestQueriesAnalyze, Elapsed Time: %s, Avg per query: %s\n", timeDiff, avg)
}

func JsonDatabases() string {
	return `{
			"databases": {
				"admin": "718d4687-8950-555b-a560-7b2795c4d2f3",
				"logs": "fd68aa5c-a9c0-58db-a05f-13270c8c09dd"				
			}
		}`
}

func JsonQueries() string {
	return `{
		"queries": {
		  "a2497c7b-dd5d-5be9-99b7-637eb8bacc4b": {
				"command": "SELECT",
				"uid": "a2497c7b-dd5d-5be9-99b7-637eb8bacc4b",
				"database_uid": "fd68aa5c-a9c0-58db-a05f-13270c8c09dd",
				"source_uid": "2aef85a0-30b5-5299-9635-35a6a553e0ef",
				"total_count": 2,
				"total_duration": 8,
				"masked_query": "(SELECT * FROM users WHERE (id = ?));",
				"unmasked_query": "(SELECT * FROM users WHERE (id = 42));",
				"source": "SELECT * FROM users WHERE id = 42;"
		  }
	  }
	}`
}

func JsonSources() string {
	return `{
		"sources": {
			"homepage": {
				"uid": "2aef85a0-30b5-5299-9635-35a6a553e0ef",
				"name": "homepage",				
				"url": "https://kludged.io"
			}
		}
	}`
}
