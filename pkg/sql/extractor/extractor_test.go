package extractor

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/token"
	"github.com/stretchr/testify/assert"
)

func TestExtractAlias(t *testing.T) {
	maskParams := false
	t1 := time.Now()

	tests := []struct {
		input   string
		output  string
		columns [][]string
		tables  [][]string
		joins   [][]string
	}{
		{"select coalesce(id, name) from users", "(SELECT coalesce(id, name) FROM users);",
			[][]string{{"id", "name"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select * from users; select * from addresses", "(SELECT * FROM users);(SELECT * FROM addresses);",
			[][]string{{"users"}, {"addresses"}}, [][]string{{"users"}, {"addresses"}}, [][]string{{"users"}, {"addresses"}}},
		{"select * from users u", "(SELECT * FROM users);",
			[][]string{{"users.*"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select u.id from users u", "(SELECT users.id FROM users);",
			[][]string{{"users.id"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select u.id, u.name from users u", "(SELECT users.id, users.name FROM users);",
			[][]string{{"users.id", "users.name"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select u.id, u.name as my_name from users u", "(SELECT users.id, users.name AS my_name FROM users);",
			[][]string{{"users.id", "users.name"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select u.id + 7 as my_alias from users u", "(SELECT (users.id + 7) AS my_alias FROM users);",
			[][]string{{"users.id"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select u.* from users u;", "(SELECT users.* FROM users);",
			[][]string{{"users.*"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select coalesce ( u.first_name || ' ' || u.last_name, u.first_name, u.last_name ) AS name from users u",
			"(SELECT coalesce(((users.first_name || ' ') || users.last_name), users.first_name, users.last_name) AS name FROM users);",
			[][]string{{"users.first_name", "users.last_name"}}, [][]string{{"users"}}, [][]string{{}}}, // coalesce
		{"select c.id from customers c join addresses a on c.id = a.customer_id;",
			"(SELECT customers.id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id));",
			[][]string{{"customers.id"}}, [][]string{{"customers", "addresses"}}, [][]string{{"customers", "addresses"}}},
		{"select c.id from customers c join addresses a on (c.id = a.customer_id) join states s on (s.id = a.state_id);",
			"(SELECT customers.id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id) INNER JOIN states ON (states.id = addresses.state_id));",
			[][]string{{"customers.id"}}, [][]string{{"customers", "addresses", "states"}}, [][]string{{"customers", "addresses", "states"}}},
		{"select id from addresses AS a JOIN states AS s ON (s.id = a.state_id AND s.code > 'ut')",
			"(SELECT id FROM addresses INNER JOIN states ON ((states.id = addresses.state_id) AND (states.code > 'ut')));",
			[][]string{{"id"}}, [][]string{{"addresses", "states"}}, [][]string{{"states", "addresses"}}},
		{"SELECT r.* FROM roles r, rights ri WHERE r.id = ri.role_id AND ri.deleted_by IS NULL AND ri.id = 12;",
			"(SELECT roles.* FROM roles , rights WHERE (((roles.id = rights.role_id) AND (rights.deleted_by IS NULL)) AND (rights.id = 12)));",
			[][]string{{"roles.*"}}, [][]string{{"roles", "rights"}}, [][]string{{}}},
		{"select left('abc', 2);", "(SELECT left('abc', 2));",
			[][]string{{}}, [][]string{{}}, [][]string{{}}}, // in this case, left is a string function getting the left 2 characters
		{"select u.category from users u where u.category is true", "(SELECT users.category FROM users WHERE (users.category IS TRUE));",
			[][]string{{"users.category"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select s.baz from sales s group by s.bar;", "(SELECT sales.baz FROM sales GROUP BY sales.bar);",
			[][]string{{"sales.baz"}}, [][]string{{"sales"}}, [][]string{{}}},
		{"select s.baz from sales s group by s.bar having s.baz > 5;",
			"(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5));",
			[][]string{{"sales.baz"}}, [][]string{{"sales"}}, [][]string{{}}},
		{"select s.baz from sales s group by s.bar having s.baz > 5 order by s.baz;",
			"(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5) ORDER BY sales.baz);",
			[][]string{{"sales.baz"}}, [][]string{{"sales"}}, [][]string{{}}},
		{"select s.baz from sales s group by s.bar having s.baz > 5 order by s.baz desc;",
			"(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5) ORDER BY sales.baz DESC);",
			[][]string{{"sales.baz"}}, [][]string{{"sales"}}, [][]string{{}}},
		{"select sub.id from (select id from users) sub;", "(SELECT sub.id FROM (SELECT id FROM users) sub);",
			[][]string{{"sub.id", "id"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select sub.id from (select u.id from users u) sub;", "(SELECT sub.id FROM (SELECT users.id FROM users) sub);",
			[][]string{{"sub.id", "users.id"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select car.id from (select c.id from cars c) car join customers c on c.car_id = car.id;",
			"(SELECT car.id FROM (SELECT cars.id FROM cars) car INNER JOIN customers ON (customers.car_id = car.id));",
			[][]string{{"car.id", "cars.id"}}, [][]string{{"cars", "customers"}}, [][]string{{"customers", "car"}}},
		{"select u.id from users u where u.id = 42;", "(SELECT users.id FROM users WHERE (users.id = 42));",
			[][]string{{"users.id"}}, [][]string{{"users"}}, [][]string{{}}},
		{"select car.id from (select c.id from cars c) car join customers c on c.car_id = car.id inner join users u on u.id = c.user_id;",
			"(SELECT car.id FROM (SELECT cars.id FROM cars) car INNER JOIN customers ON (customers.car_id = car.id) INNER JOIN users ON (users.id = customers.user_id));",
			[][]string{{"car.id", "cars.id"}}, [][]string{{"cars", "customers", "users"}}, [][]string{{"customers", "car", "users"}}},
	}

	// TODO: add table to non-aliased columns
	// TODO: handle wildcards better

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		for i, s := range program.Statements {
			r := NewExtractor(&s)
			env := object.NewEnvironment()
			r.Extract(s, env)
			checkExtractErrors(t, r, tt.input)

			assert.Equal(t, len(tt.tables[i]), len(r.Tables), "input: %s\nNumber of tables not equal", tt.input)

			for _, table := range r.Tables {
				assert.Contains(t, tt.tables[i], table.Name, "input: %s\nTable %s not found in %v", tt.input, table.Name, tt.tables[i])
			}

			for fqcn, column := range r.Columns {
				if column.Clause == token.COLUMN {
					assert.Contains(t, tt.columns[i], fqcn, "input: %s\nColumn %s not found in %v", tt.input, fqcn, tt.columns[i])
				}
			}

			for _, join := range r.TableJoins {
				found := 0
				for _, j := range tt.joins[i] {
					if join.TableA == j {
						found++
					}
					if join.TableB == j {
						found++
					}
				}
				assert.Equal(t, 2, found, "input: %s\nDid not find all input tables: %v in join: %s %s\n",
					tt.input, tt.joins[i], printTableJoins(r))
			}

		}

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s\n", tt.input, tt.output, output)

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

func printTableJoins(r *Extractor) string {
	var joins []string
	for _, join := range r.TableJoins {
		joins = append(joins, fmt.Sprintf("%s %s", join.TableA, join.TableB))
	}
	return fmt.Sprintf("%v", joins)
}
