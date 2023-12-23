package counter

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessQuery(t *testing.T) {
	queries := NewQueries()
	t1 := time.Now()

	tests := []struct {
		input    string
		output   string
		duration int64
		sha      string
	}{
		{"select * from users where id = 42", "(SELECT * FROM users WHERE (id = $1));", 3, "913df750171f1aa8b2a6a40c94efe629545f21d3"},
		{"select * from users where id = 74", "(SELECT * FROM users WHERE (id = $1));", 5, "913df750171f1aa8b2a6a40c94efe629545f21d3"},
	}

	for _, tt := range tests {
		assert.True(t, queries.ProcessQuery(tt.input, tt.duration))
	}

	t2 := time.Now()

	assert.Equal(t, 1, len(queries.Queries))
	assert.Equal(t, int64(2), queries.Queries["913df750171f1aa8b2a6a40c94efe629545f21d3"].TotalCount)
	assert.Equal(t, int64(8), queries.Queries["913df750171f1aa8b2a6a40c94efe629545f21d3"].TotalDuration)

	timeDiff := t2.Sub(t1)
	avg := timeDiff / time.Duration(len(tests))
	fmt.Printf("TestProcessQuery, Elapsed Time: %s, Avg per query: %s\n", timeDiff, avg)
}
