package parser

import (
	"fmt"
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestMultipleStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input          string
		statementCount int
		output         string
	}{
		// Multiple Statements
		{"select id from users; select id from customers;", 2, "(SELECT id FROM users);(SELECT id FROM customers);"},
		{"select u.id from users u;", 1, "(SELECT u.id FROM users u);"},
	}

	for _, tt := range tests {
		fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, tt.statementCount, len(program.Statements), "program.Statements does not contain %d statements. got=%d\n", tt.statementCount, len(program.Statements))

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "program.String() not '%s'. got=%s", tt.output, output)
		fmt.Printf("output: %s\n", output)
	}
}

func TestSingleSelectStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input      string
		tableCount int
		output     string
	}{
		// Select: simple
		{"select id from users;", 1, "(SELECT id FROM users);"},                                                          // super basic select
		{"select u.* from users u;", 1, "(SELECT u.* FROM users u);"},                                                    // check for a wildcard with a table alias
		{"select 2*3 from users;", 1, "(SELECT (2 * 3) FROM users);"},                                                    // check that the asterisk is not treated as a wildcard
		{"select \"blah\".id from users", 1, "(SELECT blah.id FROM users);"},                                             // check for double quotes around the table name
		{"select 1 * (2 + (6 / 4)) - 9 from users;", 1, "(SELECT ((1 * (2 + (6 / 4))) - 9) FROM users);"},                // math expression
		{"select id, name from users", 1, "(SELECT id, name FROM users);"},                                               // multiple columns
		{"select id, first_name from users;", 1, "(SELECT id, first_name FROM users);"},                                  // underscore in a column name
		{"select id, first_name as name from users", 1, "(SELECT id, first_name AS name FROM users);"},                   // column alias
		{"select u.id, u.first_name as name from users u;", 1, "(SELECT u.id, u.first_name AS name FROM users u);"},      // column alias with table alias
		{"select id from no_semi_colons", 1, "(SELECT id FROM no_semi_colons);"},                                         // no semicolon
		{"select 1 + 2 as math, foo + 7 as seven from foo", 1, "(SELECT (1 + 2) AS math, (foo + 7) AS seven FROM foo);"}, // multiple column aliases with expressions
		{"select 1 + 2 * 3 / value as math from foo", 1, "(SELECT (1 + ((2 * 3) / value)) AS math FROM foo);"},           // more complex math expression
		{"select id from addresses a;", 1, "(SELECT id FROM addresses a);"},                                              // table alias
		{"select sum(a,b) from users;", 1, "(SELECT sum(a, b) FROM users);"},                                             // function call

		// Select: distinct & all tokens
		{"select distinct id from users;", 1, "(SELECT DISTINCT id FROM users);"},
		{"select all id from users", 1, "(SELECT ALL id FROM users);"},
		{"select distinct on (location) reported_at, report from weather_reports;", 1, "(SELECT DISTINCT ON (location) reported_at, report FROM weather_reports);"},

		// Select: window functions
		{"select avg(salary) over (partition by depname) from empsalary;", 1, "(SELECT (avg(salary) OVER (PARTITION BY depname)) FROM empsalary);"},
		{"select avg(salary) over (order by depname) from empsalary", 1, "(SELECT (avg(salary) OVER (ORDER BY depname)) FROM empsalary);"},
		{"select avg(salary) over (partition by salary order by depname) from empsalary;", 1, "(SELECT (avg(salary) OVER (PARTITION BY salary ORDER BY depname)) FROM empsalary);"},
		{"select avg(salary) over (partition by salary order by depname desc) from empsalary", 1, "(SELECT (avg(salary) OVER (PARTITION BY salary ORDER BY depname DESC)) FROM empsalary);"},
		{"select wf1() over w from table_name;", 1, "(SELECT (wf1() OVER w) FROM table_name);"},
		{"select wf1() over w, wf2() over w from table_name;", 1, "(SELECT (wf1() OVER w), (wf2() OVER w) FROM table_name);"},
		{"select wf1() over w, wf2() over w from table_name window w as (partition by c1 order by c2);", 1, "(SELECT (wf1() OVER w), (wf2() OVER w) FROM table_name WINDOW w AS (PARTITION BY c1 ORDER BY c2));"},
		{"select wf1() over w, wf2() over w from table_name window w as (partition by c1 order by c2), foo as (partition by c3 order by c4);", 1, "(SELECT (wf1() OVER w), (wf2() OVER w) FROM table_name WINDOW w AS (PARTITION BY c1 ORDER BY c2), foo AS (PARTITION BY c3 ORDER BY c4));"},

		// Select: joins
		{"select c.id from customers c join addresses a on c.id = a.customer_id;", 2, "(SELECT c.id FROM customers c INNER JOIN addresses a ON (c.id = a.customer_id));"},
		{"select id from customers join addresses on id = customer_id;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id));"},
		{"select id from customers join addresses on id = customer_id join phones on id = phone_id;", 3, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) INNER JOIN phones ON (id = phone_id));"},
		{"select id from customers join addresses on customers.id = addresses.customer_id", 2, "(SELECT id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id));"},
		{"select id from customers join addresses on id = customer_id where id = 46;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = 46));"},
		{"select id from customers left join addresses on id = customer_id;", 2, "(SELECT id FROM customers LEFT JOIN addresses ON (id = customer_id));"},
		{"select id from customers left outer join addresses on id = customer_id;", 2, "(SELECT id FROM customers LEFT JOIN addresses ON (id = customer_id));"},

		// Select: where clause
		{"select id from users where id = 42;", 1, "(SELECT id FROM users WHERE (id = 42));"},
		{"select id from users where id = 42 and customer_id = 74", 1, "(SELECT id FROM users WHERE ((id = 42) AND (customer_id = 74)));"},
		{"select id from users where id = 42 and customer_id > 74;", 1, "(SELECT id FROM users WHERE ((id = 42) AND (customer_id > 74)));"},
		{"select id from users where name = 'brian';", 1, "(SELECT id FROM users WHERE (name = 'brian'));"},
		{"select id from users where name = 'brian'", 1, "(SELECT id FROM users WHERE (name = 'brian'));"},

		// Select: group by
		{"select id from users group by id", 1, "(SELECT id FROM users GROUP BY id);"},
		{"select id from users group by id, name;", 1, "(SELECT id FROM users GROUP BY id, name);"},

		// Select: combined clauses
		{"select id from users where id = 42 group by id, name", 1, "(SELECT id FROM users WHERE (id = 42) GROUP BY id, name);"},
		{"select id from customers join addresses on id = customer_id where id = 46 group by id;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = 46) GROUP BY id);"},

		// Select: having clause
		{"select id from users group by id having id > 2;", 1, "(SELECT id FROM users GROUP BY id HAVING (id > 2));"},
		{"select id from users group by id having id > 2 and name = 'frodo';", 1, "(SELECT id FROM users GROUP BY id HAVING ((id > 2) AND (name = 'frodo')));"},

		// Select: order by
		{"select id from users order by id;", 1, "(SELECT id FROM users ORDER BY id);"},
		{"select id from users order by id desc, name", 1, "(SELECT id FROM users ORDER BY id DESC, name);"},
		{"select id from users order by id desc nulls first, name nulls last;", 1, "(SELECT id FROM users ORDER BY id DESC NULLS FIRST, name NULLS LAST);"},

		// Select: limit
		{"select id from users limit 10;", 1, "(SELECT id FROM users LIMIT 10);"},
		{"select id from users limit ALL;", 1, "(SELECT id FROM users LIMIT ALL);"},
		{"select id from users limit ALL", 1, "(SELECT id FROM users LIMIT ALL);"},

		// Select: offset
		{"select id from users limit ALL offset 10;", 1, "(SELECT id FROM users LIMIT ALL OFFSET 10);"},
		{"select id from users limit 10 offset 10;", 1, "(SELECT id FROM users LIMIT 10 OFFSET 10);"},
		{"select id from users limit 10 offset 1 ROW", 1, "(SELECT id FROM users LIMIT 10 OFFSET 1);"},
		{"select id from users limit 10 offset 2 ROWS;", 1, "(SELECT id FROM users LIMIT 10 OFFSET 2);"},

		// Select: combined order by, limit, offset
		{"select id from users order by id desc limit 10 offset 10;", 1, "(SELECT id FROM users ORDER BY id DESC LIMIT 10 OFFSET 10);"},
		{"select id from users order by id desc nulls last limit 10 offset 10;", 1, "(SELECT id FROM users ORDER BY id DESC NULLS LAST LIMIT 10 OFFSET 10);"},

		// Select: fetch
		{"select id from users order by id fetch first row only;", 1, "(SELECT id FROM users ORDER BY id FETCH NEXT 1 ROWS ONLY);"},
		{"select id from users order by id fetch first 3 rows only;", 1, "(SELECT id FROM users ORDER BY id FETCH NEXT 3 ROWS ONLY);"},
		{"select id from users order by id fetch first 10 rows with ties;", 1, "(SELECT id FROM users ORDER BY id FETCH NEXT 10 ROWS WITH TIES);"},

		// Select: for update
		{"select id from users for update;", 1, "(SELECT id FROM users FOR UPDATE);"},
		{"select id from users for no key update;;", 1, "(SELECT id FROM users FOR NO KEY UPDATE);"},
		{"select id from users for share;", 1, "(SELECT id FROM users FOR SHARE);"},
		{"select id from users for key share", 1, "(SELECT id FROM users FOR KEY SHARE);"},
		{"select id from users for update of users;", 1, "(SELECT id FROM users FOR UPDATE OF users);"},
		{"select id from users for update of users, addresses;", 1, "(SELECT id FROM users FOR UPDATE OF users, addresses);"},
		{"select id from users for update of users, addresses nowait;", 1, "(SELECT id FROM users FOR UPDATE OF users, addresses NOWAIT);"},
		{"select id from users for update of users, addresses skip locked;", 1, "(SELECT id FROM users FOR UPDATE OF users, addresses SKIP LOCKED);"},

		// Select: IN clause
		{"select id from users where id IN ('1','2','3','4');", 1, "(SELECT id FROM users WHERE id IN ('1', '2', '3', '4'));"},
		{"select id from users where id IN ('1','2','3','4') AND name = 'brian';", 1, "(SELECT id FROM users WHERE (id IN ('1', '2', '3', '4') AND (name = 'brian')));"},
		{"select id from users where id IN (1,2,3,4);", 1, "(SELECT id FROM users WHERE id IN (1, 2, 3, 4));"},

		// Select: UNION clause
		{"select id from users union select id from customers;", 1, "(SELECT id FROM users) UNION (SELECT id FROM customers);"},
		{"select id from users except select id from customers;", 1, "(SELECT id FROM users) EXCEPT (SELECT id FROM customers);"},
		{"select id from users intersect select id from customers;", 1, "(SELECT id FROM users) INTERSECT (SELECT id FROM customers);"},

		// Select: Cast literals
		{"select '100'::integer from a;", 1, "(SELECT '100'::INTEGER FROM a);"},
		{"select 100::text from a;", 1, "(SELECT 100::TEXT FROM a);"},
		{"select a::text from b;", 1, "(SELECT a::TEXT FROM b);"},

		// Select: JSONB
		{"select id from users where data->'name' = 'brian';", 1, "(SELECT id FROM users WHERE ((data -> 'name') = 'brian'));"},
		{"select id from users where data->>'name' = 'brian';", 1, "(SELECT id FROM users WHERE ((data ->> 'name') = 'brian'));"},
		{"select id from users where data#>'{name}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #> '{name}') = 'brian'));"},
		{"select id from users where data#>>'{name}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #>> '{name}') = 'brian'));"},
		{"select id from users where data#>>'{name,first}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #>> '{name,first}') = 'brian'));"},
		{"select id from users where data#>>'{name,first}' = 'brian' and data#>>'{name,last}' = 'broderick';", 1, "(SELECT id FROM users WHERE (((data #>> '{name,first}') = 'brian') AND ((data #>> '{name,last}') = 'broderick')));"},
		{"select * from users where metadata @> '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata @> '{\"age\": 42}'));"},
		{"select * from users where metadata <@ '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata <@ '{\"age\": 42}'));"},
		{"select * from users where metadata ? '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ? '{\"age\": 42}'));"},
		{"select * from users where metadata ?| '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ?| '{\"age\": 42}'));"},
		{"select * from users where metadata ?& '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ?& '{\"age\": 42}'));"},
		{"select * from users where metadata || '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata || '{\"age\": 42}'));"},

		// Select: ARRAYs
		{"select array['John'] from users;", 1, "(SELECT array['John'] FROM users);"},
		{"select array['John', 'Joseph'] from users;", 1, "(SELECT array['John', 'Joseph'] FROM users);"},
		{"select array['John', 'Joseph', 'Anna', 'Henry'] && array['Henry', 'John'] from users;", 1, "(SELECT (array['John', 'Joseph', 'Anna', 'Henry'] && array['Henry', 'John']) FROM users);"},
	}

	for _, tt := range tests {
		fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements), "program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "select", stmt.TokenLiteral(), "program.Statements[0] is not ast.SelectStatement. got=%T", stmt)

		selectStmt, ok := stmt.(*ast.SelectStatement)
		assert.True(t, ok, "stmt is not *ast.SelectStatement. got=%T", stmt)

		selectExp, ok := selectStmt.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "stmt is not *ast.SelectExpression. got=%T", selectExp)

		assert.Equal(t, tt.tableCount, len(selectExp.Tables), "len(selectStmt.Tables) not %d. got=%d", tt.tableCount, len(selectExp.Tables))
		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "program.String() not '%s'. got=%s", tt.output, output)
		fmt.Printf("output: %s\n", output)
	}
}

