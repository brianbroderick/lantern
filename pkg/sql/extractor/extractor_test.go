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

func TestExtractSelectedColumns(t *testing.T) {
	// maskParams := false
	t1 := time.Now()

	tests := []struct {
		input   string
		columns [][]string
	}{
		// no column (useful for exists functions)
		{"select from users;",
			[][]string{{}}},
		// super basic select
		{"select id from users;",
			[][]string{{"id"}}},
		// check for a wildcard with a table alias
		{"select u.* from users u;", [][]string{{"users.*"}}},
		// check that the asterisk is not treated as a wildcard
		{"select 2*3 from users;",
			[][]string{{}}},
		{`select 2 % 4 from users;`,
			[][]string{{}}},
		// PG allows + as a prefix operator
		{"select +2 -3 from users;",
			[][]string{{}}},
		// Negative numbers
		{"select -2 +3 from users;", [][]string{{}}},
		// check for double quotes around the table name
		{`select "blah".id from blah`, [][]string{{`"blah".id`}}},
		// math expression
		{"select 1 * (2 + (6 / 4)) - 9 from users;",
			[][]string{{}}},
		// multiple columns
		{"select id, name from users",
			[][]string{{"id", "name"}}},
		// underscore in a column name
		{"select id, first_name from users;",
			[][]string{{"id", "first_name"}}},
		// column alias with as
		{"select id, first_name as name from users",
			[][]string{{"id", "first_name"}}},
		// column alias
		{"select id, first_name name from users",
			[][]string{{"id", "first_name"}}},
		// column alias with table alias
		{"select u.id, u.first_name as name from users u;",
			[][]string{{"users.id", "users.first_name"}}},
		// no semicolon
		{"select id from no_semi_colons",
			[][]string{{"id"}}},
		// multiple column aliases with expressions
		{"select 1 + 2 as math, foo + 7 as seven from foo",
			[][]string{{"foo"}}},
		// more complex math expression
		{"select 1 + 2 * 3 / value as math from foo",
			[][]string{{"value"}}},
		// table alias
		{"select a.id from addresses a;",
			[][]string{{"addresses.id"}}},
		// function call
		{"select sum(a,b) from users;",
			[][]string{{"a", "b"}}},
		// removed the token KEY since it's not a PG reserved key word: https://www.postgresql.org/docs/13/sql-keywords-appendix.html
		{"select key, value from example where id = 20 AND key IN ( 'a', 'b', 'c' );",
			[][]string{{"key", "value"}}},
		// escaped apostrophes
		{"SELECT translate(name, '''', '' ) as name FROM people WHERE id = 0;",
			[][]string{{"name"}}},
		// coalesce
		{"select coalesce ( u.first_name || ' ' || u.last_name, u.first_name, u.last_name ) AS name from users u",
			[][]string{{"users.first_name", "users.last_name"}}},

		// Select: distinct & all tokens
		{"select distinct id from users;",
			[][]string{{"id"}}},
		{"select all id from users",
			[][]string{{"id"}}},
		{"select distinct on (location) reported_at, report from weather_reports;",
			[][]string{{"location", "reported_at", "report"}}},
		{"select c.id, string_agg ( distinct c.name, ', ' ) as value FROM companies c",
			[][]string{{"companies.id", "companies.name"}}},
		{"select array_agg(distinct sub.id) from sub",
			[][]string{{"sub.id"}}},

		// Select: window functions
		{"select avg(salary) over (partition by depname) from empsalary;",
			[][]string{{"salary", "depname"}}},
		{"select avg(salary) over (order by depname) from empsalary",
			[][]string{{"salary", "depname"}}},
		{"select avg(salary) over (partition by salary order by depname) from empsalary;",
			[][]string{{"salary", "depname"}}},
		{"select avg(salary) over (partition by salary order by depname desc) from empsalary",
			[][]string{{"salary", "depname"}}},
		{"select wf1() over w from table_name;",
			[][]string{{"w"}}},
		{"select wf1() over w, wf2() over w from table_name;",
			[][]string{{"w"}}},
		// TODO: remove window function from the column list, add window partition by and order by to the column list
		{"select wf1() over w, wf2() over w from table_name window w as (partition by c1 order by c2);",
			[][]string{{"w"}}},
		{"select wf1() over w, wf2() over w from table_name window w as (partition by c1 order by c2), foo as (partition by c3 order by c4);",
			[][]string{{"w"}}},

		// Select: joins
		{"select c.id from customers c join addresses a on c.id = a.customer_id;",
			[][]string{{"customers.id"}}},
		// {"select c.id from customers c join addresses a on (c.id = a.customer_id) join states s on (s.id = a.state_id);", [][]string{{"id", "name"}}},
		// // This is a complex join with multiple tables
		// {"select c.id, c.name from customers c join addresses a on c.id = a.customer_id join states s on s.id = a.state_id join phone_numbers ph ON ph.customer_id = c.id;",
		// 	[][]string{{"id", "name"}}},
		// {"select id from customers join addresses on id = customer_id join phones on id = phone_id;", [][]string{{"id", "name"}}},
		// {"select id from customers join addresses on customers.id = addresses.customer_id", [][]string{{"id", "name"}}},
		// {"select id from customers join addresses on id = customer_id where id = 46;", [][]string{{"id", "name"}}},
		// {"select id from customers left join addresses on id = customer_id;", [][]string{{"id", "name"}}},
		// {"select id from customers left outer join addresses on id = customer_id;", [][]string{{"id", "name"}}},
		// {"select id from addresses AS a JOIN states AS s ON (s.id = a.state_id AND s.code > 'ut')", [][]string{{"id", "name"}}},
		// {"SELECT r.* FROM roles r, rights ri WHERE r.id = ri.role_id AND ri.deleted_by IS NULL AND ri.id = 12;", [][]string{{"id", "name"}}},
		// {"select left('abc', 2);", [][]string{{"id", "name"}}},  // in this case, left is a string function getting the left 2 characters
		// {"select right('abc', 2);", [][]string{{"id", "name"}}}, // in this case, right is a string function getting the right 2 characters

		// Select: where clause
		{"select id from users where id = 42;",
			[][]string{{"id"}}},
		// {"select id from users where id = 42 and customer_id = 74", [][]string{{"id", "name"}}},
		// {"select id from users where id = 42 and customer_id > 74;", [][]string{{"id", "name"}}},
		// {"select id from users where name = 'brian';", [][]string{{"id", "name"}}},
		// {"select id from users where name = 'brian'", [][]string{{"id", "name"}}},
		// {"select id from users where name is null", [][]string{{"id", "name"}}},
		// {"select id from users where name is not null", [][]string{{"id", "name"}}},

		// Select: IS comparisons
		{"select category from users where category is null",
			[][]string{{"category"}}},
		// {"select category from users where category is not null", [][]string{{"id", "name"}}},
		// {"select category from users where category is null and type = 1", [][]string{{"id", "name"}}},
		// {"select category from users where category is true", [][]string{{"id", "name"}}},
		// {"select category from users where category is not true", [][]string{{"id", "name"}}},
		// {"select category from users where category is false", [][]string{{"id", "name"}}},
		// {"select category from users where category is not false", [][]string{{"id", "name"}}},
		// {"select category from users where category is unknown", [][]string{{"id", "name"}}},
		// {"select category from users where category is not unknown", [][]string{{"id", "name"}}},
		// {"select foo,bar from my_table where foo is distinct from bar;", [][]string{{"id", "name"}}},
		// {"select foo,bar from my_table where foo is not distinct from bar;", [][]string{{"id", "name"}}},

		// Select: group by
		{"select baz from sales group by bar;",
			[][]string{{"baz"}}},
		// {"select id from users group by id, name;", [][]string{{"id", "name"}}},

		// Select: combined clauses
		{"select id from users where id = 42 group by id, name",
			[][]string{{"id"}}},
		// {"select id from customers join addresses on id = customer_id where id = 46 group by id;", [][]string{{"id", "name"}}},

		// Select: having clause
		{"select id from users group by id having id > 2;",
			[][]string{{"id"}}},
		// {"select id from users group by id having id > 2 and name = 'frodo';", [][]string{{"id", "name"}}},

		// Select: order by
		{"select id from users order by id;",
			[][]string{{"id"}}},
		// {"select id from users order by id desc, name", [][]string{{"id", "name"}}},
		// {"select id from users order by id desc nulls first, name nulls last;", [][]string{{"id", "name"}}},

		// Select: limit
		{"select id from users limit 10;",
			[][]string{{"id"}}},
		// {"select id from users limit ALL;", [][]string{{"id", "name"}}},
		// {"select id from users limit ALL", [][]string{{"id", "name"}}},
		// {"select id from users order by id limit 34", [][]string{{"id", "name"}}},
		// {"select * from users order by id offset 0 limit 34", [][]string{{"id", "name"}}},
		// {"select * from users order by id limit 34 offset 0 ", [][]string{{"id", "name"}}},

		// Select: offset
		{"select id from users limit ALL offset 10;",
			[][]string{{"id"}}},
		// {"select id from users limit 10 offset 10;", [][]string{{"id", "name"}}},
		// {"select id from users limit 10 offset 1 ROW", [][]string{{"id", "name"}}},
		// {"select id from users limit 10 offset 2 ROWS;", [][]string{{"id", "name"}}},

		// Select: combined order by, limit, offset
		{"select id from users order by id desc limit 10 offset 10;",
			[][]string{{"id"}}},
		// {"select id from users order by id desc nulls last limit 10 offset 10;", [][]string{{"id", "name"}}},

		// Select: fetch
		{"select id from users order by id fetch first row only;",
			[][]string{{"id"}}},
		// {"select id from users order by id fetch first 3 rows only;", [][]string{{"id", "name"}}},
		// {"select id from users order by id fetch first 10 rows with ties;", [][]string{{"id", "name"}}},

		// Select: for update
		{"select id from users for update;",
			[][]string{{"id"}}},
		// {"select id from users for no key update;;", [][]string{{"id", "name"}}},
		// {"select id from users for share;", [][]string{{"id", "name"}}},
		// {"select id from users for key share", [][]string{{"id", "name"}}},
		// {"select id from users for update of users;", [][]string{{"id", "name"}}},
		// {"select id from users for update of users, addresses;", [][]string{{"id", "name"}}},
		// {"select id from users for update of users, addresses nowait;", [][]string{{"id", "name"}}},
		// {"select id from users for update of users, addresses skip locked;", [][]string{{"id", "name"}}},

		// Select: IN operator
		{"select id from users where id in ('1','2','3','4');",
			[][]string{{"id"}}},
		// {"select id from users where id not in ('1','2','3','4');", [][]string{{"id", "name"}}},
		// {"select id from users where id IN ('1','2','3','4') AND name = 'brian';", [][]string{{"id", "name"}}},
		// {"select id from users where id IN (1,2,3,4);", [][]string{{"id", "name"}}},
		// {"select id from modules where (option_id, external_id) IN ((1, 7))", [][]string{{"id", "name"}}},         // single tuple
		// {"select id from modules where (option_id, external_id) IN ((1, 7), (2, 9))", [][]string{{"id", "name"}}}, // multiple tuples
		// {"select option_id, external_id from modules group by option_id, external_id having (option_id, external_id) IN ((1, 7), (2, 9))", [][]string{{"id", "name"}}},
		// {"select position('b' IN 'brian') as foo from cars", [][]string{{"id", "name"}}}, // in inside a function call
		// {"select count(1) from cars c left join models m ON ( position((c.key) in m.formula)<>0 )", [][]string{{"id", "name"}}},
		// {"SELECT u.* from users u where u.id IN (42);", [][]string{{"id", "name"}}},

		// Select: LIKE operator
		// basic like
		{"select id from users where name like 'brian';",
			[][]string{{"id"}}},
		// {"select id from users where name not like 'brian';", [][]string{{"id", "name"}}},               // basic not like
		// {"select id from users where rownum between 1 and sample_size", [][]string{{"id", "name"}}},     // BETWEEN
		// {"select id from users where rownum not between 1 and sample_size", [][]string{{"id", "name"}}}, // BETWEEN
		// {"select if from users where rownum between 1 and sample_size group by property_id;", [][]string{{"id", "name"}}},
		// {"select * from mytable where mycolumn ~ 'regexp';", [][]string{{"id", "name"}}},   // basic regex (case sensitive)
		// {"select * from mytable where mycolumn ~* 'regexp';", [][]string{{"id", "name"}}},  // basic regex (case insensitive)
		// {"select * from mytable where mycolumn !~ 'regexp';", [][]string{{"id", "name"}}},  // basic not regex (case sensitive)
		// {"select * from mytable where mycolumn !~* 'regexp';", [][]string{{"id", "name"}}}, // basic not regex (case insensitive)
		// // {"select select 'abc' similar to 'abc' from users;", ""}, // TODO: handle similar to
		// // {"select select 'abc' not similar to 'abc' from users;", ""}, // TODO: handle similar to

		// Select: EXISTS operator. In this case, NOT is a prefix operator
		{"select id from users where exists (select id from addresses where user_id = users.id);",
			[][]string{{"id"}}},
		// {"select id from users where not exists (select id from addresses where user_id = users.id);",
		// 	[][]string{{"id", "name"}}},

		// Select: UNION clause
		// No tables
		{"SELECT '08/22/2023'::DATE;",
			[][]string{{}}},
		// {"SELECT '123' union select '456';", [][]string{{"id", "name"}}},
		// {"SELECT '08/22/2023'::DATE union select '08/23/2023'::DATE;", [][]string{{"id", "name"}}},
		// {"SELECT '123' union select '456';", [][]string{{"id", "name"}}},

		// Single tables
		{"select id from users union select id from customers;",
			[][]string{{"id"}}},
		// {"select id from users except select id from customers;", [][]string{{"id", "name"}}},
		// {"select id from users intersect select id from customers;", [][]string{{"id", "name"}}},
		// {"select id from users union all select id from customers;", [][]string{{"id", "name"}}},
		// {"select id from users except all select id from customers;", [][]string{{"id", "name"}}},
		// {"select id from users intersect all select id from customers;", [][]string{{"id", "name"}}},
		// {"select name from users union select fname from people", [][]string{{"id", "name"}}},

		// Select: Cast literals
		{"select '100'::integer from a;",
			[][]string{{}}},
		// {"select 100::text from a;", [][]string{{"id", "name"}}},
		// {"select a::text from b;", [][]string{{"id", "name"}}},
		// {"select load( array[ 1 ], array[ 2] ) from a", [][]string{{"id", "name"}}},
		// {"select array[2]::integer from a", [][]string{{"id", "name"}}},
		// {"select load( array[ 1 ]::integer[], array[ 2]::integer[] ) from a", [][]string{{"id", "name"}}},
		// {"select jsonb_array_length ( ( options ->> 'country_codes' ) :: jsonb ) from modules", [][]string{{"id", "name"}}},
		// {"select now()::timestamp from users;", [][]string{{"id", "name"}}},
		// {"select ( junk_drawer->>'ids' )::INT[] from dashboards", [][]string{{"id", "name"}}},
		// {"select end_date + '1 day' ::INTERVAL from some_dates;", [][]string{{"id", "name"}}},
		// {`select CAST(u.depth AS DECIMAL(18, 2)) from users`, [][]string{{"id", "name"}}},

		// Select: JSONB
		{"select id from users where data->'name' = 'brian';",
			[][]string{{"id"}}},
		// {"select id from users where data->>'name' = 'brian';", [][]string{{"id", "name"}}},
		// {"select id from users where data#>'{name}' = 'brian';", [][]string{{"id", "name"}}},
		// {"select id from users where data#>>'{name}' = 'brian';", [][]string{{"id", "name"}}},
		// {"select id from users where data#>>'{name,first}' = 'brian';", [][]string{{"id", "name"}}},
		// {"select id from users where data#>>'{name,first}' = 'brian' and data#>>'{name,last}' = 'broderick';", [][]string{{"id", "name"}}},
		// {"select * from users where metadata @> '{\"age\": 42}';", [][]string{{"id", "name"}}},
		// {"select * from users where metadata <@ '{\"age\": 42}';", [][]string{{"id", "name"}}},
		// {"select * from users where metadata ? '{\"age\": 42}';", [][]string{{"id", "name"}}},
		// {"select * from users where metadata ?| '{\"age\": 42}';", [][]string{{"id", "name"}}},
		// {"select * from users where metadata ?& '{\"age\": 42}';", [][]string{{"id", "name"}}},
		// {"select * from users where metadata || '{\"age\": 42}';", [][]string{{"id", "name"}}},

		// Select: ARRAYs
		{"select array['John'] from users;",
			[][]string{{}}},
		// {"select array['John', 'Joseph'] from users;", [][]string{{"id", "name"}}},
		// {"select array['John', 'Joseph', 'Anna', 'Henry'] && array['Henry', 'John'] from users;", [][]string{{"id", "name"}}},
		// {"select pay[3] FROM emp;", [][]string{{"id", "name"}}},
		// {"select pay[1:2] FROM emp;", [][]string{{"id", "name"}}},
		// {"select pay[1:] FROM emp;", [][]string{{"id", "name"}}},
		// {"select pay[:1] FROM emp;", [][]string{{"id", "name"}}},

		// Select: CASE
		{"select case when id = 1 then 'one' when id = 2 then 'two' else 'other' end from users;",
			[][]string{{"id"}}},
		// {`SELECT case when (type_id = 3) then 1 else 0 end::text as is_complete FROM users`, [][]string{{"id", "name"}}},
		// {"select array_agg(name order by name) as names from users", [][]string{{"id", "name"}}},
		// {"SELECT case when (id = 3) then 1 else 0 end AS names from users", [][]string{{"id", "name"}}},
		// {"SELECT (case when (id = 3) then 1 else 0 end) AS names from users", [][]string{{"id", "name"}}},
		// {"SELECT array_agg(case when id = 3 then 1 else 0 end order by name, id) AS names from users", [][]string{{"id", "name"}}},
		// {"SELECT array_agg(case when (uid = 3) then 1 else 0 end order by name) AS names from users",
		// 	[][]string{{"id", "name"}}},

		// Select: Functions
		{"select w() from a",
			[][]string{{}}},
		// {"select id from load(1,2)", [][]string{{"id", "name"}}},
		// {"select json_object_agg(foo order by id) from bar", [][]string{{"id", "name"}}},                                       // aggregate function with order by
		// {"select array_agg( distinct sub.name_first order by sub.name_first ) from customers sub", [][]string{{"id", "name"}}}, // aggregate with distinct

		// Subqueries
		{"select * from (select id from a) b order by id",
			[][]string{{"id"}}},
		// {"select id from (select id from users union select id from people) u;", [][]string{{"id", "name"}}},
		// {"SELECT id FROM ( SELECT id FROM users u UNION SELECT id FROM users u ) as SubQ ;", [][]string{{"id", "name"}}}, // with union

		// Select: With Ordinality
		{"select * from unnest(array [ 4, 2, 1, 3, 7 ]) ;",
			[][]string{{}}},
		// {"select * from unnest(array [ 4, 2, 1, 3, 7 ]) with ordinality;", [][]string{{"id", "name"}}},
		// {"select * from unnest(array [ 4, 2, 1, 3, 7 ]) with ordinality as t(key, index);", [][]string{{"id", "name"}}},

		// Select: reserved words
		// any
		{"select id from users where any(type_ids) = 10;",
			[][]string{{"id"}}},
		// {"select null::integer AS id from users;", [][]string{{"id", "name"}}},                // null
		// {"select id from users where login_date < current_date;", [][]string{{"id", "name"}}}, // CURRENT_DATE
		// {"select cast('100' as integer) from users", [][]string{{"id", "name"}}},              // cast
		// {"select id from account_triggers at", [][]string{{"id", "name"}}},
		// {"select u.id from users u join account_triggers at on at.user_id = u.id;", [][]string{{"id", "name"}}},
		// {"select v as values from users;", [][]string{{"id", "name"}}},
		// {`select e.details->>'values' as values	from events;`, [][]string{{"id", "name"}}},
		// {`select set.* from server_event_types as set;`, [][]string{{"id", "name"}}},
		// {`select e.* from events e join server_event_types as set on (set.id = e.server_event_type_id);`, [][]string{{"id", "name"}}},
		// {`select fname || lname as user from users;`, [][]string{{"id", "name"}}},
		// {`select 1 as order from users;`, [][]string{{"id", "name"}}},

		// // Less common expressions
		{"select count(*) as unfiltered from generate_series(1,10) as s(i)",
			[][]string{{"*"}}},
		// {"select COUNT(*) FILTER (WHERE i < 5) AS filtered from generate_series(1,10) s(i)", [][]string{{"id", "name"}}},
		// {"select trim(both 'x' from 'xTomxx') from users;", [][]string{{"id", "name"}}},
		// {"select trim(leading 'x' from 'xTomxx') from users;", [][]string{{"id", "name"}}},
		// {"select trim(trailing 'x' from 'xTomxx') from users;", [][]string{{"id", "name"}}},
		// {"select substring('or' from 'Hello World!') from users;", [][]string{{"id", "name"}}},
		// {"select substring('Hello World!' from 2 for 4) from users;", [][]string{{"id", "name"}}},
		// {"select id from extra where number ilike e'001';", [][]string{{"id", "name"}}}, // escapestring of the form e'...'
		// {"select id from extra where number ilike E'%6%';", [][]string{{"id", "name"}}}, // escapestring of the form E'...'
		// {"SELECT * FROM ( VALUES (41, 1), (42, 2), (43, 3), (44, 4), (45, 5), (46, 6) ) AS t ( id, type_id )", [][]string{{"id", "name"}}},

		// Intervals
		// TODO: current_date is actually a function, not a column
		{"select current_date - INTERVAL '7 DAY' from users;",
			[][]string{{"current_date"}}},
		// {"select 1 from users where a = interval '1' year", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' month", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' day", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' hour", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' minute", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' SECOND", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' year to month", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' day to hour", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' day to minute", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' day to second", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' hour to minute", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' hour to second", [][]string{{"id", "name"}}},
		// {"select 1 from users where a = interval '1' minute to second", [][]string{{"id", "name"}}},

		// Multi word keywords: AT TIME ZONE, and TIMESTAMP WITH TIME ZONE
		{"select id from my_table where '2020-01-01' at time zone 'MDT' = '2023-01-01';",
			[][]string{{"id"}}},
		// {"select * from tasks where date_trunc('day', created_at) = date_trunc('day', now()::timestamp with time zone at time zone 'America/Denver') LIMIT 1;", [][]string{{"id", "name"}}},
		// {"select now()::timestamp with time zone from users;", [][]string{{"id", "name"}}},
		// {"select '2020-01-01' at time zone 'MDT' from my_table;", [][]string{{"id", "name"}}},
		// {"select '2020-01-01' + 'MDT' from my_table;", [][]string{{"id", "name"}}},
		// {"select id from my_table where my_date at time zone my_zone > '2001-01-01';", [][]string{{"id", "name"}}},
		// {`select timestamp '2001-09-23';`, [][]string{{"id", "name"}}},
		// {`select timestamp '2001-09-23' + interval '72 hours';`, [][]string{{"id", "name"}}},

		// TODO: Doesn't handle CTE yet
		// Select with a CTE expression
		// {"select count(1) from (with my_list as (select i from generate_series(1,10) s(i)) select i from my_list where i > 5) as t;",
		// 	[][]string{{"i"}}},

		// Dots
		{`select id as "my.id" from users u`,
			[][]string{{"id"}}},
		// {`select id as "my" from users u`, [][]string{{"id", "name"}}},
		// {"select u . id from users u", [][]string{{"id", "name"}}}, // "no prefix parse function for DOT found at line 0 char 25"

		// Comments
		// multiple single line comments
		{`select *
			-- id
			-- name
				from users u
				JOIN addresses a ON (u.id = a.user_id)
				where id = 42;`,
			[][]string{{"id", "name"}}},
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

			for fqcn, column := range r.Columns {
				if column.Clause == token.COLUMN {
					assert.Contains(t, tt.columns[i], fqcn, "input: %s\nColumn %s not found in %v", tt.input, fqcn, tt.columns[i])
				}
			}
		}
	}
	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	fmt.Printf("TestExtractAlias, Elapsed Time: %s\n", timeDiff)
}

