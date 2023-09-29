package parser

import (
	"fmt"
	"testing"

	"github.com/brianbroderick/lantern/internal/sql/ast"
	"github.com/brianbroderick/lantern/internal/sql/lexer"
	"github.com/stretchr/testify/assert"
)

// join_type table     alias join_condition
// source    customers c
// inner     addresses a     Expression

// select * from users join addresses on users.id = addresses.user_id;
// select * from users u inner join addresses a on u.id = a.user_id;

// Window frame_clause:
// SELECT
// wf1() OVER w,
// wf2() OVER w,
// FROM table_name
// WINDOW w AS (PARTITION BY c1 ORDER BY c2);

// Subquery:
// SELECT	film_id, title,	rental_rate FROM film WHERE	rental_rate > (	SELECT AVG (rental_rate) FROM film);
// SELECT film_id, title FROM	film WHERE film_id IN (SELECT inventory.film_id	FROM rental INNER JOIN inventory ON inventory.inventory_id = rental.inventory_id WHERE return_date BETWEEN '2005-05-29'	AND '2005-05-30');

// Common Table Expression:
// WITH regional_sales AS (
// 	SELECT region, SUM(amount) AS total_sales
// 	FROM orders
// 	GROUP BY region
// ), top_regions AS (
// 	SELECT region
// 	FROM regional_sales
// 	WHERE total_sales > (SELECT SUM(total_sales)/10 FROM regional_sales)
// )
// SELECT region,
// 		 product,
// 		 SUM(quantity) AS product_units,
// 		 SUM(amount) AS product_sales
// FROM orders
// WHERE region IN (SELECT region FROM top_regions)
// GROUP BY region, product;

func TestSubSelects(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		// Select: Subqueries
		{"select film_id, title, rental_rate from film where rental_rate > (select avg(rental_rate) from film);", "(SELECT film_id, title, rental_rate FROM film WHERE (rental_rate > (SELECT avg(rental_rate) FROM film)));"},
		{"select id from customers where id > (select avg(id) from customers where id > (select min(id) from customers));", "(SELECT id FROM customers WHERE (id > (SELECT avg(id) FROM customers WHERE (id > (SELECT min(id) FROM customers)))));"},
		{"select film_id from film where rental_rate > (select avg(rental_rate) from film) order by film_id;", "(SELECT film_id FROM film WHERE (rental_rate > (SELECT avg(rental_rate) FROM film)) ORDER BY film_id);"},
		{"select id from customers where id > (select avg(id) from customers where id > (42)) order by id desc;", "(SELECT id FROM customers WHERE (id > (SELECT avg(id) FROM customers WHERE (id > 42))) ORDER BY id DESC);"},
		{"select id from customers where id > (select id from customers where id > (select id from customers)) order by id desc;", "(SELECT id FROM customers WHERE (id > (SELECT id FROM customers WHERE (id > (SELECT id FROM customers)))) ORDER BY id DESC);"},
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

		output := program.String()
		assert.Equal(t, tt.output, output, "program.String() not '%s'. got=%s", tt.output, output)
		fmt.Printf("output: %s\n", output)
	}
}

// func TestSingleSelectStatements(t *testing.T) {
// 	tests := []struct {
// 		input      string
// 		tableCount int
// 		output     string
// 	}{
// 		// Select: simple
// 		{"select id from users;", 1, "(SELECT id FROM users);"},
// 		{"select id, name from users", 1, "(SELECT id, name FROM users);"},
// 		{"select id, first_name from users;", 1, "(SELECT id, first_name FROM users);"},
// 		{"select id, first_name as name from users", 1, "(SELECT id, first_name AS name FROM users);"},
// 		{"select u.id, u.first_name as name from users u;", 1, "(SELECT u.id, u.first_name AS name FROM users u);"},
// 		{"select id from no_semi_colons", 1, "(SELECT id FROM no_semi_colons);"},
// 		{"select 1 + 2 as math, foo + 7 as seven from foo", 1, "(SELECT (1 + 2) AS math, (foo + 7) AS seven FROM foo);"},
// 		{"select 1 + 2 * 3 / value as math from foo", 1, "(SELECT (1 + ((2 * 3) / value)) AS math FROM foo);"},
// 		{"select id from addresses a;", 1, "(SELECT id FROM addresses a);"},
// 		{"select \"blah\".id from users", 1, "(SELECT blah.id FROM users);"},
// 		{"select sum(a,b) from users;", 1, "(SELECT sum(a, b) FROM users);"},

