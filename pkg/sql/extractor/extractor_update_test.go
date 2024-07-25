package extractor

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/stretchr/testify/assert"
)

func TestExtractUpdateStatements(t *testing.T) {
	t1 := time.Now()

	tests := []struct {
		input  string
		tables []string
	}{
		{"update users set name = 'Brian';",
			[]string{"public.users"}},
		{"update films set kind = 'Dramatic' where kind = 'Drama';",
			[]string{"public.films"}},
		{"update weather set temp_lo = temp_lo+1, temp_hi = temp_lo+15, prcp = default where city = 'San Francisco' and date = '2003-07-03';",
			[]string{"public.weather"}},
		{"update weather set temp_lo = temp_lo+1, temp_hi = temp_lo+15, prcp = default where city = 'San Francisco' and date = '2003-07-03' returning temp_lo, temp_hi, prcp;",
			[]string{"public.weather"}},
		{"update weather set (temp_lo, temp_hi, prcp) = (temp_lo+1, temp_lo+15, default) where city = 'San Francisco' and date = '2003-07-03';",
			[]string{"public.weather"}},
		{"UPDATE films SET kind = 'Dramatic' WHERE CURRENT OF c_films;",
			[]string{"public.films"}},
		{"update employees set sales_count = sales_count + 1 from accounts where accounts.name = 'Acme Corporation' and employees.id = accounts.sales_person;",
			[]string{"public.employees", "public.accounts"}},
		{"update employees set sales_count = sales_count + 1 where id =	(select sales_person from accounts where name = 'Acme Corporation');",
			[]string{"public.employees", "public.accounts"}},
		{"update accounts SET (contact_first_name, contact_last_name) = (select first_name, last_name from employees where employees.id = accounts.sales_person);",
			[]string{"public.accounts", "public.employees"}},
		{"update accounts set contact_first_name = first_name, contact_last_name = last_name from employees where employees.id = accounts.sales_person;",
			[]string{"public.accounts", "public.employees"}},
		{"UPDATE summary s SET (sum_x, sum_y, avg_x, avg_y) = (SELECT sum(x), sum(y), avg(x), avg(y) FROM data d WHERE d.group_id = s.group_id);",
			[]string{"public.summary", "public.data"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		for _, s := range program.Statements {
			r := NewExtractor(&s, true)
			r.Execute(s)
			checkExtractErrors(t, r, tt.input)

			assert.Equal(t, len(tt.tables), len(r.TablesInQueries), "input: %s\nNumber of tables not equal", tt.input)

			for _, table := range r.TablesInQueries {
				fqtn := fmt.Sprintf("%s.%s", table.Schema, table.Name)
				assert.Contains(t, tt.tables, fqtn, "input: %s\nTable %s not found in %v", tt.input, table.Name, fqtn)
			}
		}
	}

	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	fmt.Printf("TestExtractUpdateStatements, Elapsed Time: %s\n", timeDiff)
}

// func TestExtractUpdateCommandTags(t *testing.T) {
// 	t1 := time.Now()

// 	tests := []struct {
// 		input  string
// 		tables [][]string
// 	}{
// 		{"update users set name = 'Brian';",
// 			[][]string{{"public.users", "UPDATE"}}},
// 		{"update films set kind = 'Dramatic' where kind = 'Drama';",
// 			[][]string{{"public.films", "UPDATE"}}},
// 		{"update weather set temp_lo = temp_lo+1, temp_hi = temp_lo+15, prcp = default where city = 'San Francisco' and date = '2003-07-03';",
// 			[][]string{{"public.weather", "UPDATE"}}},
// 		{"update weather set temp_lo = temp_lo+1, temp_hi = temp_lo+15, prcp = default where city = 'San Francisco' and date = '2003-07-03' returning temp_lo, temp_hi, prcp;",
// 			[][]string{{"public.weather", "UPDATE"}}},
// 		{"update weather set (temp_lo, temp_hi, prcp) = (temp_lo+1, temp_lo+15, default) where city = 'San Francisco' and date = '2003-07-03';",
// 			[][]string{{"public.weather", "UPDATE"}}},
// 		{"UPDATE films SET kind = 'Dramatic' WHERE CURRENT OF c_films;",
// 			[][]string{{"public.films", "UPDATE"}}},
// 		{"update employees set sales_count = sales_count + 1 from accounts where accounts.name = 'Acme Corporation' and employees.id = accounts.sales_person;",
// 			[][]string{{"public.employees", "UPDATE"}, {"public.accounts", "SELECT"}}},
// 		{"update employees set sales_count = sales_count + 1 where id =	(select sales_person from accounts where name = 'Acme Corporation');",
// 			[][]string{{"public.employees", "UPDATE"}, {"public.accounts", "SELECT"}}},
// 		{"update accounts SET (contact_first_name, contact_last_name) = (select first_name, last_name from employees where employees.id = accounts.sales_person);",
// 			[][]string{{"public.accounts", "UPDATE"}, {"public.employees", "SELECT"}}},
// 		{"update accounts set contact_first_name = first_name, contact_last_name = last_name from employees where employees.id = accounts.sales_person;",
// 			[][]string{{"public.accounts", "UPDATE"}, {"public.employees", "SELECT"}}},
// 		{"UPDATE summary s SET (sum_x, sum_y, avg_x, avg_y) = (SELECT sum(x), sum(y), avg(x), avg(y) FROM data d WHERE d.group_id = s.group_id);",
// 			[][]string{{"public.summary", "UPDATE"}, {"public.data", "SELECT"}}},
// 	}

// 	for _, tt := range tests {
// 		l := lexer.New(tt.input)
// 		p := parser.New(l)
// 		program := p.ParseProgram()

// 		for _, s := range program.Statements {
// 			r := NewExtractor(&s, true)
// 			r.Execute(s)
// 			checkExtractErrors(t, r, tt.input)

// 			assert.Equal(t, len(tt.tables), len(r.TablesInQueries), "input: %s\nNumber of tables not equal", tt.input)

// 			for _, table := range r.TablesInQueries {
// 				fqtn := fmt.Sprintf("%s.%s", table.Schema, table.Name)
// 				for _, ss := range tt.tables {
// 					for _, testTbl := range ss {
// 						if testTbl == fqtn {
// 							assert.Equal(t, table.Command.String(), ss[1], "input: %s\nCommand not equal", tt.input)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}

// 	t2 := time.Now()
// 	timeDiff := t2.Sub(t1)
// 	fmt.Printf("TestExtractUpdateCommandTags, Elapsed Time: %s\n", timeDiff)
// }
