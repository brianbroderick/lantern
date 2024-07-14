package parser

import (
	"fmt"
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

		// left and right can be function names

		// {"SELECT rank() OVER (PARTITION BY id,command ORDER BY total_count desc) AS rank from queries;", 1,
		// 	"(SELECT (rank() OVER (PARTITION BY id, command ORDER BY total_count)) AS rank FROM queries);"},

		{"SELECT COUNT((f.id)) AS a_count, f.file_id as child_file_id FROM files f", 1, "(SELECT COUNT(f.id) AS a_count, f.file_id AS child_file_id FROM files f);"},
		{"select left('abc', 2); select right('abc', 2);", 2, "(SELECT left('abc', 2));(SELECT right('abc', 2));"},
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
		input  string
		output string
	}{
		// Select: simple
		{"select from users;", "(SELECT FROM users);"},             // no column (useful for exists functions)
		{"select id from users;", "(SELECT id FROM users);"},       // super basic select
		{"select u.* from users u;", "(SELECT u.* FROM users u);"}, // check for a wildcard with a table alias
		{"select 2*3 from users;", "(SELECT (2 * 3) FROM users);"},
		{`select 2 % 4 from users;`, "(SELECT (2 % 4) FROM users);"},                                                                                                     // check that the asterisk is not treated as a wildcard
		{"select +2 -3 from users;", "(SELECT ((+2) - 3) FROM users);"},                                                                                                  // PG allows + as a prefix operator
		{"select -2 +3 from users;", "(SELECT ((-2) + 3) FROM users);"},                                                                                                  // Negative numbers
		{`select "blah".id from blah`, `(SELECT "blah".id FROM blah);`},                                                                                                  // check for double quotes around the table name
		{"select 1 * (2 + (6 / 4)) - 9 from users;", "(SELECT ((1 * (2 + (6 / 4))) - 9) FROM users);"},                                                                   // math expression
		{"select id, name from users", "(SELECT id, name FROM users);"},                                                                                                  // multiple columns
		{"select id, first_name from users;", "(SELECT id, first_name FROM users);"},                                                                                     // underscore in a column name
		{"select id, first_name as name from users", "(SELECT id, first_name AS name FROM users);"},                                                                      // column alias
		{"select id, first_name name from users", "(SELECT id, first_name AS name FROM users);"},                                                                         // column alias
		{"select u.id, u.first_name as name from users u;", "(SELECT u.id, u.first_name AS name FROM users u);"},                                                         // column alias with table alias
		{"select id from no_semi_colons", "(SELECT id FROM no_semi_colons);"},                                                                                            // no semicolon
		{"select 1 + 2 as math, foo + 7 as seven from foo", "(SELECT (1 + 2) AS math, (foo + 7) AS seven FROM foo);"},                                                    // multiple column aliases with expressions
		{"select 1 + 2 * 3 / value as math from foo", "(SELECT (1 + ((2 * 3) / value)) AS math FROM foo);"},                                                              // more complex math expression
		{"select id from addresses a;", "(SELECT id FROM addresses a);"},                                                                                                 // table alias
		{"select sum(a,b) from users;", "(SELECT sum(a, b) FROM users);"},                                                                                                // function call
		{"select key, value from example where id = 20 AND key IN ( 'a', 'b', 'c' );", "(SELECT key, value FROM example WHERE ((id = 20) AND key IN ('a', 'b', 'c')));"}, // removed the token KEY since it's not a PG reserved key word: https://www.postgresql.org/docs/13/sql-keywords-appendix.html
		{"SELECT translate(name, '''', '' ) as name FROM people WHERE id = 0;", "(SELECT translate(name, '''', '') AS name FROM people WHERE (id = 0));"},                // escaped apostrophes
		{"select coalesce ( u.first_name || ' ' || u.last_name, u.first_name, u.last_name ) AS name from users u", "(SELECT coalesce(((u.first_name || ' ') || u.last_name), u.first_name, u.last_name) AS name FROM users u);"}, // coalesce

		// Select: distinct & all tokens
		{"select distinct id from users;", "(SELECT DISTINCT id FROM users);"},
		{"select all id from users", "(SELECT ALL id FROM users);"},
		{"select distinct on (location) reported_at, report from weather_reports;", "(SELECT DISTINCT ON (location) reported_at, report FROM weather_reports);"},
		{"select c.id, string_agg ( distinct c.name, ', ' ) as value FROM companies c", "(SELECT c.id, string_agg(DISTINCT c.name, ', ') AS value FROM companies c);"},
		{"select array_agg(distinct sub.id) from sub", "(SELECT array_agg(DISTINCT sub.id) FROM sub);"},

		// Select: window functions
		{"select avg(salary) over (partition by depname) from empsalary;", "(SELECT (avg(salary) OVER (PARTITION BY depname)) FROM empsalary);"},
		{"select avg(salary) over (order by depname) from empsalary", "(SELECT (avg(salary) OVER (ORDER BY depname)) FROM empsalary);"},
		{"select avg(salary) over (partition by salary order by depname) from empsalary;", "(SELECT (avg(salary) OVER (PARTITION BY salary ORDER BY depname)) FROM empsalary);"},
		{"select avg(salary) over (partition by salary order by depname desc) from empsalary", "(SELECT (avg(salary) OVER (PARTITION BY salary ORDER BY depname DESC)) FROM empsalary);"},
		{"select wf1() over w from table_name;", "(SELECT (wf1() OVER w) FROM table_name);"},
		{"select wf1() over w, wf2() over w from table_name;", "(SELECT (wf1() OVER w), (wf2() OVER w) FROM table_name);"},
		{"select wf1() over w, wf2() over w from table_name window w as (partition by c1 order by c2);", "(SELECT (wf1() OVER w), (wf2() OVER w) FROM table_name WINDOW w AS (PARTITION BY c1 ORDER BY c2));"},
		{"select wf1() over w, wf2() over w from table_name window w as (partition by c1 order by c2), foo as (partition by c3 order by c4);", "(SELECT (wf1() OVER w), (wf2() OVER w) FROM table_name WINDOW w AS (PARTITION BY c1 ORDER BY c2), foo AS (PARTITION BY c3 ORDER BY c4));"},

		// Select: joins
		{"select c.id from customers c join addresses a on c.id = a.customer_id;", "(SELECT c.id FROM customers c INNER JOIN addresses a ON (c.id = a.customer_id));"},
		{"select c.id from customers c join addresses a on (c.id = a.customer_id) join states s on (s.id = a.state_id);", "(SELECT c.id FROM customers c INNER JOIN addresses a ON (c.id = a.customer_id) INNER JOIN states s ON (s.id = a.state_id));"},
		// This is a complex join with multiple tables
		{"select c.id, c.name from customers c join addresses a on c.id = a.customer_id join states s on s.id = a.state_id join phone_numbers ph ON ph.customer_id = c.id;",
			"(SELECT c.id, c.name FROM customers c INNER JOIN addresses a ON (c.id = a.customer_id) INNER JOIN states s ON (s.id = a.state_id) INNER JOIN phone_numbers ph ON (ph.customer_id = c.id));"},
		{"select id from customers join addresses on id = customer_id;", "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id));"},
		{"select id from customers join addresses on id = customer_id join phones on id = phone_id;", "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) INNER JOIN phones ON (id = phone_id));"},
		{"select id from customers join addresses on customers.id = addresses.customer_id", "(SELECT id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id));"},
		{"select id from customers join addresses on id = customer_id where id = 46;", "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = 46));"},
		{"select id from customers left join addresses on id = customer_id;", "(SELECT id FROM customers LEFT JOIN addresses ON (id = customer_id));"},
		{"select id from customers left outer join addresses on id = customer_id;", "(SELECT id FROM customers LEFT JOIN addresses ON (id = customer_id));"},
		{"select id from addresses AS a JOIN states AS s ON (s.id = a.state_id AND s.code > 'ut')", "(SELECT id FROM addresses a INNER JOIN states s ON ((s.id = a.state_id) AND (s.code > 'ut')));"},
		{"SELECT r.* FROM roles r, rights ri WHERE r.id = ri.role_id AND ri.deleted_by IS NULL AND ri.id = 12;", "(SELECT r.* FROM roles r , rights ri WHERE (((r.id = ri.role_id) AND (ri.deleted_by IS NULL)) AND (ri.id = 12)));"},
		{"select left('abc', 2);", "(SELECT left('abc', 2));"},   // in this case, left is a string function getting the left 2 characters
		{"select right('abc', 2);", "(SELECT right('abc', 2));"}, // in this case, right is a string function getting the right 2 characters

		// Select: where clause
		{"select id from users where id = 42;", "(SELECT id FROM users WHERE (id = 42));"},
		{"select id from users where id = 42 and customer_id = 74", "(SELECT id FROM users WHERE ((id = 42) AND (customer_id = 74)));"},
		{"select id from users where id = 42 and customer_id > 74;", "(SELECT id FROM users WHERE ((id = 42) AND (customer_id > 74)));"},
		{"select id from users where name = 'brian';", "(SELECT id FROM users WHERE (name = 'brian'));"},
		{"select id from users where name = 'brian'", "(SELECT id FROM users WHERE (name = 'brian'));"},
		{"select id from users where name is null", "(SELECT id FROM users WHERE (name IS NULL));"},
		{"select id from users where name is not null", "(SELECT id FROM users WHERE (name IS NOT NULL));"},

		// Select: IS comparisons
		{"select category from users where category is null", "(SELECT category FROM users WHERE (category IS NULL));"},
		{"select category from users where category is not null", "(SELECT category FROM users WHERE (category IS NOT NULL));"},
		{"select category from users where category is null and type = 1", "(SELECT category FROM users WHERE ((category IS NULL) AND (type = 1)));"},
		{"select category from users where category is true", "(SELECT category FROM users WHERE (category IS TRUE));"},
		{"select category from users where category is not true", "(SELECT category FROM users WHERE (category IS NOT TRUE));"},
		{"select category from users where category is false", "(SELECT category FROM users WHERE (category IS FALSE));"},
		{"select category from users where category is not false", "(SELECT category FROM users WHERE (category IS NOT FALSE));"},
		{"select category from users where category is unknown", "(SELECT category FROM users WHERE (category IS UNKNOWN));"},
		{"select category from users where category is not unknown", "(SELECT category FROM users WHERE (category IS NOT UNKNOWN));"},
		{"select foo,bar from my_table where foo is distinct from bar;", "(SELECT foo, bar FROM my_table WHERE (foo IS DISTINCT FROM bar));"},
		{"select foo,bar from my_table where foo is not distinct from bar;", "(SELECT foo, bar FROM my_table WHERE (foo IS NOT DISTINCT FROM bar));"},

		// Select: group by
		{"select baz from sales group by bar;", "(SELECT baz FROM sales GROUP BY bar);"},
		{"select id from users group by id, name;", "(SELECT id FROM users GROUP BY id, name);"},

		// Select: combined clauses
		{"select id from users where id = 42 group by id, name", "(SELECT id FROM users WHERE (id = 42) GROUP BY id, name);"},
		{"select id from customers join addresses on id = customer_id where id = 46 group by id;", "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = 46) GROUP BY id);"},

		// Select: having clause
		{"select id from users group by id having id > 2;", "(SELECT id FROM users GROUP BY id HAVING (id > 2));"},
		{"select id from users group by id having id > 2 and name = 'frodo';", "(SELECT id FROM users GROUP BY id HAVING ((id > 2) AND (name = 'frodo')));"},

		// Select: order by
		{"select id from users order by id;", "(SELECT id FROM users ORDER BY id);"},
		{"select id from users order by id desc, name", "(SELECT id FROM users ORDER BY id DESC, name);"},
		{"select id from users order by id desc nulls first, name nulls last;", "(SELECT id FROM users ORDER BY id DESC NULLS FIRST, name NULLS LAST);"},

		// Select: limit
		{"select id from users limit 10;", "(SELECT id FROM users LIMIT 10);"},
		{"select id from users limit ALL;", "(SELECT id FROM users LIMIT ALL);"},
		{"select id from users limit ALL", "(SELECT id FROM users LIMIT ALL);"},
		{"select id from users order by id limit 34", "(SELECT id FROM users ORDER BY id LIMIT 34);"},
		{"select * from users order by id offset 0 limit 34", "(SELECT * FROM users ORDER BY id LIMIT 34 OFFSET 0);"},
		{"select * from users order by id limit 34 offset 0 ", "(SELECT * FROM users ORDER BY id LIMIT 34 OFFSET 0);"},

		// Select: offset
		{"select id from users limit ALL offset 10;", "(SELECT id FROM users LIMIT ALL OFFSET 10);"},
		{"select id from users limit 10 offset 10;", "(SELECT id FROM users LIMIT 10 OFFSET 10);"},
		{"select id from users limit 10 offset 1 ROW", "(SELECT id FROM users LIMIT 10 OFFSET 1);"},
		{"select id from users limit 10 offset 2 ROWS;", "(SELECT id FROM users LIMIT 10 OFFSET 2);"},

		// Select: combined order by, limit, offset
		{"select id from users order by id desc limit 10 offset 10;", "(SELECT id FROM users ORDER BY id DESC LIMIT 10 OFFSET 10);"},
		{"select id from users order by id desc nulls last limit 10 offset 10;", "(SELECT id FROM users ORDER BY id DESC NULLS LAST LIMIT 10 OFFSET 10);"},

		// Select: fetch
		{"select id from users order by id fetch first row only;", "(SELECT id FROM users ORDER BY id FETCH NEXT 1 ROWS ONLY);"},
		{"select id from users order by id fetch first 3 rows only;", "(SELECT id FROM users ORDER BY id FETCH NEXT 3 ROWS ONLY);"},
		{"select id from users order by id fetch first 10 rows with ties;", "(SELECT id FROM users ORDER BY id FETCH NEXT 10 ROWS WITH TIES);"},

		// Select: for update
		{"select id from users for update;", "(SELECT id FROM users FOR UPDATE);"},
		{"select id from users for no key update;;", "(SELECT id FROM users FOR NO KEY UPDATE);"},
		{"select id from users for share;", "(SELECT id FROM users FOR SHARE);"},
		{"select id from users for key share", "(SELECT id FROM users FOR KEY SHARE);"},
		{"select id from users for update of users;", "(SELECT id FROM users FOR UPDATE OF users);"},
		{"select id from users for update of users, addresses;", "(SELECT id FROM users FOR UPDATE OF users, addresses);"},
		{"select id from users for update of users, addresses nowait;", "(SELECT id FROM users FOR UPDATE OF users, addresses NOWAIT);"},
		{"select id from users for update of users, addresses skip locked;", "(SELECT id FROM users FOR UPDATE OF users, addresses SKIP LOCKED);"},

		// Select: IN operator
		{"select id from users where id in ('1','2','3','4');", "(SELECT id FROM users WHERE id IN ('1', '2', '3', '4'));"},
		{"select id from users where id not in ('1','2','3','4');", "(SELECT id FROM users WHERE id NOT IN ('1', '2', '3', '4'));"},
		{"select id from users where id IN ('1','2','3','4') AND name = 'brian';", "(SELECT id FROM users WHERE (id IN ('1', '2', '3', '4') AND (name = 'brian')));"},
		{"select id from users where id IN (1,2,3,4);", "(SELECT id FROM users WHERE id IN (1, 2, 3, 4));"},
		{"select id from modules where (option_id, external_id) IN ((1, 7))", "(SELECT id FROM modules WHERE (option_id, external_id) IN ((1, 7)));"},                 // single tuple
		{"select id from modules where (option_id, external_id) IN ((1, 7), (2, 9))", "(SELECT id FROM modules WHERE (option_id, external_id) IN ((1, 7), (2, 9)));"}, // multiple tuples
		{"select option_id, external_id from modules group by option_id, external_id having (option_id, external_id) IN ((1, 7), (2, 9))", "(SELECT option_id, external_id FROM modules GROUP BY option_id, external_id HAVING (option_id, external_id) IN ((1, 7), (2, 9)));"},
		{"select position('b' IN 'brian') as foo from cars", "(SELECT position('b' IN ('brian')) AS foo FROM cars);"}, // in inside a function call
		{"select count(1) from cars c left join models m ON ( position((c.key) in m.formula)<>0 )", "(SELECT count(1) FROM cars c LEFT JOIN models m ON (position(c.key IN (m.formula)) <> 0));"},
		{"SELECT u.* from users u where u.id IN (42);", "(SELECT u.* FROM users u WHERE u.id IN (42));"},

		// Select: LIKE operator
		{"select id from users where name like 'brian';", "(SELECT id FROM users WHERE (name LIKE 'brian'));"},                                        // basic like
		{"select id from users where name not like 'brian';", "(SELECT id FROM users WHERE (name NOT LIKE 'brian'));"},                                // basic not like
		{"select id from users where rownum between 1 and sample_size", "(SELECT id FROM users WHERE (rownum BETWEEN (1 AND sample_size)));"},         // BETWEEN
		{"select id from users where rownum not between 1 and sample_size", "(SELECT id FROM users WHERE (rownum NOT BETWEEN (1 AND sample_size)));"}, // BETWEEN
		{"select if from users where rownum between 1 and sample_size group by property_id;", "(SELECT if FROM users WHERE (rownum BETWEEN (1 AND sample_size)) GROUP BY property_id);"},
		{"select * from mytable where mycolumn ~ 'regexp';", "(SELECT * FROM mytable WHERE (mycolumn ~ 'regexp'));"},     // basic regex (case sensitive)
		{"select * from mytable where mycolumn ~* 'regexp';", "(SELECT * FROM mytable WHERE (mycolumn ~* 'regexp'));"},   // basic regex (case insensitive)
		{"select * from mytable where mycolumn !~ 'regexp';", "(SELECT * FROM mytable WHERE (mycolumn !~ 'regexp'));"},   // basic not regex (case sensitive)
		{"select * from mytable where mycolumn !~* 'regexp';", "(SELECT * FROM mytable WHERE (mycolumn !~* 'regexp'));"}, // basic not regex (case insensitive)
		// {"select select 'abc' similar to 'abc' from users;", ""}, // TODO: handle similar to
		// {"select select 'abc' not similar to 'abc' from users;", ""}, // TODO: handle similar to

		// Select: EXISTS operator. In this case, NOT is a prefix operator
		{"select id from users where exists (select id from addresses where user_id = users.id);", "(SELECT id FROM users WHERE exists((SELECT id FROM addresses WHERE (user_id = users.id))));"},
		{"select id from users where not exists (select id from addresses where user_id = users.id);",
			"(SELECT id FROM users WHERE (NOT exists((SELECT id FROM addresses WHERE (user_id = users.id)))));"},

		// Select: UNION clause
		//   No tables
		{"SELECT '08/22/2023'::DATE;", "(SELECT '08/22/2023'::DATE);"},
		{"SELECT '123' union select '456';", "((SELECT '123') UNION (SELECT '456'));"},
		{"SELECT '08/22/2023'::DATE union select '08/23/2023'::DATE;", "((SELECT '08/22/2023'::DATE) UNION (SELECT '08/23/2023'::DATE));"},
		{"SELECT '123' union select '456';", "((SELECT '123') UNION (SELECT '456'));"},

		//   Single tables
		{"select id from users union select id from customers;", "((SELECT id FROM users) UNION (SELECT id FROM customers));"},
		{"select id from users except select id from customers;", "((SELECT id FROM users) EXCEPT (SELECT id FROM customers));"},
		{"select id from users intersect select id from customers;", "((SELECT id FROM users) INTERSECT (SELECT id FROM customers));"},
		{"select id from users union all select id from customers;", "((SELECT id FROM users) UNION ALL (SELECT id FROM customers));"},
		{"select id from users except all select id from customers;", "((SELECT id FROM users) EXCEPT ALL (SELECT id FROM customers));"},
		{"select id from users intersect all select id from customers;", "((SELECT id FROM users) INTERSECT ALL (SELECT id FROM customers));"},
		{"select name from users union select fname from people", "((SELECT name FROM users) UNION (SELECT fname FROM people));"},

		// Select: Cast literals
		{"select '100'::integer from a;", "(SELECT '100'::INTEGER FROM a);"},
		{"select 100::text from a;", "(SELECT 100::TEXT FROM a);"},
		{"select a::text from b;", "(SELECT a::TEXT FROM b);"},
		{"select load( array[ 1 ], array[ 2] ) from a", "(SELECT load(array[1], array[2]) FROM a);"},
		{"select array[2]::integer from a", "(SELECT array[2]::INTEGER FROM a);"},
		{"select load( array[ 1 ]::integer[], array[ 2]::integer[] ) from a", "(SELECT load(array[1]::INTEGER[], array[2]::INTEGER[]) FROM a);"},
		{"select jsonb_array_length ( ( options ->> 'country_codes' ) :: jsonb ) from modules", "(SELECT jsonb_array_length((options ->> 'country_codes')::JSONB) FROM modules);"},
		{"select now()::timestamp from users;", "(SELECT now()::TIMESTAMP FROM users);"},
		{"select ( junk_drawer->>'ids' )::INT[] from dashboards", "(SELECT (junk_drawer ->> 'ids')::INT[] FROM dashboards);"},
		{"select end_date + '1 day' ::INTERVAL from some_dates;", "(SELECT (end_date + '1 day'::INTERVAL) FROM some_dates);"},
		{`select CAST(u.depth AS DECIMAL(18, 2)) from users`, "(SELECT CAST(u.depth AS DECIMAL(18, 2)) FROM users);"},

		// Select: JSONB
		{"select id from users where data->'name' = 'brian';", "(SELECT id FROM users WHERE ((data -> 'name') = 'brian'));"},
		{"select id from users where data->>'name' = 'brian';", "(SELECT id FROM users WHERE ((data ->> 'name') = 'brian'));"},
		{"select id from users where data#>'{name}' = 'brian';", "(SELECT id FROM users WHERE ((data #> '{name}') = 'brian'));"},
		{"select id from users where data#>>'{name}' = 'brian';", "(SELECT id FROM users WHERE ((data #>> '{name}') = 'brian'));"},
		{"select id from users where data#>>'{name,first}' = 'brian';", "(SELECT id FROM users WHERE ((data #>> '{name,first}') = 'brian'));"},
		{"select id from users where data#>>'{name,first}' = 'brian' and data#>>'{name,last}' = 'broderick';", "(SELECT id FROM users WHERE (((data #>> '{name,first}') = 'brian') AND ((data #>> '{name,last}') = 'broderick')));"},
		{"select * from users where metadata @> '{\"age\": 42}';", "(SELECT * FROM users WHERE (metadata @> '{\"age\": 42}'));"},
		{"select * from users where metadata <@ '{\"age\": 42}';", "(SELECT * FROM users WHERE (metadata <@ '{\"age\": 42}'));"},
		{"select * from users where metadata ? '{\"age\": 42}';", "(SELECT * FROM users WHERE (metadata ? '{\"age\": 42}'));"},
		{"select * from users where metadata ?| '{\"age\": 42}';", "(SELECT * FROM users WHERE (metadata ?| '{\"age\": 42}'));"},
		{"select * from users where metadata ?& '{\"age\": 42}';", "(SELECT * FROM users WHERE (metadata ?& '{\"age\": 42}'));"},
		{"select * from users where metadata || '{\"age\": 42}';", "(SELECT * FROM users WHERE (metadata || '{\"age\": 42}'));"},

		// Select: ARRAYs
		{"select array['John'] from users;", "(SELECT array['John'] FROM users);"},
		{"select array['John', 'Joseph'] from users;", "(SELECT array['John', 'Joseph'] FROM users);"},
		{"select array['John', 'Joseph', 'Anna', 'Henry'] && array['Henry', 'John'] from users;", "(SELECT (array['John', 'Joseph', 'Anna', 'Henry'] && array['Henry', 'John']) FROM users);"},
		{"select pay[3] FROM emp;", "(SELECT pay[3] FROM emp);"},
		{"select pay[1:2] FROM emp;", "(SELECT pay[(1 : 2)] FROM emp);"},
		{"select pay[1:] FROM emp;", "(SELECT pay[(1 : )] FROM emp);"},
		{"select pay[:1] FROM emp;", "(SELECT pay[( : 1)] FROM emp);"},

		// Select: CASE
		{"select case when id = 1 then 'one' when id = 2 then 'two' else 'other' end from users;", "(SELECT CASE WHEN (id = 1) THEN 'one' WHEN (id = 2) THEN 'two' ELSE 'other' END FROM users);"},
		{`SELECT case when (type_id = 3) then 1 else 0 end::text as is_complete FROM users`, "(SELECT CASE WHEN (type_id = 3) THEN 1 ELSE 0 END::TEXT AS is_complete FROM users);"},
		{"select array_agg(name order by name) as names from users", "(SELECT array_agg(name ORDER BY name) AS names FROM users);"},
		{"SELECT case when (id = 3) then 1 else 0 end AS names from users", "(SELECT CASE WHEN (id = 3) THEN 1 ELSE 0 END AS names FROM users);"},
		{"SELECT (case when (id = 3) then 1 else 0 end) AS names from users", "(SELECT CASE WHEN (id = 3) THEN 1 ELSE 0 END AS names FROM users);"},
		{"SELECT array_agg(case when id = 3 then 1 else 0 end order by name, id) AS names from users",
			"(SELECT array_agg(CASE WHEN (id = 3) THEN 1 ELSE 0 END ORDER BY name, id) AS names FROM users);"},
		{"SELECT array_agg(case when (uid = 3) then 1 else 0 end order by name) AS names from users",
			"(SELECT array_agg(CASE WHEN (uid = 3) THEN 1 ELSE 0 END ORDER BY name) AS names FROM users);"},

		// Select: Functions
		{"select w() from a", "(SELECT w() FROM a);"},
		{"select id from load(1,2)", "(SELECT id FROM load(1, 2));"},
		{"select json_object_agg(foo order by id) from bar", "(SELECT json_object_agg(foo ORDER BY id) FROM bar);"}, // aggregate function with order by
		{"select array_agg( distinct sub.name_first order by sub.name_first ) from customers sub", "(SELECT array_agg(DISTINCT sub.name_first ORDER BY sub.name_first) FROM customers sub);"}, // aggregate with distinct

		// Subqueries
		{"select * from (select id from a) b order by id", "(SELECT * FROM (SELECT id FROM a) b ORDER BY id);"},
		{"select id from (select id from users union select id from people) u;", "(SELECT id FROM ((SELECT id FROM users) UNION (SELECT id FROM people)) u);"},
		{"SELECT id FROM ( SELECT id FROM users u UNION SELECT id FROM users u ) as SubQ ;", "(SELECT id FROM ((SELECT id FROM users u) UNION (SELECT id FROM users u)) SubQ);"}, // with union

		// Select: With Ordinality
		{"select * from unnest(array [ 4, 2, 1, 3, 7 ]) ;", "(SELECT * FROM unnest(array[4, 2, 1, 3, 7]));"},
		{"select * from unnest(array [ 4, 2, 1, 3, 7 ]) with ordinality;", "(SELECT * FROM unnest(array[4, 2, 1, 3, 7]) WITH ORDINALITY);"},
		{"select * from unnest(array [ 4, 2, 1, 3, 7 ]) with ordinality as t(key, index);", "(SELECT * FROM unnest(array[4, 2, 1, 3, 7]) WITH ORDINALITY t(key, index));"},
		{"select * from unnest(array [ 4, 2, 1, 3, 7 ]) with ordinality t(key, index);", "(SELECT * FROM unnest(array[4, 2, 1, 3, 7]) WITH ORDINALITY t(key, index));"},

		// Select: reserved words
		{"select id from users where any(type_ids) = 10;", "(SELECT id FROM users WHERE (any(type_ids) = 10));"},               // any
		{"select null::integer AS id from users;", "(SELECT NULL::INTEGER AS id FROM users);"},                                 // null
		{"select id from users where login_date < current_date;", "(SELECT id FROM users WHERE (login_date < current_date));"}, // CURRENT_DATE
		{"select cast('100' as integer) from users", "(SELECT CAST('100' AS integer) FROM users);"},                            // cast
		{"select id from account_triggers at", "(SELECT id FROM account_triggers at);"},
		{"select u.id from users u join account_triggers at on at.user_id = u.id;", "(SELECT u.id FROM users u INNER JOIN account_triggers at ON (at.user_id = u.id));"},
		{"select v as values from users;", "(SELECT v AS values FROM users);"},
		{`select e.details->>'values' as values	from events;`, "(SELECT (e.details ->> 'values') AS values FROM events);"},
		{`select set.* from server_event_types as set;`, "(SELECT set.* FROM server_event_types set);"},
		{`select e.* from events e join server_event_types as set on (set.id = e.server_event_type_id);`, "(SELECT e.* FROM events e INNER JOIN server_event_types set ON (set.id = e.server_event_type_id));"},
		{`select fname || lname as user from users;`, "(SELECT (fname || lname) AS user FROM users);"},
		{`select 1 as order from users;`, "(SELECT 1 AS order FROM users);"},

		// Less common expressions
		{"select count(*) as unfiltered from generate_series(1,10) as s(i)", "(SELECT count(*) AS unfiltered FROM generate_series(1, 10) s(i));"},
		{"select COUNT(*) FILTER (WHERE i < 5) AS filtered from generate_series(1,10) s(i)", "(SELECT (COUNT(*) FILTER WHERE((i < 5))) AS filtered FROM generate_series(1, 10) s(i));"},
		{"select trim(both 'x' from 'xTomxx') from users;", "(SELECT trim(BOTH 'x' FROM 'xTomxx') FROM users);"},
		{"select trim(leading 'x' from 'xTomxx') from users;", "(SELECT trim(LEADING 'x' FROM 'xTomxx') FROM users);"},
		{"select trim(trailing 'x' from 'xTomxx') from users;", "(SELECT trim(TRAILING 'x' FROM 'xTomxx') FROM users);"},
		{"select substring('or' from 'Hello World!') from users;", "(SELECT substring('or' FROM 'Hello World!') FROM users);"},
		{"select substring('Hello World!' from 2 for 4) from users;", "(SELECT substring('Hello World!' FROM 2 FOR 4) FROM users);"},
		{"select id from extra where number ilike e'001';", "(SELECT id FROM extra WHERE (number ILIKE E'001'));"}, // escapestring of the form e'...'
		{"select id from extra where number ilike E'%6%';", "(SELECT id FROM extra WHERE (number ILIKE E'%6%'));"}, // escapestring of the form E'...'
		{"SELECT * FROM ( VALUES (41, 1), (42, 2), (43, 3), (44, 4), (45, 5), (46, 6) ) AS t ( id, type_id )", "(SELECT * FROM (VALUES (41, 1), (42, 2), (43, 3), (44, 4), (45, 5), (46, 6)) t(id, type_id));"},

		// Intervals
		{"select current_date - INTERVAL '7 DAY' from users;", "(SELECT (current_date - INTERVAL '7 DAY') FROM users);"},
		{"select 1 from users where a = interval '1' year", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' YEAR));"},
		{"select 1 from users where a = interval '1' month", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' MONTH));"},
		{"select 1 from users where a = interval '1' day", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' DAY));"},
		{"select 1 from users where a = interval '1' hour", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' HOUR));"},
		{"select 1 from users where a = interval '1' minute", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' MINUTE));"},
		{"select 1 from users where a = interval '1' SECOND", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' SECOND));"},
		{"select 1 from users where a = interval '1' year to month", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' YEAR TO MONTH));"},
		{"select 1 from users where a = interval '1' day to hour", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' DAY TO HOUR));"},
		{"select 1 from users where a = interval '1' day to minute", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' DAY TO MINUTE));"},
		{"select 1 from users where a = interval '1' day to second", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' DAY TO SECOND));"},
		{"select 1 from users where a = interval '1' hour to minute", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' HOUR TO MINUTE));"},
		{"select 1 from users where a = interval '1' hour to second", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' HOUR TO SECOND));"},
		{"select 1 from users where a = interval '1' minute to second", "(SELECT 1 FROM users WHERE (a = INTERVAL '1' MINUTE TO SECOND));"},

		// Multi word keywords: AT TIME ZONE, and TIMESTAMP WITH TIME ZONE
		{"select id from my_table where '2020-01-01' at time zone 'MDT' = '2023-01-01';", "(SELECT id FROM my_table WHERE (('2020-01-01' AT TIME ZONE 'MDT') = '2023-01-01'));"},
		{"select * from tasks where date_trunc('day', created_at) = date_trunc('day', now()::timestamp with time zone at time zone 'America/Denver') LIMIT 1;", "(SELECT * FROM tasks WHERE (date_trunc('day', created_at) = date_trunc('day', (now()::TIMESTAMP WITH TIME ZONE AT TIME ZONE 'America/Denver'))) LIMIT 1);"},
		{"select now()::timestamp with time zone from users;", "(SELECT now()::TIMESTAMP WITH TIME ZONE FROM users);"},
		{"select '2020-01-01' at time zone 'MDT' from my_table;", "(SELECT ('2020-01-01' AT TIME ZONE 'MDT') FROM my_table);"},
		{"select '2020-01-01' + 'MDT' from my_table;", "(SELECT ('2020-01-01' + 'MDT') FROM my_table);"},
		{"select id from my_table where my_date at time zone my_zone > '2001-01-01';", "(SELECT id FROM my_table WHERE ((my_date AT TIME ZONE my_zone) > '2001-01-01'));"},
		{`select timestamp '2001-09-23';`, "(SELECT TIMESTAMP '2001-09-23');"},
		{`select timestamp '2001-09-23' + interval '72 hours';`, "(SELECT (TIMESTAMP '2001-09-23' + INTERVAL '72 hours'));"},

		// // Select with a CTE expression
		{"select count(1) from (with my_list as (select i from generate_series(1,10) s(i)) select i from my_list where i > 5) as t;",
			"(SELECT count(1) FROM (WITH my_list AS (SELECT i FROM generate_series(1, 10) s(i)) (SELECT i FROM my_list WHERE (i > 5))) t);"},

		// Dots
		{`select id as "my.id" from users u`, `(SELECT id AS "my.id" FROM users u);`},
		{`select id as "my" from users u`, `(SELECT id AS "my" FROM users u);`},
		{"select u . id from users u", "(SELECT u.id FROM users u);"}, // "no prefix parse function for DOT found at line 0 char 25"

		// Comments
		{`select *
			-- id
			-- name 
				from users u
				JOIN addresses a ON (u.id = a.user_id)
				where id = 42;`, "(SELECT * FROM users u INNER JOIN addresses a ON (u.id = a.user_id) WHERE (id = 42));"}, // multiple single line comments

		{`select id from users where id = $1`, "(SELECT id FROM users WHERE (id = $1));"}, // $1 is a parameter
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

		reflectType := fmt.Sprintf("%T", selectStmt.Expressions[0])
		assert.True(t, reflectType == "*ast.SelectExpression" || reflectType == "*ast.UnionExpression",
			"input: %s\nstmt is the wrong type. got=%T", tt.input, selectStmt.Expressions[0])

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
		assert.Equal(t, "SELECT", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.SelectStatement. got=%T", tt.input, stmt)

		selectStmt, ok := stmt.(*ast.SelectStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectStatement. got=%T", tt.input, stmt)

		selectExp, ok := selectStmt.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectExpression. got=%T", tt.input, selectExp)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}

// TestMaskParams tests that the maskParams flag works as expected.
// The purpose of masking parameters is to make it easier to total up the number of the same query.
// It is not intended to be used to convert into a prepared statement.
// Lantern returns a ? instead of how Postgres currently returns $1, $2, etc. because I'm squashing lists down to a single parameter.
// The problem is when query contains a list along with addtional parameters, the additional parameter's number would be different depending on
// the number of items in the list.
// For example: select id from users where id in (1,2,3,4) and name = 'brian'; would be (SELECT id FROM users WHERE (id IN ($1)) AND (name = '$5'));
// The $5 is a problem because it would be different if there is a different count in the IN clause.

func TestMaskParams(t *testing.T) {
	maskParams := true

	tests := []struct {
		input      string
		tableCount int
		output     string
	}{
		// Some Selects
		{"select id from users;", 1, "(SELECT id FROM users);"},
		{"select 1 * (2 + (6 / 4)) - 9 from users;", 1, "(SELECT ((? * (? + (? / ?))) - ?) FROM users);"},
		{"select id from customers join addresses on id = customer_id where id = 46;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = ?));"},

		// Select: where clause
		{"select id from users where id = 42;", 1, "(SELECT id FROM users WHERE (id = ?));"},
		{"select id from users where id = 42 and customer_id = 74", 1, "(SELECT id FROM users WHERE ((id = ?) AND (customer_id = ?)));"},
		{"select id from users where id = 42 and customer_id > 74;", 1, "(SELECT id FROM users WHERE ((id = ?) AND (customer_id > ?)));"},
		{"select id from users where name = 'brian';", 1, "(SELECT id FROM users WHERE (name = '?'));"},
		{"select id from users where name = 'brian'", 1, "(SELECT id FROM users WHERE (name = '?'));"},
		{"select id from users where name != 'brian'", 1, "(SELECT id FROM users WHERE (name != '?'));"},
		{"select id from users where name like 'brian%'", 1, "(SELECT id FROM users WHERE (name LIKE '?'));"},
		{"select id from users where name not like 'brian%'", 1, "(SELECT id FROM users WHERE (name NOT LIKE '?'));"},

		// Select: combined clauses
		{"select id from users where id = 42 group by id, name", 1, "(SELECT id FROM users WHERE (id = ?) GROUP BY id, name);"},
		{"select id from customers join addresses on id = customer_id where id = 46 group by id;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = ?) GROUP BY id);"},

		// Select: having clause
		{"select id from users group by id having id > 2;", 1, "(SELECT id FROM users GROUP BY id HAVING (id > ?));"},
		{"select id from users group by id having id > 2 and name = 'frodo';", 1, "(SELECT id FROM users GROUP BY id HAVING ((id > ?) AND (name = '?')));"},

		// Select: limit
		{"select id from users limit 10;", 1, "(SELECT id FROM users LIMIT ?);"},
		{"select id from users limit ALL;", 1, "(SELECT id FROM users LIMIT ALL);"},

		// Select: offset
		{"select id from users limit ALL offset 10;", 1, "(SELECT id FROM users LIMIT ALL OFFSET ?);"},
		{"select id from users limit 10 offset 10;", 1, "(SELECT id FROM users LIMIT ? OFFSET ?);"},
		{"select id from users limit 10 offset 1 ROW", 1, "(SELECT id FROM users LIMIT ? OFFSET ?);"},
		{"select id from users limit 10 offset 2 ROWS;", 1, "(SELECT id FROM users LIMIT ? OFFSET ?);"},

		// Select: combined order by, limit, offset
		{"select id from users order by id desc limit 10 offset 10;", 1, "(SELECT id FROM users ORDER BY id DESC LIMIT ? OFFSET ?);"},
		{"select id from users order by id desc nulls last limit 10 offset 10;", 1, "(SELECT id FROM users ORDER BY id DESC NULLS LAST LIMIT ? OFFSET ?);"},

		// Select: fetch
		{"select a from users order by a fetch first row only;", 1, "(SELECT a FROM users ORDER BY a FETCH NEXT ? ROWS ONLY);"},
		{"select b from users order by b fetch first 3 rows only;", 1, "(SELECT b FROM users ORDER BY b FETCH NEXT ? ROWS ONLY);"},
		{"select c from users order by c fetch first 10 rows with ties;", 1, "(SELECT c FROM users ORDER BY c FETCH NEXT ? ROWS WITH TIES);"},

		// Select: IN clause
		{"select id from users where id IN ('7','8','9','14');", 1, "(SELECT id FROM users WHERE id IN ('?'));"},
		{"select id from users where id IN ('17','21','34','48') AND name = 'brian';", 1, "(SELECT id FROM users WHERE (id IN ('?') AND (name = '?')));"},

		// Select: IS
		// Note: both the IS expression and the boolean, null, etc. advance the parameter index.
		// This will cause the parameter index to be off by one for the second parameter.
		// But this is ok because it squashes both positive and negative values into a single parameter.
		{"select id from users where id IS NULL;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS NOT NULL;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS TRUE;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS NOT TRUE;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS FALSE;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS NOT FALSE;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS UNKNOWN;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS NOT UNKNOWN;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS distinct from xid;", 1, "(SELECT id FROM users WHERE (id IS ?));"},
		{"select id from users where id IS not distinct from xid;", 1, "(SELECT id FROM users WHERE (id IS ?));"},

		// Select: Cast literals
		{"select '100'::integer from a;", 1, "(SELECT '?'::INTEGER FROM a);"},
		{"select 100::text from a;", 1, "(SELECT ?::TEXT FROM a);"},
		{"select a::text from b;", 1, "(SELECT a::TEXT FROM b);"},

		// Select: JSONB
		{"select id from users where data->'name' = 'brian';", 1, "(SELECT id FROM users WHERE ((data -> '?') = '?'));"},
		{"select id from users where data->>'name' = 'brian';", 1, "(SELECT id FROM users WHERE ((data ->> '?') = '?'));"},
		{"select id from users where data#>'{name}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #> '?') = '?'));"},
		{"select id from users where data#>>'{name}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #>> '?') = '?'));"},
		{"select id from users where data#>>'{name,first}' = 'brian';", 1, "(SELECT id FROM users WHERE ((data #>> '?') = '?'));"},
		{"select id from users where data#>>'{name,first}' = 'brian' and data#>>'{name,last}' = 'broderick';", 1, "(SELECT id FROM users WHERE (((data #>> '?') = '?') AND ((data #>> '?') = '?')));"},
		{"select * from users where metadata @> '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata @> '?'));"},
		{"select * from users where metadata <@ '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata <@ '?'));"},
		{"select * from users where metadata ? '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ? '?'));"},
		{"select * from users where metadata ?| '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ?| '?'));"},
		{"select * from users where metadata ?& '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata ?& '?'));"},
		{"select * from users where metadata || '{\"age\": 42}';", 1, "(SELECT * FROM users WHERE (metadata || '?'));"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, 1, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "SELECT", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.SelectStatement. got=%T", tt.input, stmt)

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

func TestExpressionStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		{"(select name from users limit 1)", "(SELECT name FROM users LIMIT 1)"},                                                                                   // union with selects inside parens
		{"(select name from users limit 1) union (select name from people limit 1)", "((SELECT name FROM users LIMIT 1) UNION (SELECT name FROM people LIMIT 1))"}, // union with selects inside parens

	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\n\noutput: %s\n\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
	}
}
