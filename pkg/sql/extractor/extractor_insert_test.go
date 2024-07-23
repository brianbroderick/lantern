package extractor

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/stretchr/testify/assert"
)

func TestExtractInsertStatements(t *testing.T) {
	t1 := time.Now()

	tests := []struct {
		input  string
		tables [][]string
	}{
		{"insert into films values ('UA502', 'Bananas', 105, '1971-07-13', 'Comedy', '82 minutes');",
			[][]string{{"public.films"}}},
		{"insert into films values ('UA502', 'Bananas', 105, DEFAULT, 'Comedy', '82 minutes');",
			[][]string{{"public.films"}}},
		{"insert into users (name, email) values ('Brian', 'foo@bar.com');",
			[][]string{{"public.users"}}},
		{"insert into users (name, email) values ('Brian', 'foo@bar.com'), ('Bob', 'bar@foo.com');",
			[][]string{{"public.users"}}},
		{"insert into films default values;",
			[][]string{{"public.films"}}},
		{"insert into products (name, price, product_no) values ('Cheese', 9.99, 1);",
			[][]string{{"public.products"}}},
		// TODO: Need to extract if tables are being read or written to (which is an insert vs a select)
		{"insert into products (product_no, name, price) select product_no, name, price from new_products where release_date = 'today';",
			[][]string{{"public.products"}, {"public.new_products"}}},
		{"insert into films select * from tmp_films where date_prod < '2004-05-07';",
			[][]string{{"public.films"}, {"public.tmp_films"}}},
		{"insert into distributors (did, dname) values (default, 'XYZ Widgets')	returning did;",
			[][]string{{"public.distributors"}}},
		{"insert into distributors (did, dname) values (default, 'XYZ Widgets')	returning did, dname;",
			[][]string{{"public.distributors"}}},
		{"insert into distributors (did, dname) values (7, 'Redline GmbH') on conflict (did) do nothing;",
			[][]string{{"public.distributors"}}},
		{"insert into distributors (did, dname) values (7, 'Redline GmbH') on conflict (did) do update set dname = excluded.dname;",
			[][]string{{"public.distributors"}}},
		{"insert into people (id, fname, lname) values (42, 'Brian', 'Broderick') on conflict (id) do update set fname = excluded.fname, lname = excluded.lname;",
			[][]string{{"public.people"}}},
		{"insert into people (id, fname, lname) values (42, 'Brian', 'Broderick') on conflict (id) do update set (fname = excluded.fname, lname = excluded.lname);",
			[][]string{{"public.people"}}},
		{"insert into distributors as d (did, dname) values (8, 'Anvil Distribution') on conflict (did) do update set dname = excluded.dname where d.zipcode <> '21201';",
			[][]string{{"public.distributors"}}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		for i, s := range program.Statements {
			r := NewExtractor(&s, true)
			r.Execute(s)
			checkExtractErrors(t, r, tt.input)

			assert.Equal(t, len(tt.tables[i]), len(r.TablesInQueries), "input: %s\nNumber of tables not equal", tt.input)

			for _, table := range r.TablesInQueries {
				assert.Contains(t, tt.tables[i], fmt.Sprintf("%s.%s", table.Schema, table.Name), "input: %s\nTable %s not found in %v", tt.input, table.Name, tt.tables[i])
			}
		}
	}

	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	fmt.Printf("TestExtractInsertStatements, Elapsed Time: %s\n", timeDiff)
}