// 		// Select: distinct & all tokens
// 		{"select distinct id from users;", 1, "(SELECT DISTINCT id FROM users);"},
// 		{"select all id from users", 1, "(SELECT ALL id FROM users);"},
// 		{"select distinct on (location) time, report from weather_reports;", 1, "(SELECT DISTINCT ON (location) time, report FROM weather_reports);"},

// 		// Select: window functions
// 		{"select avg(salary) over (partition by depname) from empsalary;", 1, "(SELECT (avg(salary) OVER (PARTITION BY depname)) FROM empsalary);"},
// 		{"select avg(salary) over (order by depname) from empsalary", 1, "(SELECT (avg(salary) OVER (ORDER BY depname)) FROM empsalary);"},
// 		{"select avg(salary) over (partition by salary order by depname) from empsalary;", 1, "(SELECT (avg(salary) OVER (PARTITION BY salary ORDER BY depname)) FROM empsalary);"},
// 		{"select avg(salary) over (partition by salary order by depname desc) from empsalary", 1, "(SELECT (avg(salary) OVER (PARTITION BY salary ORDER BY depname DESC)) FROM empsalary);"},

// 		// Select: joins
// 		{"select c.id from customers c join addresses a on c.id = a.customer_id;", 2, "(SELECT c.id FROM customers c INNER JOIN addresses a ON (c.id = a.customer_id));"},
// 		{"select id from customers join addresses on id = customer_id", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id));"},
// 		{"select id from customers join addresses on id = customer_id join phones on id = phone_id;", 3, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) INNER JOIN phones ON (id = phone_id));"},
// 		{"select id from customers join addresses on customers.id = addresses.customer_id", 2, "(SELECT id FROM customers INNER JOIN addresses ON (customers.id = addresses.customer_id));"},
// 		{"select id from customers join addresses on id = customer_id where id = 46;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = 46));"},

// 		// Select: where clause
// 		{"select id from users where id = 42;", 1, "(SELECT id FROM users WHERE (id = 42));"},
// 		{"select id from users where id = 42 and customer_id = 74", 1, "(SELECT id FROM users WHERE ((id = 42) AND (customer_id = 74)));"},
// 		{"select id from users where id = 42 and customer_id > 74;", 1, "(SELECT id FROM users WHERE ((id = 42) AND (customer_id > 74)));"},
// 		{"select id from users where name = 'brian';", 1, "(SELECT id FROM users WHERE (name = 'brian'));"},
// 		{"select id from users where name = 'brian'", 1, "(SELECT id FROM users WHERE (name = 'brian'));"},

// 		// Select: group by
// 		{"select id from users group by id", 1, "(SELECT id FROM users GROUP BY id);"},
// 		{"select id from users group by id, name;", 1, "(SELECT id FROM users GROUP BY id, name);"},

// 		// Select: combined clauses
// 		{"select id from users where id = 42 group by id, name", 1, "(SELECT id FROM users WHERE (id = 42) GROUP BY id, name);"},
// 		{"select id from customers join addresses on id = customer_id where id = 46 group by id;", 2, "(SELECT id FROM customers INNER JOIN addresses ON (id = customer_id) WHERE (id = 46) GROUP BY id);"},

// 		// Select: having clause
// 		{"select id from users group by id having id > 2;", 1, "(SELECT id FROM users GROUP BY id HAVING (id > 2));"},
// 		{"select id from users group by id having id > 2 and name = 'frodo';", 1, "(SELECT id FROM users GROUP BY id HAVING ((id > 2) AND (name = 'frodo')));"},

// 		// Select: order by
// 		{"select id from users order by id;", 1, "(SELECT id FROM users ORDER BY id);"},
// 		{"select id from users order by id desc, name", 1, "(SELECT id FROM users ORDER BY id DESC, name);"},
// 		{"select id from users order by id desc nulls first, name nulls last;", 1, "(SELECT id FROM users ORDER BY id DESC NULLS FIRST, name NULLS LAST);"},

// 		// Select: limit
// 		{"select id from users limit 10;", 1, "(SELECT id FROM users LIMIT 10);"},
// 		{"select id from users limit ALL;", 1, "(SELECT id FROM users LIMIT ALL);"},
// 		{"select id from users limit ALL", 1, "(SELECT id FROM users LIMIT ALL);"},