func TestSubSelects(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Select: Subqueries
		{"select film_id, title, rental_rate from film where rental_rate > (select avg(rental_rate) from film);", "(SELECT film_id, title, rental_rate FROM film WHERE (rental_rate > (SELECT avg(rental_rate) FROM film)));"},
		{"select id from customers where id > (select avg(id) from customers where id > (select min(id) from customers));", "(SELECT id FROM customers WHERE (id > (SELECT avg(id) FROM customers WHERE (id > (SELECT min(id) FROM customers)))));"},
		{"select film_id from film where rental_rate > (select avg(rental_rate) from film) order by film_id;", "(SELECT film_id FROM film WHERE (rental_rate > (SELECT avg(rental_rate) FROM film)) ORDER BY film_id);"},
		{"select id from customers where id > (select avg(id) from customers where id > 42) order by id desc;", "(SELECT id FROM customers WHERE (id > (SELECT avg(id) FROM customers WHERE (id > 42))) ORDER BY id DESC);"},
		{"select one from table_one where one > (select two from table_two where two > (select three from table_three)) order by one desc;", "(SELECT one FROM table_one WHERE (one > (SELECT two FROM table_two WHERE (two > (SELECT three FROM table_three)))) ORDER BY one DESC);"},
	}

	for _, tt := range tests {
		fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements), "program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "select", stmt.TokenLiteral(), "program.Statements[0] is not ast.SelectStatement. got=%T", stmt)

		selectStmt, ok := stmt.(*ast.SelectStatement)
		assert.True(t, ok, "stmt is not *ast.SelectStatement. got=%T", stmt)

		selectExp, ok := selectStmt.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "stmt is not *ast.SelectExpression. got=%T", selectExp)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "program.String() not '%s'. got=%s", tt.output, output)
		fmt.Printf("output: %s\n", output)
	}
}

