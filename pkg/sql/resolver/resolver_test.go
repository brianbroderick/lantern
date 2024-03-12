package resolver

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/stretchr/testify/assert"
)

func TestResolveAlias(t *testing.T) {
	maskParams := false
	t1 := time.Now()

	tests := []struct {
		input  string
		output string
	}{
		{"select * from users u", "(SELECT * FROM users u);"},
		{"select u.id from users u", "(SELECT users.id FROM users u);"},
		{"select u.id, u.name from users u", "(SELECT users.id, users.name FROM users u);"},
		{"select u.id, u.name as my_name from users u", "(SELECT users.id, users.name AS my_name FROM users u);"},
		{"select u.id + 7 as my_alias from users u", "(SELECT (users.id + 7) AS my_alias FROM users u);"},
		{"select u.* from users u;", "(SELECT users.* FROM users u);"},
		{"select coalesce ( u.first_name || ' ' || u.last_name, u.first_name, u.last_name ) AS name from users u", "(SELECT coalesce(((users.first_name || ' ') || users.last_name), users.first_name, users.last_name) AS name FROM users u);"}, // coalesce
		{"select c.id from customers c join addresses a on c.id = a.customer_id;", "(SELECT customers.id FROM customers c INNER JOIN addresses a ON (customers.id = addresses.customer_id));"},
		{"select c.id from customers c join addresses a on (c.id = a.customer_id) join states s on (s.id = a.state_id);", "(SELECT customers.id FROM customers c INNER JOIN addresses a ON (customers.id = addresses.customer_id) INNER JOIN states s ON (states.id = addresses.state_id));"},
		{"select id from addresses AS a JOIN states AS s ON (s.id = a.state_id AND s.code > 'ut')", "(SELECT id FROM addresses a INNER JOIN states s ON ((states.id = addresses.state_id) AND (states.code > 'ut')));"},
		{"SELECT r.* FROM roles r, rights ri WHERE r.id = ri.role_id AND ri.deleted_by IS NULL AND ri.id = 12;", "(SELECT roles.* FROM roles r , rights ri WHERE (((roles.id = rights.role_id) AND (rights.deleted_by IS NULL)) AND (rights.id = 12)));"},
		{"select left('abc', 2);", "(SELECT left('abc', 2));"}, // in this case, left is a string function getting the left 2 characters
		{"select u.category from users u where u.category is true", "(SELECT users.category FROM users u WHERE (users.category IS TRUE));"},
		{"select s.baz from sales s group by s.bar;", "(SELECT sales.baz FROM sales s GROUP BY sales.bar);"},
		{"select s.baz from sales s group by s.bar having s.baz > 5;", "(SELECT sales.baz FROM sales s GROUP BY sales.bar HAVING (sales.baz > 5));"},
		{"select s.baz from sales s group by s.bar having s.baz > 5 order by s.baz;", "(SELECT sales.baz FROM sales s GROUP BY sales.bar HAVING (sales.baz > 5) ORDER BY sales.baz);"},
		{"select s.baz from sales s group by s.bar having s.baz > 5 order by s.baz desc;", "(SELECT sales.baz FROM sales s GROUP BY sales.bar HAVING (sales.baz > 5) ORDER BY sales.baz DESC);"},
		{"select sub.id from (select id from users) sub;", "(SELECT sub.id FROM (SELECT id FROM users) sub);"},
		{"select sub.id from (select u.id from users u) sub;", "(SELECT sub.id FROM (SELECT users.id FROM users u) sub);"},
		{"select car.id from (select c.id from cars c) car join customers c on c.car_id = car.id;", "(SELECT car.id FROM (SELECT cars.id FROM cars c) car INNER JOIN customers c ON (customers.car_id = car.id));"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		r := NewResolver(program)
		r.Resolve(r.Ast, env)
		checkResolveErrors(t, r, tt.input)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
	}
	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	fmt.Printf("TestResolveAlias, Elapsed Time: %s\n", timeDiff)
}

func checkResolveErrors(t *testing.T, r *Resolver, input string) {
	errors := r.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("input: %s\nresolver has %d errors", input, len(errors))
	for _, msg := range errors {
		t.Errorf("resolver error: %q", msg)
	}
	t.FailNow()
}