// 		// Select: offset
// 		{"select id from users limit ALL offset 10;", 1, "(SELECT id FROM users LIMIT ALL OFFSET 10);"},
// 		{"select id from users limit 10 offset 10;", 1, "(SELECT id FROM users LIMIT 10 OFFSET 10);"},
// 		{"select id from users limit 10 offset 1 ROW", 1, "(SELECT id FROM users LIMIT 10 OFFSET 1);"},
// 		{"select id from users limit 10 offset 2 ROWS;", 1, "(SELECT id FROM users LIMIT 10 OFFSET 2);"},

// 		// Select: combined order by, limit, offset
// 		{"select id from users order by id desc limit 10 offset 10;", 1, "(SELECT id FROM users ORDER BY id DESC LIMIT 10 OFFSET 10);"},
// 		{"select id from users order by id desc nulls last limit 10 offset 10;", 1, "(SELECT id FROM users ORDER BY id DESC NULLS LAST LIMIT 10 OFFSET 10);"},

// 		// Select: fetch
// 		{"select id from users order by id fetch first row only;", 1, "(SELECT id FROM users ORDER BY id FETCH NEXT 1 ROWS ONLY);"},
// 		{"select id from users order by id fetch first 3 rows only;", 1, "(SELECT id FROM users ORDER BY id FETCH NEXT 3 ROWS ONLY);"},
// 		{"select id from users order by id fetch first 10 rows with ties;", 1, "(SELECT id FROM users ORDER BY id FETCH NEXT 10 ROWS WITH TIES);"},

// 		// Select: for update
// 		{"select id from users for update;", 1, "(SELECT id FROM users FOR UPDATE);"},
// 		{"select id from users for no key update;;", 1, "(SELECT id FROM users FOR NO KEY UPDATE);"},
// 		{"select id from users for share;", 1, "(SELECT id FROM users FOR SHARE);"},
// 		{"select id from users for key share", 1, "(SELECT id FROM users FOR KEY SHARE);"},
// 		{"select id from users for update of users;", 1, "(SELECT id FROM users FOR UPDATE OF users);"},
// 		{"select id from users for update of users, addresses;", 1, "(SELECT id FROM users FOR UPDATE OF users, addresses);"},
// 		{"select id from users for update of users, addresses nowait;", 1, "(SELECT id FROM users FOR UPDATE OF users, addresses NOWAIT);"},
// 		{"select id from users for update of users, addresses skip locked;", 1, "(SELECT id FROM users FOR UPDATE OF users, addresses SKIP LOCKED);"},

// 		// Select: IN clause
// 		{"select id from users where id IN ('1','2','3','4');", 1, "(SELECT id FROM users WHERE id IN ('1', '2', '3', '4'));"},
// 		{"select id from users where id IN ('1','2','3','4') AND name = 'brian';", 1, "(SELECT id FROM users WHERE (id IN ('1', '2', '3', '4') AND (name = 'brian')));"},
// 	}

// 	for _, tt := range tests {
// 		// fmt.Printf("\ninput:  %s\n", tt.input)
// 		l := lexer.New(tt.input)
// 		p := New(l)
// 		program := p.ParseProgram()
// 		checkParserErrors(t, p)

// 		assert.Equal(t, 1, len(program.Statements), "program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))

// 		stmt := program.Statements[0]
// 		assert.Equal(t, "select", stmt.TokenLiteral(), "program.Statements[0] is not ast.SelectStatement. got=%T", stmt)

// 		selectStmt, ok := stmt.(*ast.SelectStatement)
// 		assert.True(t, ok, "stmt is not *ast.SelectStatement. got=%T", stmt)

// 		// fmt.Printf("%+v\n", selectStmt.Tables[0])

// 		selectExp, ok := selectStmt.Expressions[0].(*ast.SelectExpression)
// 		assert.True(t, ok, "stmt is not *ast.SelectExpression. got=%T", selectExp)

// 		assert.Equal(t, tt.tableCount, len(selectExp.Tables), "len(selectStmt.Tables) not %d. got=%d", tt.tableCount, len(selectExp.Tables))
// 		output := program.String()
// 		assert.Equal(t, tt.output, output, "program.String() not '%s'. got=%s", tt.output, output)
// 		// fmt.Printf("output: %s\n", output)
// 		// for _, columns := range selectStmt.Columns {
// 		// 	fmt.Printf("\n\nselect: %+v\n\n", columns)
// 		// }
// 	}
// }

// func TestIdentifierExpression(t *testing.T) {
// 	input := "foobar;"

