package counter

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/stretchr/testify/assert"
)

func TestMarshallQuery(t *testing.T) {
	q := &Query{
		Sha:           "913df750171f1aa8b2a6a40c94efe629545f21d3",
		Command:       token.SELECT,
		OriginalQuery: "select * from users where id = 42",
		MaskedQuery:   "(SELECT * FROM users WHERE (id = $1));",
		TotalCount:    2,
		TotalDuration: 8,
	}

	b, err := q.MarshalJSON()
	assert.NoError(t, err)

	newQuery := &Query{}

	err = newQuery.UnmarshalJSON(b)
	assert.NoError(t, err)

	assert.Equal(t, q.Sha, newQuery.Sha)
	assert.Equal(t, q.Command, newQuery.Command)
}

func TestMarshallQueries(t *testing.T) {
	q := &Queries{
		Queries: map[string]*Query{
			"b90be791fd20b7fc4925c087456e73493496364d": {
				Sha:           "b90be791fd20b7fc4925c087456e73493496364d",
				Command:       token.SELECT,
				OriginalQuery: "select * from users where id = 42",
				MaskedQuery:   "(SELECT * FROM users WHERE (id = ?));",
				TotalCount:    2,
				TotalDuration: 8,
			},
		},
	}

	newQueries := &Queries{}
	// Marshal and unmarshal the queries to a new struct, then check that the values are the same
	json := MarshalJSON(q)
	UnmarshalJSON([]byte(json), newQueries)

	assert.Equal(t, "b90be791fd20b7fc4925c087456e73493496364d", newQueries.Queries["b90be791fd20b7fc4925c087456e73493496364d"].Sha)
	assert.Equal(t, token.SELECT, newQueries.Queries["b90be791fd20b7fc4925c087456e73493496364d"].Command)
}
