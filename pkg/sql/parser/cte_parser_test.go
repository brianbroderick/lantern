package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/stretchr/testify/assert"
)

func TestCTEs(t *testing.T) {
	maskParams := false

	tests := []struct {
		input  string
		output string
	}{
		// Select: CTEs
		{"with sales as (select sum(amount) as total_sales from orders) select total_sales from sales;", "(WITH sales AS (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"},
		{"with sales as (select sum(amount) as total_sales from orders) select total_sales from sales", "(WITH sales AS (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"}, // no semi-colon
		{"with regional_sales as (select region, sum(amount) as total_sales from orders group by region), top_regions as (select region from regional_sales where total_sales > 42)	select region, product, sum(quantity) AS product_units, sum(amount) as product_sales from orders group by region, product;",
			"(WITH regional_sales AS (SELECT region, sum(amount) AS total_sales FROM orders GROUP BY region), top_regions AS (SELECT region FROM regional_sales WHERE (total_sales > 42)) (SELECT region, product, sum(quantity) AS product_units, sum(amount) AS product_sales FROM orders GROUP BY region, product));"},
		{"with regional_sales as (select region, sum(amount) as total_sales from orders group by region), top_regions AS (select region from regional_sales where total_sales > (select sum(total_sales)/10 from regional_sales))	select region, product, sum(quantity) AS product_units, sum(amount) as product_sales from orders group by region, product;",
			"(WITH regional_sales AS (SELECT region, sum(amount) AS total_sales FROM orders GROUP BY region), top_regions AS (SELECT region FROM regional_sales WHERE (total_sales > (SELECT (sum(total_sales) / 10) FROM regional_sales))) (SELECT region, product, sum(quantity) AS product_units, sum(amount) AS product_sales FROM orders GROUP BY region, product));"},
		{"with regional_sales as (select region, sum(amount) as total_sales from orders group by region), top_regions AS (select region from regional_sales where total_sales > (select sum(total_sales)/10 from regional_sales))	select region, product, sum(quantity) AS product_units, sum(amount) as product_sales from orders where region in (SELECT region from top_regions) group by region, product;",
			"(WITH regional_sales AS (SELECT region, sum(amount) AS total_sales FROM orders GROUP BY region), top_regions AS (SELECT region FROM regional_sales WHERE (total_sales > (SELECT (sum(total_sales) / 10) FROM regional_sales))) (SELECT region, product, sum(quantity) AS product_units, sum(amount) AS product_sales FROM orders WHERE region IN ((SELECT region FROM top_regions)) GROUP BY region, product));"},
		{"with recursive sales as (select sum(amount) as total_sales from orders) select total_sales from sales;", "(WITH RECURSIVE sales AS (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"},
		{"with sales as materialized (select sum(amount) as total_sales from orders) select total_sales from sales;", "(WITH sales AS MATERIALIZED (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"},
		{"with sales as not materialized (select sum(amount) as total_sales from orders) select total_sales from sales;", "(WITH sales AS NOT MATERIALIZED (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales));"},
		{"with sales as (select sum(amount) as total_sales from orders) select total_sales from sales union select total_sales from sales;", "(WITH sales AS (SELECT sum(amount) AS total_sales FROM orders) (SELECT total_sales FROM sales UNION (SELECT total_sales FROM sales)));"},
		{"with recursive parts(part) as (select part from parts) select count(part) from parts;", "(WITH RECURSIVE parts(part) AS (SELECT part FROM parts) (SELECT count(part) FROM parts));"},
		{"with my_list as (select i from generate_series(1,10) s(i)) select i from my_list where i > 5;", "(WITH my_list AS (SELECT i FROM generate_series(1, 10) s(i)) (SELECT i FROM my_list WHERE (i > 5)));"},
	}

	for _, tt := range tests {
		// fmt.Printf("\ninput:  %s\n", tt.input)
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p, tt.input)

		assert.Equal(t, 1, len(program.Statements), "input: %s\nprogram.Statements does not contain %d statements. got=%d\n", tt.input, 1, len(program.Statements))

		stmt := program.Statements[0]
		assert.Equal(t, "with", stmt.TokenLiteral(), "input: %s\nprogram.Statements[0] is not ast.CTEStatement. got=%T", tt.input, stmt)

		cteStmt, ok := stmt.(*ast.CTEStatement)
		assert.True(t, ok, "input: %s\nstmt is not *ast.CTEStatement. got=%T", tt.input, stmt)

		cteExpr, ok := cteStmt.Expression.(*ast.CTEExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.CTEExpression. got=%T", tt.input, cteExpr)

		cteExp, ok := cteExpr.Expressions[0].(*ast.SelectExpression)
		assert.True(t, ok, "input: %s\nstmt is not *ast.SelectExpression. got=%T", tt.input, cteExp)

		// program.Statements[0].Inspect()

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)
		// fmt.Printf("output: %s\n", output)
	}
}
