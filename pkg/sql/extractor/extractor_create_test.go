package extractor

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/stretchr/testify/assert"
)

func TestExtractCreateStatements(t *testing.T) {
	t1 := time.Now()

	tests := []struct {
		input string
		name  string
	}{
		{"create temp table if not exists temp_my_table as ( select id from my_table );",
			"temp_my_table"},
		{"create temp table temp_my_table as (select id from my_table);",
			"temp_my_table"},
		{"create index idx_person_id ON temp_my_table( person_id );",
			"idx_person_id"},
		{"create INDEX idx_person_id ON temp_my_table( person_id );",
			"idx_person_id"},
		{"CREATE INDEX idx_temp_person on pg_temp.people using btree ( account_id, person_id );",
			"idx_temp_person"},
		{"create temp table temp_my_table on commit drop as (select id from users);",
			"temp_my_table"},
		// TODO: This is a bug, the name should be temp_my_table
		{"create temp table temp_my_table( like my_reports );",
			"temp_my_table((LIKE my_reports))"},
		{"create index idx_temp_stuff ON my_temp USING btree( id, temp_id ) WHERE ( blah_id = 1 );",
			"idx_temp_stuff"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		for _, s := range program.Statements {
			ext := NewExtractor(&s, true)
			ext.Execute(s)
			checkExtractErrors(t, ext, tt.input)

			for _, c := range ext.CreateStatements {
				fqcs := fmt.Sprintf("%s|%t|%t|%t|%s|%s|%s|%s",
					c.Scope, c.IsUnique, c.IsTemp, c.IsUnlogged, c.ObjectType, c.Name, c.Expression, c.WhereClause)
				uid := UuidV5(fqcs)

				assert.Equal(t, uid, c.UID, "input: %s\nCreate Statement UID %s not found in %s", tt.input, c.UID.String(), uid.String())
				assert.Equal(t, tt.name, c.Name, "input: %s\nCreate Statement Name %s not found in %v", tt.input, c.Name, fqcs)
			}
		}
	}

	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	fmt.Printf("TestExtractCreateStatements, Elapsed Time: %s\n", timeDiff)
}
