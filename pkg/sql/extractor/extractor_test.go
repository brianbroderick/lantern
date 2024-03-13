package extractor

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/stretchr/testify/assert"
)

func TestExtractAlias(t *testing.T) {
	maskParams := false
	t1 := time.Now()

	tests := []struct {
		input  string
		output string
	}{
		{"select * from users u", "(SELECT * FROM users);"},
		{"select u.id from users u", "(SELECT users.id FROM users);"},
		{"select u.id, u.name from users u", "(SELECT users.id, users.name FROM users);"},
		{"select u.id, u.name as my_name from users u", "(SELECT users.id, users.name AS my_name FROM users);"},
		{"select u.id + 7 as my_alias from users u", "(SELECT (users.id + 7) AS my_alias FROM users);"},
		{"select u.* from users u;", "(SELECT users.* FROM users);"},
		{"select coalesce ( u.first_name || ' ' || u.last_name, u.first_name, u.last_name ) AS name from users u",
			"(SELECT coalesce(((users.first_name || ' ') || users.last_name), users.first_name, users.last_name) AS name FROM users);"}, // coalesce
		{"select c.id from customers c join addresses a on c.id = a.customer_id;",
			"(SELECT customers.id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id));"},
		{"select c.id from customers c join addresses a on (c.id = a.customer_id) join states s on (s.id = a.state_id);",
			"(SELECT customers.id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id) INNER JOIN states ON (states.id = addresses.state_id));"},
		{"select id from addresses AS a JOIN states AS s ON (s.id = a.state_id AND s.code > 'ut')",
			"(SELECT id FROM addresses INNER JOIN states ON ((states.id = addresses.state_id) AND (states.code > 'ut')));"},
		{"SELECT r.* FROM roles r, rights ri WHERE r.id = ri.role_id AND ri.deleted_by IS NULL AND ri.id = 12;",
			"(SELECT roles.* FROM roles , rights WHERE (((roles.id = rights.role_id) AND (rights.deleted_by IS NULL)) AND (rights.id = 12)));"},
		{"select left('abc', 2);", "(SELECT left('abc', 2));"}, // in this case, left is a string function getting the left 2 characters
		{"select u.category from users u where u.category is true", "(SELECT users.category FROM users WHERE (users.category IS TRUE));"},
		{"select s.baz from sales s group by s.bar;", "(SELECT sales.baz FROM sales GROUP BY sales.bar);"},
		{"select s.baz from sales s group by s.bar having s.baz > 5;", "(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5));"},
		{"select s.baz from sales s group by s.bar having s.baz > 5 order by s.baz;", "(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5) ORDER BY sales.baz);"},
		{"select s.baz from sales s group by s.bar having s.baz > 5 order by s.baz desc;", "(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5) ORDER BY sales.baz DESC);"},
		{"select sub.id from (select id from users) sub;", "(SELECT sub.id FROM (SELECT id FROM users) sub);"},
		{"select sub.id from (select u.id from users u) sub;", "(SELECT sub.id FROM (SELECT users.id FROM users) sub);"},
		{"select car.id from (select c.id from cars c) car join customers c on c.car_id = car.id;", "(SELECT car.id FROM (SELECT cars.id FROM cars) car INNER JOIN customers ON (customers.car_id = car.id));"},
		{"select u.id from users u where u.id = 42;", "(SELECT users.id FROM users WHERE (users.id = 42));"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		r := NewExtractor(program)
		r.Extract(r.Ast, env)
		checkExtractErrors(t, r, tt.input)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
	}
	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	fmt.Printf("TestExtractAlias, Elapsed Time: %s\n", timeDiff)
}

func checkExtractErrors(t *testing.T, r *Extractor, input string) {
	errors := r.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("input: %s\nExtractor has %d errors", input, len(errors))
	for _, msg := range errors {
		t.Errorf("Extractor error: %q", msg)
	}
	t.FailNow()
}