// func TestExtractAlias(t *testing.T) {
// 	maskParams := false
// 	t1 := time.Now()

// 	tests := []struct {
// 		input   string
// 		output  string
// 		columns [][]string
// 		tables  [][]string
// 		joins   [][]string
// 	}{
// 		{"select coalesce(id, name) from users", "(SELECT coalesce(id, name) FROM users);",
// 			[][]string{{"id", "name"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select * from users; select * from addresses", "(SELECT * FROM users);(SELECT * FROM addresses);",
// 			[][]string{{"users"}, {"addresses"}}, [][]string{{"users"}, {"addresses"}}, [][]string{{"users"}, {"addresses"}}},
// 		{"select * from users u", "(SELECT * FROM users);",
// 			[][]string{{"users.*"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select u.id from users u", "(SELECT users.id FROM users);",
// 			[][]string{{"users.id"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select u.id, u.name from users u", "(SELECT users.id, users.name FROM users);",
// 			[][]string{{"users.id", "users.name"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select u.id, u.name as my_name from users u", "(SELECT users.id, users.name AS my_name FROM users);",
// 			[][]string{{"users.id", "users.name"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select u.id + 7 as my_alias from users u", "(SELECT (users.id + 7) AS my_alias FROM users);",
// 			[][]string{{"users.id"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select u.* from users u;", "(SELECT users.* FROM users);",
// 			[][]string{{"users.*"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select coalesce ( u.first_name || ' ' || u.last_name, u.first_name, u.last_name ) AS name from users u",
// 			"(SELECT coalesce(((users.first_name || ' ') || users.last_name), users.first_name, users.last_name) AS name FROM users);",
// 			[][]string{{"users.first_name", "users.last_name"}}, [][]string{{"users"}}, [][]string{{}}}, // coalesce
// 		{"select c.id from customers c join addresses a on c.id = a.customer_id;",
// 			"(SELECT customers.id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id));",
// 			[][]string{{"customers.id"}}, [][]string{{"customers", "addresses"}}, [][]string{{"customers", "addresses"}}},
// 		{"select c.id from customers c join addresses a on (c.id = a.customer_id) join states s on (s.id = a.state_id);",
// 			"(SELECT customers.id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id) INNER JOIN states ON (states.id = addresses.state_id));",
// 			[][]string{{"customers.id"}}, [][]string{{"customers", "addresses", "states"}}, [][]string{{"customers", "addresses", "states"}}},
// 		{"select id from addresses AS a JOIN states AS s ON (s.id = a.state_id AND s.code > 'ut')",
// 			"(SELECT id FROM addresses INNER JOIN states ON ((states.id = addresses.state_id) AND (states.code > 'ut')));",
// 			[][]string{{"id"}}, [][]string{{"addresses", "states"}}, [][]string{{"states", "addresses"}}},
// 		{"SELECT r.* FROM roles r, rights ri WHERE r.id = ri.role_id AND ri.deleted_by IS NULL AND ri.id = 12;",
// 			"(SELECT roles.* FROM roles , rights WHERE (((roles.id = rights.role_id) AND (rights.deleted_by IS NULL)) AND (rights.id = 12)));",
// 			[][]string{{"roles.*"}}, [][]string{{"roles", "rights"}}, [][]string{{}}},
// 		{"select left('abc', 2);", "(SELECT left('abc', 2));",
// 			[][]string{{}}, [][]string{{}}, [][]string{{}}}, // in this case, left is a string function getting the left 2 characters
// 		{"select u.category from users u where u.category is true", "(SELECT users.category FROM users WHERE (users.category IS TRUE));",
// 			[][]string{{"users.category"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select s.baz from sales s group by s.bar;", "(SELECT sales.baz FROM sales GROUP BY sales.bar);",
// 			[][]string{{"sales.baz"}}, [][]string{{"sales"}}, [][]string{{}}},
// 		{"select s.baz from sales s group by s.bar having s.baz > 5;",
// 			"(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5));",
// 			[][]string{{"sales.baz"}}, [][]string{{"sales"}}, [][]string{{}}},
// 		{"select s.baz from sales s group by s.bar having s.baz > 5 order by s.baz;",
// 			"(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5) ORDER BY sales.baz);",
// 			[][]string{{"sales.baz"}}, [][]string{{"sales"}}, [][]string{{}}},
// 		{"select s.baz from sales s group by s.bar having s.baz > 5 order by s.baz desc;",
// 			"(SELECT sales.baz FROM sales GROUP BY sales.bar HAVING (sales.baz > 5) ORDER BY sales.baz DESC);",
// 			[][]string{{"sales.baz"}}, [][]string{{"sales"}}, [][]string{{}}},
// 		{"select sub.id from (select id from users) sub;", "(SELECT sub.id FROM (SELECT id FROM users) sub);",
// 			[][]string{{"sub.id", "id"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select sub.id from (select u.id from users u) sub;", "(SELECT sub.id FROM (SELECT users.id FROM users) sub);",
// 			[][]string{{"sub.id", "users.id"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select car.id from (select c.id from cars c) car join customers c on c.car_id = car.id;",
// 			"(SELECT car.id FROM (SELECT cars.id FROM cars) car INNER JOIN customers ON (customers.car_id = car.id));",
// 			[][]string{{"car.id", "cars.id"}}, [][]string{{"cars", "customers"}}, [][]string{{"customers", "car"}}},
// 		{"select u.id from users u where u.id = 42;", "(SELECT users.id FROM users WHERE (users.id = 42));",
// 			[][]string{{"users.id"}}, [][]string{{"users"}}, [][]string{{}}},
// 		{"select car.id from (select c.id from cars c) car join customers c on c.car_id = car.id inner join users u on u.id = c.user_id;",
// 			"(SELECT car.id FROM (SELECT cars.id FROM cars) car INNER JOIN customers ON (customers.car_id = car.id) INNER JOIN users ON (users.id = customers.user_id));",
// 			[][]string{{"car.id", "cars.id"}}, [][]string{{"cars", "customers", "users"}}, [][]string{{"customers", "car", "users"}}},
// 	}