func TestCTEs(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Select: Subqueries
		{"with sales as (select sum(amount) as total_sales from orders) select total_sales from sales;", "(WITH sales AS (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"},
		{"with regional_sales as (select region, sum(amount) as total_sales from orders group by region), top_regions AS (select region from regional_sales where total_sales > 42)	select region, product, sum(quantity) AS product_units, sum(amount) as product_sales from orders group by region, product;",
			"(WITH regional_sales AS (SELECT region, sum(amount) AS total_sales FROM orders GROUP BY region), top_regions AS (SELECT region FROM regional_sales WHERE (total_sales > 42)) (SELECT region, product, sum(quantity) AS product_units, sum(amount) AS product_sales FROM orders GROUP BY region, product));"},
		{"with regional_sales as (select region, sum(amount) as total_sales from orders group by region), top_regions AS (select region from regional_sales where total_sales > (select sum(total_sales)/10 from regional_sales))	select region, product, sum(quantity) AS product_units, sum(amount) as product_sales from orders group by region, product;",
			"(WITH regional_sales AS (SELECT region, sum(amount) AS total_sales FROM orders GROUP BY region), top_regions AS (SELECT region FROM regional_sales WHERE (total_sales > (SELECT (sum(total_sales) / 10) FROM regional_sales))) (SELECT region, product, sum(quantity) AS product_units, sum(amount) AS product_sales FROM orders GROUP BY region, product));"},
		{"with regional_sales as (select region, sum(amount) as total_sales from orders group by region), top_regions AS (select region from regional_sales where total_sales > (select sum(total_sales)/10 from regional_sales))	select region, product, sum(quantity) AS product_units, sum(amount) as product_sales from orders where region in (SELECT region from top_regions) group by region, product;",
			"(WITH regional_sales AS (SELECT region, sum(amount) AS total_sales FROM orders GROUP BY region), top_regions AS (SELECT region FROM regional_sales WHERE (total_sales > (SELECT (sum(total_sales) / 10) FROM regional_sales))) (SELECT region, product, sum(quantity) AS product_units, sum(amount) AS product_sales FROM orders WHERE region IN ((SELECT region FROM top_regions)) GROUP BY region, product));"},
		{"with recursive sales as (select sum(amount) as total_sales from orders) select total_sales from sales;", "(WITH RECURSIVE sales AS (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"},
		{"with sales as materialized (select sum(amount) as total_sales from orders) select total_sales from sales;", "(WITH sales AS MATERIALIZED (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"},
		{"with sales as not materialized (select sum(amount) as total_sales from orders) select total_sales from sales;", "(WITH sales AS NOT MATERIALIZED (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"},
	}

	for _, tt := range tests {
		fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements), "program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "with", stmt.TokenLiteral(), "program.Statements[0] is not ast.CTEStatement. got=%T", stmt)

		cteStmt, ok := stmt.(*ast.CTEStatement)
		assert.True(t, ok, "stmt is not *ast.CTEStatement. got=%T", stmt)

		cteExp, ok := cteStmt.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "stmt is not *ast.SelectExpression. got=%T", cteExp)

		// program.Statements[0].Inspect()

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "program.String() not '%s'. got=%s", tt.output, output)
		fmt.Printf("output: %s\n", output)
	}
}

