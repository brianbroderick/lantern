package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestUpdateStatements(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		{"update users set name = 'Brian';", "(UPDATE users SET (name = 'Brian'));"},
		{"update films set kind = 'Dramatic' where kind = 'Drama';", "(UPDATE films SET (kind = 'Dramatic') WHERE (kind = 'Drama'));"},
		{"update weather set temp_lo = temp_lo+1, temp_hi = temp_lo+15, prcp = default where city = 'San Francisco' and date = '2003-07-03';", "(UPDATE weather SET (temp_lo = (temp_lo + 1)), (temp_hi = (temp_lo + 15)), (prcp = default) WHERE ((city = 'San Francisco') AND (date = '2003-07-03')));"},
		{"update weather set temp_lo = temp_lo+1, temp_hi = temp_lo+15, prcp = default where city = 'San Francisco' and date = '2003-07-03' returning temp_lo, temp_hi, prcp;", "(UPDATE weather SET (temp_lo = (temp_lo + 1)), (temp_hi = (temp_lo + 15)), (prcp = default) WHERE ((city = 'San Francisco') AND (date = '2003-07-03')) RETURNING temp_lo, temp_hi, prcp);"},
		{"update weather set (temp_lo, temp_hi, prcp) = (temp_lo+1, temp_lo+15, default) where city = 'San Francisco' and date = '2003-07-03';", "(UPDATE weather SET ((temp_lo, temp_hi, prcp) = ((temp_lo + 1), (temp_lo + 15), default)) WHERE ((city = 'San Francisco') AND (date = '2003-07-03')));"},
		{"update employees set sales_count = sales_count + 1 from accounts where accounts.name = 'Acme Corporation' and employees.id = accounts.sales_person;", "(UPDATE employees SET (sales_count = (sales_count + 1)) FROM accounts WHERE ((accounts.name = 'Acme Corporation') AND (employees.id = accounts.sales_person)));"},
		{"update employees set sales_count = sales_count + 1 where id =	(select sales_person from accounts where name = 'Acme Corporation');", "(UPDATE employees SET (sales_count = (sales_count + 1)) WHERE (id = (SELECT sales_person FROM accounts WHERE (name = 'Acme Corporation'))));"},
		{"update accounts SET (contact_first_name, contact_last_name) = (select first_name, last_name from employees where employees.id = accounts.sales_person);", "(UPDATE accounts SET ((contact_first_name, contact_last_name) = (SELECT first_name, last_name FROM employees WHERE (employees.id = accounts.sales_person))));"},
		{"update accounts set contact_first_name = first_name, contact_last_name = last_name from employees where employees.id = accounts.sales_person;", "(UPDATE accounts SET (contact_first_name = first_name), (contact_last_name = last_name) FROM employees WHERE (employees.id = accounts.sales_person));"},
		{"UPDATE summary s SET (sum_x, sum_y, avg_x, avg_y) = (SELECT sum(x), sum(y), avg(x), avg(y) FROM data d WHERE d.group_id = s.group_id);",
			"(UPDATE summary s SET ((sum_x, sum_y, avg_x, avg_y) = (SELECT sum(x), sum(y), avg(x), avg(y) FROM data d WHERE (d.group_id = s.group_id))));"},
		{"update customers c set name = 'new name' from suppliers where c.id = suppliers.id;", "(UPDATE customers c SET (name = 'new name') FROM suppliers WHERE (c.id = suppliers.id));"},
		{"update customers c set name = 'new name' from suppliers s where c.id = s.id;", "(UPDATE customers c SET (name = 'new name') FROM suppliers s WHERE (c.id = s.id));"},
		{"UPDATE films SET kind = 'Dramatic' WHERE CURRENT OF c_films;", "(UPDATE films SET (kind = 'Dramatic') WHERE CURRENT OF c_films);"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		stmt := program.Statements[0]
		assert.Equal(t, "UPDATE", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.UpdateStatement. got=%T", tt.input, stmt)

		_, ok := stmt.(*ast.UpdateStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.UpdateStatement. got=%T", tt.input, stmt)

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