// 	// TODO: add table to non-aliased columns
// 	// TODO: handle wildcards better

// 	for _, tt := range tests {
// 		l := lexer.New(tt.input)
// 		p := parser.New(l)
// 		program := p.ParseProgram()

// 		for i, s := range program.Statements {
// 			r := NewExtractor(&s)
// 			env := object.NewEnvironment()
// 			r.Extract(s, env)
// 			checkExtractErrors(t, r, tt.input)

// 			assert.Equal(t, len(tt.tables[i]), len(r.Tables), "input: %s\nNumber of tables not equal", tt.input)

// 			for _, table := range r.Tables {
// 				assert.Contains(t, tt.tables[i], table.Name, "input: %s\nTable %s not found in %v", tt.input, table.Name, tt.tables[i])
// 			}

// 			for fqcn, column := range r.Columns {
// 				if column.Clause == token.COLUMN {
// 					assert.Contains(t, tt.columns[i], fqcn, "input: %s\nColumn %s not found in %v", tt.input, fqcn, tt.columns[i])
// 				}
// 			}

// 			for _, join := range r.TableJoins {
// 				found := 0
// 				for _, j := range tt.joins[i] {
// 					if join.TableA == j {
// 						found++
// 					}
// 					if join.TableB == j {
// 						found++
// 					}
// 				}
// 				assert.Equal(t, 2, found, "input: %s\nDid not find all input tables: %v in join: %s %s\n",
// 					tt.input, tt.joins[i], printTableJoins(r))
// 			}

// 		}

// 		output := program.String(maskParams)
// 		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s\n", tt.input, tt.output, output)

// 	}
// 	t2 := time.Now()
// 	timeDiff := t2.Sub(t1)
// 	fmt.Printf("TestExtractAlias, Elapsed Time: %s\n", timeDiff)
// }

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