func TestMaskParams(t *testing.T) {
	maskParams := true

	tests := []struct {
		input      string
		tableCount int
		output     string
	}{
		// Some Selects
		{"select id from users;", 1, "(SELECT id FROM users);"},
		{"select 1 * (2 + (6 / 4)) - 9 from users;", 1, "(SELECT (($1 * ($2 + ($3 / $4))) - $5) FROM users);"},
		{"select id from customers join addresses on id = customer_id where id = 46;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = $1));"},

		// Select: where clause
		{"select id from users where id = 42;", 1, "(SELECT id FROM users WHERE (id = $1));"},
		{"select id from users where id = 42 and customer_id = 74", 1, "(SELECT id FROM users WHERE ((id = $1) AND (customer_id = $2)));"},
		{"select id from users where id = 42 and customer_id > 74;", 1, "(SELECT id FROM users WHERE ((id = $1) AND (customer_id > $2)));"},
		{"select id from users where name = 'brian';", 1, "(SELECT id FROM users WHERE (name = '$1'));"},
		{"select id from users where name = 'brian'", 1, "(SELECT id FROM users WHERE (name = '$1'));"},

		// Select: combined clauses
		{"select id from users where id = 42 group by id, name", 1, "(SELECT id FROM users WHERE (id = $1) GROUP BY id, name);"},
		{"select id from customers join addresses on id = customer_id where id = 46 group by id;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = $1) GROUP BY id);"},

		// Select: having clause
		{"select id from users group by id having id > 2;", 1, "(SELECT id FROM users GROUP BY id HAVING (id > $1));"},
		{"select id from users group by id having id > 2 and name = 'frodo';", 1, "(SELECT id FROM users GROUP BY id HAVING ((id > $1) AND (name = '$2')));"},

		// Select: limit
		{"select id from users limit 10;", 1, "(SELECT id FROM users LIMIT $1);"},
		{"select id from users limit ALL;", 1, "(SELECT id FROM users LIMIT ALL);"},

		// Select: offset
		{"select id from users limit ALL offset 10;", 1, "(SELECT id FROM users LIMIT ALL OFFSET $1);"},
		{"select id from users limit 10 offset 10;", 1, "(SELECT id FROM users LIMIT $1 OFFSET $2);"},
		{"select id from users limit 10 offset 1 ROW", 1, "(SELECT id FROM users LIMIT $1 OFFSET $2);"},
		{"select id from users limit 10 offset 2 ROWS;", 1, "(SELECT id FROM users LIMIT $1 OFFSET $2);"},

		// Select: combined order by, limit, offset
		{"select id from users order by id desc limit 10 offset 10;", 1, "(SELECT id FROM users ORDER BY id DESC LIMIT $1 OFFSET $2);"},
		{"select id from users order by id desc nulls last limit 10 offset 10;", 1, "(SELECT id FROM users ORDER BY id DESC NULLS LAST LIMIT $1 OFFSET $2);"},

		// Select: fetch
		{"select a from users order by a fetch first row only;", 1, "(SELECT a FROM users ORDER BY a FETCH NEXT $1 ROWS ONLY);"},
		{"select b from users order by b fetch first 3 rows only;", 1, "(SELECT b FROM users ORDER BY b FETCH NEXT $1 ROWS ONLY);"},
		{"select c from users order by c fetch first 10 rows with ties;", 1, "(SELECT c FROM users ORDER BY c FETCH NEXT $1 ROWS WITH TIES);"},

		// Select: IN clause
		{"select id from users where id IN ('7','8','9','14');", 1, "(SELECT id FROM users WHERE id IN ('$1', '$2', '$3', '$4'));"},
		{"select id from users where id IN ('17','21','34','48') AND name = 'brian';", 1, "(SELECT id FROM users WHERE (id IN ('$1', '$2', '$3', '$4') AND (name = '$5')));"},

		// Select: Cast literals
		{"select '100'::integer from a;", 1, "(SELECT '$1'::INTEGER FROM a);"},
		{"select 100::text from a;", 1, "(SELECT $1::TEXT FROM a);"},
		{"select a::text from b;", 1, "(SELECT a::TEXT FROM b);"},

		// Select: JSONB
		{"select id from users where data->'name' = 'brian';", 1, "(SELECT id FROM users WHERE ((data -> '$1') = '$2'));"},
		{"select id from users where data->>'name' = 'brian';", 1, "(SELECT id FROM users WHERE ((data ->> '$1') = '$2'));"},
		{"select id from users where data#>'{name}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #> '$1') = '$2'));"},
		{"select id from users where data#>>'{name}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #>> '$1') = '$2'));"},
		{"select id from users where data#>>'{name,first}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #>> '$1') = '$2'));"},
		{"select id from users where data#>>'{name,first}' = 'brian' and data#>>'{name,last}' = 'broderick';", 1, "(SELECT id FROM users WHERE (((data #>> '$1') = '$2') AND ((data #>> '$3') = '$4')));"},
		{"select * from users where metadata @> '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata @> '$1'));"},
		{"select * from users where metadata <@ '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata <@ '$1'));"},
		{"select * from users where metadata ? '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ? '$1'));"},
		{"select * from users where metadata ?| '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ?| '$1'));"},
		{"select * from users where metadata ?& '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ?& '$1'));"},
		{"select * from users where metadata || '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata || '$1'));"},
	}

	for _, tt := range tests {
		fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Equal(t, 1, len(program.Statements), "program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "select", stmt.TokenLiteral(), "program.Statements[0] is not ast.SelectStatement. got=%T", stmt)

		selectStmt, ok := stmt.(*ast.SelectStatement)
		assert.True(t, ok, "stmt is not *ast.SelectStatement. got=%T", stmt)

		selectExp, ok := selectStmt.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "stmt is not *ast.SelectExpression. got=%T", selectExp)

		assert.Equal(t, tt.tableCount, len(selectExp.Tables), "len(selectStmt.Tables) not %d. got=%d", tt.tableCount, len(selectExp.Tables))
		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "program.String() not '%s'. got=%s", tt.output, output)
		fmt.Printf("output: %s\n", output)
	}
}
