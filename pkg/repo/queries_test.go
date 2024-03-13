package repo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessQuery(t *testing.T) {
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
	}

	for _, tt := range tests {
		assert.True(t, queries.ProcessQuery(databases, source, "testDB", tt.input, tt.duration))
	}

	t2 := time.Now()

	assert.Equal(t, 1, len(queries.Queries))

	if _, ok := queries.Queries["a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"]; !ok {
		t.Fatalf("Expected to find sha in queries")
	}

	assert.Equal(t, int64(2), queries.Queries["a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"].TotalCount)
	assert.Equal(t, int64(8), queries.Queries["a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"].TotalDuration)

	timeDiff := t2.Sub(t1)
	avg := timeDiff / time.Duration(len(tests))
	fmt.Printf("TestProcessQuery, Elapsed Time: %s, Avg per query: %s\n", timeDiff, avg)
}
