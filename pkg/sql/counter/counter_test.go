package counter

// import (
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// )

// func TestProcessQuery(t *testing.T) {
// 	queries := NewQueries()
// 	t1 := time.Now()

// 	tests := []struct {
// 		input    string
// 		output   string
// 		duration int64
// 		sha      string
// 	}{
// 		{"select * from users where id = 42", "(SELECT * FROM users WHERE (id = ?));", 3, "b90be791fd20b7fc4925c087456e73493496364d"},
// 		{"select * from users where id = 74", "(SELECT * FROM users WHERE (id = ?));", 5, "b90be791fd20b7fc4925c087456e73493496364d"},
// 	}

// 	for _, tt := range tests {
// 		assert.True(t, queries.ProcessQuery(tt.input, tt.duration))
// 	}

// 	t2 := time.Now()

// 	assert.Equal(t, 1, len(queries.Queries))

// 	if _, ok := queries.Queries["b90be791fd20b7fc4925c087456e73493496364d"]; !ok {
// 		t.Fatalf("Expected to find sha in queries")
// 	}

// 	assert.Equal(t, int64(2), queries.Queries["b90be791fd20b7fc4925c087456e73493496364d"].TotalCount)
// 	assert.Equal(t, int64(8), queries.Queries["b90be791fd20b7fc4925c087456e73493496364d"].TotalDuration)

// 	timeDiff := t2.Sub(t1)
// 	avg := timeDiff / time.Duration(len(tests))
// 	fmt.Printf("TestProcessQuery, Elapsed Time: %s, Avg per query: %s\n", timeDiff, avg)
// }