// 	l := lexer.New(input)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	if len(program.Statements) != 1 {
// 		t.Fatalf("program has not enough statements. got=%d",
// 			len(program.Statements))
// 	}
// 	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
// 	if !ok {
// 		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
// 			program.Statements[0])
// 	}

// 	ident, ok := stmt.Expression.(*ast.Identifier)
// 	if !ok {
// 		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
// 	}
// 	if ident.Value != "foobar" {
// 		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
// 	}
// 	if ident.TokenLiteral() != "foobar" {
// 		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar",
// 			ident.TokenLiteral())
// 	}
// }

// func TestIntegerLiteralExpression(t *testing.T) {
// 	input := "5;"

// 	l := lexer.New(input)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	if len(program.Statements) != 1 {
// 		t.Fatalf("program has not enough statements. got=%d",
// 			len(program.Statements))
// 	}
// 	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
// 	if !ok {
// 		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
// 			program.Statements[0])
// 	}

// 	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
// 	if !ok {
// 		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
// 	}
// 	if literal.Value != 5 {
// 		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
// 	}
// 	if literal.TokenLiteral() != "5" {
// 		t.Errorf("literal.TokenLiteral not %s. got=%s", "5",
// 			literal.TokenLiteral())
// 	}
// }

// func TestParsingPrefixExpressions(t *testing.T) {
// 	prefixTests := []struct {
// 		input    string
// 		operator string
// 		value    interface{}
// 	}{
// 		{"!5;", "!", 5},
// 		{"-15;", "-", 15},
// 		{"!foobar;", "!", "foobar"},
// 		{"-foobar;", "-", "foobar"},
// 		{"!true;", "!", true},
// 		{"!false;", "!", false},
// 	}

// 	for _, tt := range prefixTests {
// 		l := lexer.New(tt.input)
// 		p := New(l)
// 		program := p.ParseProgram()
// 		checkParserErrors(t, p)

// 		if len(program.Statements) != 1 {
// 			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
// 				1, len(program.Statements))
// 		}

// 		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
// 		if !ok {
// 			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
// 				program.Statements[0])
// 		}

// 		exp, ok := stmt.Expression.(*ast.PrefixExpression)
// 		if !ok {
// 			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
// 		}
// 		if exp.Operator != tt.operator {
// 			t.Fatalf("exp.Operator is not '%s'. got=%s",
// 				tt.operator, exp.Operator)
// 		}
// 		if !testLiteralExpression(t, exp.Right, tt.value) {
// 			return
// 		}
// 	}
// }

// func TestParsingInfixExpressions(t *testing.T) {
// 	infixTests := []struct {
// 		input      string
// 		leftValue  interface{}
// 		operator   string
// 		rightValue interface{}
// 	}{
// 		{"5 + 5;", 5, "+", 5},
// 		{"5 - 5;", 5, "-", 5},
// 		{"5 * 5;", 5, "*", 5},
// 		{"5 / 5;", 5, "/", 5},
// 		{"5 > 5;", 5, ">", 5},
// 		{"5 < 5;", 5, "<", 5},
// 		{"5 == 5;", 5, "==", 5},
// 		{"5 != 5;", 5, "!=", 5},
// 		{"foobar + barfoo;", "foobar", "+", "barfoo"},
// 		{"foobar - barfoo;", "foobar", "-", "barfoo"},
// 		{"foobar * barfoo;", "foobar", "*", "barfoo"},
// 		{"foobar / barfoo;", "foobar", "/", "barfoo"},
// 		{"foobar > barfoo;", "foobar", ">", "barfoo"},
// 		{"foobar < barfoo;", "foobar", "<", "barfoo"},
// 		{"foobar == barfoo;", "foobar", "==", "barfoo"},
// 		{"foobar != barfoo;", "foobar", "!=", "barfoo"},
// 		{"true == true", true, "==", true},
// 		{"true != false", true, "!=", false},
// 		{"false == false", false, "==", false},
// 	}

// 	for _, tt := range infixTests {
// 		l := lexer.New(tt.input)
// 		p := New(l)
// 		program := p.ParseProgram()
// 		checkParserErrors(t, p)

// 		if len(program.Statements) != 1 {
// 			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
// 				1, len(program.Statements))
// 		}

// 		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
// 		if !ok {
// 			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
// 				program.Statements[0])
// 		}

// 		if !testInfixExpression(t, stmt.Expression, tt.leftValue,
// 			tt.operator, tt.rightValue) {
// 			return
// 		}
// 	}
// }

