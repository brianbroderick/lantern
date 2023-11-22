package parser

import (
	"strings"
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
		{"select id from my_table where '2020-01-01' at time zone 'MDT' = '2023-01-01';", 1, "(SELECT id FROM my_table WHERE (('2020-01-01' AT TIME ZONE 'MDT') = '2023-01-01'));"},
		{"select * from tasks where date_trunc('day', created_at) = date_trunc('day', now()::timestamp with time zone at time zone 'America/Denver') LIMIT 1;", 1, "(SELECT * FROM tasks WHERE (date_trunc('day', created_at) = date_trunc('day', (now()::TIMESTAMP WITH TIME ZONE AT TIME ZONE 'America/Denver'))) LIMIT 1);"},
		{"select now()::timestamp with time zone from users;", 1, "(SELECT now()::TIMESTAMP WITH TIME ZONE FROM users);"},
		{"select '2020-01-01' at time zone 'MDT' from my_table;", 1, "(SELECT ('2020-01-01' AT TIME ZONE 'MDT') FROM my_table);"},
		{"select '2020-01-01' + 'MDT' from my_table;", 1, "(SELECT ('2020-01-01' + 'MDT') FROM my_table);"},
		{"select id from my_table where my_date at time zone my_zone > '2001-01-01';", 1, "(SELECT id FROM my_table WHERE ((my_date AT TIME ZONE my_zone) > '2001-01-01'));"},
		{"select id from users; select id from customers;", 2, "(SELECT id FROM users);(SELECT id FROM customers);"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, tt.statementCount, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, tt.statementCount, len(program.Statements))

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\n\noutput: %s\n\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
	}
}

func TestSingleSelectStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input      string
		tableCount int
		output     string
	}{
		// {"select now()::timestamp from users;", 1, "(SELECT now()::timestamp FROM users);"},

		// Select: simple
		{"select id from users;", 1, "(SELECT id FROM users);"},                                                                                                             // super basic select
		{"select u.* from users u;", 1, "(SELECT u.* FROM users u);"},                                                                                                       // check for a wildcard with a table alias
		{"select 2*3 from users;", 1, "(SELECT (2 * 3) FROM users);"},                                                                                                       // check that the asterisk is not treated as a wildcard
		{"select \"blah\".id from users", 1, "(SELECT blah.id FROM users);"},                                                                                                // check for double quotes around the table name
		{"select 1 * (2 + (6 / 4)) - 9 from users;", 1, "(SELECT ((1 * (2 + (6 / 4))) - 9) FROM users);"},                                                                   // math expression
		{"select id, name from users", 1, "(SELECT id, name FROM users);"},                                                                                                  // multiple columns
		{"select id, first_name from users;", 1, "(SELECT id, first_name FROM users);"},                                                                                     // underscore in a column name
		{"select id, first_name as name from users", 1, "(SELECT id, first_name AS name FROM users);"},                                                                      // column alias
		{"select id, first_name name from users", 1, "(SELECT id, first_name AS name FROM users);"},                                                                         // column alias
		{"select u.id, u.first_name as name from users u;", 1, "(SELECT u.id, u.first_name AS name FROM users u);"},                                                         // column alias with table alias
		{"select id from no_semi_colons", 1, "(SELECT id FROM no_semi_colons);"},                                                                                            // no semicolon
		{"select 1 + 2 as math, foo + 7 as seven from foo", 1, "(SELECT (1 + 2) AS math, (foo + 7) AS seven FROM foo);"},                                                    // multiple column aliases with expressions
		{"select 1 + 2 * 3 / value as math from foo", 1, "(SELECT (1 + ((2 * 3) / value)) AS math FROM foo);"},                                                              // more complex math expression
		{"select id from addresses a;", 1, "(SELECT id FROM addresses a);"},                                                                                                 // table alias
		{"select sum(a,b) from users;", 1, "(SELECT sum(a, b) FROM users);"},                                                                                                // function call
		{"select key, value from example where id = 20 AND key IN ( 'a', 'b', 'c' );", 1, "(SELECT key, value FROM example WHERE ((id = 20) AND key IN ('a', 'b', 'c')));"}, // removed the token KEY since it's not a PG reserved key word: https://www.postgresql.org/docs/13/sql-keywords-appendix.html
		{"SELECT translate(name, '''', '' ) as name FROM people WHERE id = 0;", 1, "(SELECT translate(name, '''', '') AS name FROM people WHERE (id = 0));"},                // escaped apostrophes

		// Select: distinct & all tokens
		{"select distinct id from users;", 1, "(SELECT DISTINCT id FROM users);"},
		{"select all id from users", 1, "(SELECT ALL id FROM users);"},
		{"select distinct on (location) reported_at, report from weather_reports;", 1, "(SELECT DISTINCT ON (location) reported_at, report FROM weather_reports);"},
		{"select c.id, string_agg ( distinct c.name, ', ' ) as value FROM companies c", 1, "(SELECT c.id, string_agg(DISTINCT c.name, ', ') AS value FROM companies c);"},
		{"select array_agg(distinct sub.id) from a", 1, "(SELECT array_agg(DISTINCT sub.id) FROM a);"},

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
		{"select id from addresses AS a JOIN states AS s ON (s.id = a.state_id AND s.code > 'ut')", 2, "(SELECT id FROM addresses a INNER JOIN states s ON ((s.id = a.state_id) AND (s.code > 'ut')));"},
		{"SELECT r.* FROM roles r, rights ri WHERE r.id = ri.role_id AND ri.deleted_by IS NULL AND ri.id = 12;", 2, "(SELECT r.* FROM roles r , rights ri WHERE (((r.id = ri.role_id) AND (ri.deleted_by IS NULL)) AND (ri.id = 12)));"},

		// Select: where clause
		{"select id from users where id = 42;", 1, "(SELECT id FROM users WHERE (id = 42));"},
		{"select id from users where id = 42 and customer_id = 74", 1, "(SELECT id FROM users WHERE ((id = 42) AND (customer_id = 74)));"},
		{"select id from users where id = 42 and customer_id > 74;", 1, "(SELECT id FROM users WHERE ((id = 42) AND (customer_id > 74)));"},
		{"select id from users where name = 'brian';", 1, "(SELECT id FROM users WHERE (name = 'brian'));"},
		{"select id from users where name = 'brian'", 1, "(SELECT id FROM users WHERE (name = 'brian'));"},
		{"select id from users where name is null", 1, "(SELECT id FROM users WHERE (name IS NULL));"},
		{"select id from users where name is not null", 1, "(SELECT id FROM users WHERE (name IS NOT NULL));"},

		// Select: IS comparisons
		{"select category from users where category is null", 1, "(SELECT category FROM users WHERE (category IS NULL));"},
		{"select category from users where category is not null", 1, "(SELECT category FROM users WHERE (category IS NOT NULL));"},
		{"select category from users where category is null and type = 1", 1, "(SELECT category FROM users WHERE ((category IS NULL) AND (type = 1)));"},
		{"select category from users where category is true", 1, "(SELECT category FROM users WHERE (category IS TRUE));"},
		{"select category from users where category is not true", 1, "(SELECT category FROM users WHERE (category IS NOT TRUE));"},
		{"select category from users where category is false", 1, "(SELECT category FROM users WHERE (category IS FALSE));"},
		{"select category from users where category is not false", 1, "(SELECT category FROM users WHERE (category IS NOT FALSE));"},
		{"select category from users where category is unknown", 1, "(SELECT category FROM users WHERE (category IS UNKNOWN));"},
		{"select category from users where category is not unknown", 1, "(SELECT category FROM users WHERE (category IS NOT UNKNOWN));"},
		{"select foo,bar from my_table where foo is distinct from bar;", 1, "(SELECT foo, bar FROM my_table WHERE (foo IS DISTINCT FROM bar));"},
		{"select foo,bar from my_table where foo is not distinct from bar;", 1, "(SELECT foo, bar FROM my_table WHERE (foo IS NOT DISTINCT FROM bar));"},

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
		{"select id from users order by id limit 34", 1, "(SELECT id FROM users ORDER BY id LIMIT 34);"},
		{"select * from users order by id offset 0 limit 34", 1, "(SELECT * FROM users ORDER BY id LIMIT 34 OFFSET 0);"},
		{"select * from users order by id limit 34 offset 0 ", 1, "(SELECT * FROM users ORDER BY id LIMIT 34 OFFSET 0);"},

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
		{"select id from modules where (option_id, external_id) IN ((1, 7))", 1, "(SELECT id FROM modules WHERE (option_id, external_id) IN ((1, 7)));"},                 // single tuple
		{"select id from modules where (option_id, external_id) IN ((1, 7), (2, 9))", 1, "(SELECT id FROM modules WHERE (option_id, external_id) IN ((1, 7), (2, 9)));"}, // multiple tuples
		{"select option_id, external_id from modules group by option_id, external_id having (option_id, external_id) IN ((1, 7), (2, 9))", 1, "(SELECT option_id, external_id FROM modules GROUP BY option_id, external_id HAVING (option_id, external_id) IN ((1, 7), (2, 9)));"},

		// Select: UNION clause
		{"select id from users union select id from customers;", 1, "(SELECT id FROM users UNION (SELECT id FROM customers));"},
		{"select id from users except select id from customers;", 1, "(SELECT id FROM users EXCEPT (SELECT id FROM customers));"},
		{"select id from users intersect select id from customers;", 1, "(SELECT id FROM users INTERSECT (SELECT id FROM customers));"},
		{"select id from users union all select id from customers;", 1, "(SELECT id FROM users UNION ALL (SELECT id FROM customers));"},
		{"select id from users except all select id from customers;", 1, "(SELECT id FROM users EXCEPT ALL (SELECT id FROM customers));"},
		{"select id from users intersect all select id from customers;", 1, "(SELECT id FROM users INTERSECT ALL (SELECT id FROM customers));"},

		// Select: Cast literals
		{"select '100'::integer from a;", 1, "(SELECT '100'::INTEGER FROM a);"},
		{"select 100::text from a;", 1, "(SELECT 100::TEXT FROM a);"},
		{"select a::text from b;", 1, "(SELECT a::TEXT FROM b);"},
		{"select load( array[ 1 ], array[ 2] ) from a", 1, "(SELECT load(array[1], array[2]) FROM a);"},
		{"select array[2]::integer from a", 1, "(SELECT array[2]::INTEGER FROM a);"},
		{"select load( array[ 1 ]::integer[], array[ 2]::integer[] ) from a", 1, "(SELECT load(array[1]::INTEGER[], array[2]::INTEGER[]) FROM a);"},
		{"select jsonb_array_length ( ( options ->> 'country_codes' ) :: jsonb ) from modules", 1, "(SELECT jsonb_array_length((options ->> 'country_codes')::JSONB) FROM modules);"},
		{"select now()::timestamp from users;", 1, "(SELECT now()::TIMESTAMP FROM users);"},
		{"select ( junk_drawer->>'ids' )::INT[] from dashboards", 1, "(SELECT (junk_drawer ->> 'ids')::INT[] FROM dashboards);"},

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

		// Select: CASE
		{"select case when id = 1 then 'one' when id = 2 then 'two' else 'other' end from users;", 1, "(SELECT CASE WHEN (id = 1) THEN 'one' WHEN (id = 2) THEN 'two' ELSE 'other' END FROM users);"},

		// Select: Functions
		{"select w() from a", 1, "(SELECT w() FROM a);"},
		{"select id from load(1,2)", 1, "(SELECT id FROM load(1, 2));"},
		{"select json_object_agg(foo order by id) from bar", 1, "(SELECT json_object_agg(foo ORDER BY id) FROM bar);"}, // aggregate function with order by

		// Subqueries
		{"select * from (select id from a) b order by id", 1, "(SELECT * FROM (SELECT id FROM a) b ORDER BY id);"},
		{"SELECT id FROM ( SELECT id FROM users u UNION SELECT id FROM users u ) as SubQ ;", 1, "(SELECT id FROM (SELECT id FROM users u UNION (SELECT id FROM users u)) SubQ);"}, // with union

		// Select: With Ordinality
		{"select * from unnest(array [ 4, 2, 1, 3, 7 ]) ;", 1, "(SELECT * FROM unnest(array[4, 2, 1, 3, 7]));"},
		{"select * from unnest(array [ 4, 2, 1, 3, 7 ]) with ordinality;", 1, "(SELECT * FROM unnest(array[4, 2, 1, 3, 7]) WITH ORDINALITY);"},
		{"select * from unnest(array [ 4, 2, 1, 3, 7 ]) with ordinality as t(key, index);", 1, "(SELECT * FROM unnest(array[4, 2, 1, 3, 7]) WITH ORDINALITY t(key, index));"},

		// Select: reserved words
		{"select id from users where any(type_ids) = 10;", 1, "(SELECT id FROM users WHERE (any(type_ids) = 10));"},               // any
		{"select null::integer AS id from users;", 1, "(SELECT NULL::INTEGER AS id FROM users);"},                                 // null
		{"select id from users where login_date < current_date;", 1, "(SELECT id FROM users WHERE (login_date < current_date));"}, // CURRENT_DATE
		{"select cast('100' as integer) from users", 1, "(SELECT CAST('100' AS integer) FROM users);"},                            // cast
		{"select id from account_triggers at", 1, "(SELECT id FROM account_triggers at);"},
		{"select u.id from users u join account_triggers at on at.user_id = u.id;", 2, "(SELECT u.id FROM users u INNER JOIN account_triggers at ON (at.user_id = u.id));"},

		// Less common expressions
		{"select current_date - INTERVAL '7 DAY' from users;", 1, "(SELECT (current_date - INTERVAL '7 DAY') FROM users);"},
		{"select count(*) as unfiltered from generate_series(1,10) as s(i)", 1, "(SELECT count(*) AS unfiltered FROM generate_series(1, 10) s(i));"},
		{"select COUNT(*) FILTER (WHERE i < 5) AS filtered from generate_series(1,10) s(i)", 1, "(SELECT (COUNT(*) FILTER WHERE((i < 5))) AS filtered FROM generate_series(1, 10) s(i));"},
		{"select trim(both 'x' from 'xTomxx') from users;", 1, "(SELECT trim(BOTH 'x' FROM 'xTomxx') FROM users);"},
		{"select trim(leading 'x' from 'xTomxx') from users;", 1, "(SELECT trim(LEADING 'x' FROM 'xTomxx') FROM users);"},
		{"select trim(trailing 'x' from 'xTomxx') from users;", 1, "(SELECT trim(TRAILING 'x' FROM 'xTomxx') FROM users);"},
		{"select substring('or' from 'Hello World!') from users;", 1, "(SELECT substring('or' FROM 'Hello World!') FROM users);"},
		{"select substring('Hello World!' from 2 for 4) from users;", 1, "(SELECT substring('Hello World!' FROM 2 FOR 4) FROM users);"},
		// {"select '2020-01-01' at time zone 'MDT' from my_table;", 1, "(SELECT ('2020-01-01' AT TIME ZONE 'MDT') FROM my_table);"},
		// {"select my_date at time zone my_zone from my_table;", 1, "(SELECT (my_date AT TIME ZONE my_zone) FROM my_table);"},

		// // Select with a CTE expression
		{"select count(1) from (with my_list as (select i from generate_series(1,10) s(i)) select i from my_list where i > 5) as t;", 1,
			"(SELECT count(1) FROM (WITH my_list AS (SELECT i FROM generate_series(1, 10) s(i)) (SELECT i FROM my_list WHERE (i > 5))) t);"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, 1, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "SELECT", strings.ToUpper(stmt.TokenLiteral()), "input: %s\nprogram.Statements[0] is not ast.SelectStatement. got=%T", tt.input, stmt)

		selectStmt, ok := stmt.(*ast.SelectStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectStatement. got=%T", tt.input, stmt)

		selectExp, ok := selectStmt.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectExpression. got=%T", tt.input, selectExp)

		assert.Equal(t, tt.tableCount, len(selectExp.Tables), "input: %s\nlen(selectStmt.Tables) not %d. got=%d", tt.input, tt.tableCount, len(selectExp.Tables))
		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s\nJSON:\n%s", tt.input, tt.output, output, program.Inspect(maskParams))
		// fmt.Printf("output: %s\n", output)
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
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, 1, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "select", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.SelectStatement. got=%T", tt.input, stmt)

		selectStmt, ok := stmt.(*ast.SelectStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectStatement. got=%T", tt.input, stmt)

		selectExp, ok := selectStmt.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectExpression. got=%T", tt.input, selectExp)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
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
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, 1, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "select", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.SelectStatement. got=%T", tt.input, stmt)

		selectStmt, ok := stmt.(*ast.SelectStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectStatement. got=%T", tt.input, stmt)

		selectExp, ok := selectStmt.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectExpression. got=%T", tt.input, selectExp)

		assert.Equal(t, tt.tableCount, len(selectExp.Tables), "input: %s\nlen(selectStmt.Tables) not %d. got=%d", tt.input, tt.tableCount, len(selectExp.Tables))
		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
