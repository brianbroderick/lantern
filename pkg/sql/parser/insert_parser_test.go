package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestInsertStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Select: simple
		{"insert into films values ('UA502', 'Bananas', 105, '1971-07-13', 'Comedy', '82 minutes');", "(INSERT INTO films VALUES ('UA502', 'Bananas', 105, '1971-07-13', 'Comedy', '82 minutes'));"},
		{"insert into films values ('UA502', 'Bananas', 105, DEFAULT, 'Comedy', '82 minutes');", "(INSERT INTO films VALUES ('UA502', 'Bananas', 105, DEFAULT, 'Comedy', '82 minutes'));"},
		{"insert into users (name, email) values ('Brian', 'foo@bar.com');", "(INSERT INTO users (name, email) VALUES ('Brian', 'foo@bar.com'));"},
		{"insert into users (name, email) values ('Brian', 'foo@bar.com'), ('Bob', 'bar@foo.com');", "(INSERT INTO users (name, email) VALUES ('Brian', 'foo@bar.com'), ('Bob', 'bar@foo.com'));"},
		{"insert into films default values;", "(INSERT INTO films DEFAULT VALUES);"},
		{"insert into products (name, price, product_no) values ('Cheese', 9.99, 1);", "(INSERT INTO products (name, price, product_no) VALUES ('Cheese', 9.99, 1));"},
		{"insert into products (product_no, name, price) select product_no, name, price from new_products where release_date = 'today';", "(INSERT INTO products (product_no, name, price) (SELECT product_no, name, price FROM new_products WHERE (release_date = 'today')));"},
		{"insert into films select * from tmp_films where date_prod < '2004-05-07';", "(INSERT INTO films (SELECT * FROM tmp_films WHERE (date_prod < '2004-05-07')));"},
		{"insert into distributors (did, dname) values (default, 'XYZ Widgets')	returning did;", "(INSERT INTO distributors (did, dname) VALUES (default, 'XYZ Widgets') RETURNING did);"},
		{"insert into distributors (did, dname) values (default, 'XYZ Widgets')	returning did, dname;", "(INSERT INTO distributors (did, dname) VALUES (default, 'XYZ Widgets') RETURNING did, dname);"},
		{"insert into distributors (did, dname) values (7, 'Redline GmbH') on conflict (did) do nothing;", "(INSERT INTO distributors (did, dname) VALUES (7, 'Redline GmbH') ON CONFLICT (did) DO NOTHING);"},
		{"insert into distributors (did, dname) values (7, 'Redline GmbH') on conflict (did) do update set dname = excluded.dname;", "(INSERT INTO distributors (did, dname) VALUES (7, 'Redline GmbH') ON CONFLICT (did) DO UPDATE SET (dname = excluded.dname));"},
		{"insert into people (id, fname, lname) values (42, 'Brian', 'Broderick') on conflict (id) do update set fname = excluded.fname, lname = excluded.lname;", "(INSERT INTO people (id, fname, lname) VALUES (42, 'Brian', 'Broderick') ON CONFLICT (id) DO UPDATE SET (fname = excluded.fname), (lname = excluded.lname));"},
		{"insert into people (id, fname, lname) values (42, 'Brian', 'Broderick') on conflict (id) do update set (fname = excluded.fname, lname = excluded.lname);", "(INSERT INTO people (id, fname, lname) VALUES (42, 'Brian', 'Broderick') ON CONFLICT (id) DO UPDATE SET ((fname = excluded.fname), (lname = excluded.lname)));"},
		{"insert into distributors as d (did, dname) values (8, 'Anvil Distribution') on conflict (did) do update set dname = excluded.dname where d.zipcode <> '21201';", "(INSERT INTO distributors AS d (did, dname) VALUES (8, 'Anvil Distribution') ON CONFLICT (did) DO UPDATE SET (dname = excluded.dname) WHERE (d.zipcode <> '21201'));"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt := program.Statements[0]
		assert.Equal(t, "INSERT", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.InsertStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.InsertStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.InsertStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