// func TestOperatorPrecedenceParsing(t *testing.T) {
// 	tests := []struct {
// 		input    string
// 		expected string
// 	}{
// 		{
// 			"-a * b",
// 			"((-a) * b)",
// 		},
// 		{
// 			"!-a",
// 			"(!(-a))",
// 		},
// 		{
// 			"a + b + c",
// 			"((a + b) + c)",
// 		},
// 		{
// 			"a + b - c",
// 			"((a + b) - c)",
// 		},
// 		{
// 			"a * b * c",
// 			"((a * b) * c)",
// 		},
// 		{
// 			"a * b / c",
// 			"((a * b) / c)",
// 		},
// 		{
// 			"a + b / c",
// 			"(a + (b / c))",
// 		},
// 		{
// 			"a + b * c + d / e - f",
// 			"(((a + (b * c)) + (d / e)) - f)",
// 		},
// 		{
// 			"3 + 4; -5 * 5",
// 			"(3 + 4)((-5) * 5)",
// 		},
// 		{
// 			"5 > 4 == 3 < 4",
// 			"((5 > 4) == (3 < 4))",
// 		},
// 		{
// 			"5 < 4 != 3 > 4",
// 			"((5 < 4) != (3 > 4))",
// 		},
// 		{
// 			"3 + 4 * 5 == 3 * 1 + 4 * 5",
// 			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
// 		},
// 		{
// 			"true",
// 			"true",
// 		},
// 		{
// 			"false",
// 			"false",
// 		},
// 		{
// 			"3 > 5 == false",
// 			"((3 > 5) == false)",
// 		},
// 		{
// 			"3 < 5 == true",
// 			"((3 < 5) == true)",
// 		},
// 		{
// 			"1 + (2 + 3) + 4",
// 			"((1 + (2 + 3)) + 4)",
// 		},
// 		{
// 			"(5 + 5) * 2",
// 			"((5 + 5) * 2)",
// 		},
// 		{
// 			"2 / (5 + 5)",
// 			"(2 / (5 + 5))",
// 		},
// 		{
// 			"(5 + 5) * 2 * (5 + 5)",
// 			"(((5 + 5) * 2) * (5 + 5))",
// 		},
// 		{
// 			"-(5 + 5)",
// 			"(-(5 + 5))",
// 		},
// 		{
// 			"!(true == true)",
// 			"(!(true == true))",
// 		},
// 		{
// 			"a + add(b * c) + d",
// 			"((a + add((b * c))) + d)",
// 		},
// 		{
// 			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
// 			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
// 		},
// 		{
// 			"add(a + b + c * d / f + g)",
// 			"add((((a + b) + ((c * d) / f)) + g))",
// 		},
// 		{
// 			"a * [1, 2, 3, 4][b * c] * d",
// 			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
// 		},
// 		{
// 			"add(a * b[2], b[1], 2 * [1, 2][1])",
// 			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
// 		},
// 	}

// 	for _, tt := range tests {
// 		l := lexer.New(tt.input)
// 		p := New(l)
// 		program := p.ParseProgram()
// 		checkParserErrors(t, p)

// 		actual := program.String()
// 		if actual != tt.expected {
// 			t.Errorf("expected=%q, got=%q", tt.expected, actual)
// 		}
// 	}
// }

// func TestBooleanExpression(t *testing.T) {
// 	tests := []struct {
// 		input           string
// 		expectedBoolean bool
// 	}{
// 		{"true;", true},
// 		{"false;", false},
// 	}

// 	for _, tt := range tests {
// 		l := lexer.New(tt.input)
// 		p := New(l)
// 		program := p.ParseProgram()
// 		checkParserErrors(t, p)

// 		if len(program.Statements) != 1 {
// 			t.Fatalf("program has not enough statements. got=%d",
// 				len(program.Statements))
// 		}

// 		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
// 		if !ok {
// 			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
// 				program.Statements[0])
// 		}

// 		boolean, ok := stmt.Expression.(*ast.Boolean)
// 		if !ok {
// 			t.Fatalf("exp not *ast.Boolean. got=%T", stmt.Expression)
// 		}
// 		if boolean.Value != tt.expectedBoolean {
// 			t.Errorf("boolean.Value not %t. got=%t", tt.expectedBoolean,
// 				boolean.Value)
// 		}
// 	}
// }

// func TestCallExpressionParsing(t *testing.T) {
// 	input := "add(1, 2 * 3, 4 + 5);"

// 	l := lexer.New(input)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	if len(program.Statements) != 1 {
// 		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
// 			1, len(program.Statements))
// 	}

// 	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
// 	if !ok {
// 		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
// 			program.Statements[0])
// 	}

// 	exp, ok := stmt.Expression.(*ast.CallExpression)
// 	if !ok {
// 		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
// 			stmt.Expression)
// 	}

// 	if !testIdentifier(t, exp.Function, "add") {
// 		return
// 	}

// 	if len(exp.Arguments) != 3 {
// 		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
// 	}

// 	testLiteralExpression(t, exp.Arguments[0], 1)
// 	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
// 	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
// }

// func TestCallExpressionParameterParsing(t *testing.T) {
// 	tests := []struct {
// 		input         string
// 		expectedIdent string
// 		expectedArgs  []string
// 	}{
// 		{
// 			input:         "add();",
// 			expectedIdent: "add",
// 			expectedArgs:  []string{},
// 		},
// 		{
// 			input:         "add(1);",
// 			expectedIdent: "add",
// 			expectedArgs:  []string{"1"},
// 		},
// 		{
// 			input:         "add(1, 2 * 3, 4 + 5);",
// 			expectedIdent: "add",
// 			expectedArgs:  []string{"1", "(2 * 3)", "(4 + 5)"},
// 		},
// 	}

// 	for _, tt := range tests {
// 		l := lexer.New(tt.input)
// 		p := New(l)
// 		program := p.ParseProgram()
// 		checkParserErrors(t, p)

// 		stmt := program.Statements[0].(*ast.ExpressionStatement)
// 		exp, ok := stmt.Expression.(*ast.CallExpression)
// 		if !ok {
// 			t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
// 				stmt.Expression)
// 		}

// 		if !testIdentifier(t, exp.Function, tt.expectedIdent) {
// 			return
// 		}

// 		if len(exp.Arguments) != len(tt.expectedArgs) {
// 			t.Fatalf("wrong number of arguments. want=%d, got=%d",
// 				len(tt.expectedArgs), len(exp.Arguments))
// 		}

// 		for i, arg := range tt.expectedArgs {
// 			if exp.Arguments[i].String() != arg {
// 				t.Errorf("argument %d wrong. want=%q, got=%q", i,
// 					arg, exp.Arguments[i].String())
// 			}
// 		}
// 	}
// }

// // String literals are single quoted in SQL
// func TestStringLiteralExpression(t *testing.T) {
// 	input := "'hello world';"

// 	l := lexer.New(input)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	stmt := program.Statements[0].(*ast.ExpressionStatement)
// 	literal, ok := stmt.Expression.(*ast.StringLiteral)
// 	if !ok {
// 		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
// 	}

// 	if literal.Value != "hello world" {
// 		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
// 	}
// }

// func TestParsingEmptyArrayLiterals(t *testing.T) {
// 	input := "[]"

// 	l := lexer.New(input)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
// 	array, ok := stmt.Expression.(*ast.ArrayLiteral)
// 	if !ok {
// 		t.Fatalf("exp not ast.ArrayLiteral. got=%T", stmt.Expression)
// 	}

// 	if len(array.Elements) != 0 {
// 		t.Errorf("len(array.Elements) not 0. got=%d", len(array.Elements))
// 	}
// }

// func TestParsingArrayLiterals(t *testing.T) {
// 	input := "[1, 2 * 2, 3 + 3]"

// 	l := lexer.New(input)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
// 	array, ok := stmt.Expression.(*ast.ArrayLiteral)
// 	if !ok {
// 		t.Fatalf("exp not ast.ArrayLiteral. got=%T", stmt.Expression)
// 	}

// 	if len(array.Elements) != 3 {
// 		t.Fatalf("len(array.Elements) not 3. got=%d", len(array.Elements))
// 	}

// 	testIntegerLiteral(t, array.Elements[0], 1)
// 	testInfixExpression(t, array.Elements[1], 2, "*", 2)
// 	testInfixExpression(t, array.Elements[2], 3, "+", 3)
// }

// func TestParsingIndexExpressions(t *testing.T) {
// 	input := "myArray[1 + 1]"

// 	l := lexer.New(input)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
// 	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
// 	if !ok {
// 		t.Fatalf("exp not *ast.IndexExpression. got=%T", stmt.Expression)
// 	}

// 	if !testIdentifier(t, indexExp.Left, "myArray") {
// 		return
// 	}

// 	if !testInfixExpression(t, indexExp.Index, 1, "+", 1) {
// 		return
// 	}
// }

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {

	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral())
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s",
			value, bo.TokenLiteral())
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
