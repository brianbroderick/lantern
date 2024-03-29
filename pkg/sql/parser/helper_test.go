package parser

import (
	"fmt"
	"testing"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
)

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

	simple, ok := ident.Value[0].(*ast.SimpleIdentifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if simple.Value != value {
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

// func testSimpleIdentifier(t *testing.T, exp ast.Expression, value string) bool {
// 	ident, ok := exp.(*ast.SimpleIdentifier)
// 	if !ok {
// 		t.Errorf("exp not *ast.SimpleIdentifier. got=%T", exp)
// 		return false
// 	}

// 	if ident.Value != value {
// 		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
// 		return false
// 	}

// 	if ident.TokenLiteral() != value {
// 		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
// 			ident.TokenLiteral())
// 		return false
// 	}

// 	return true
// }

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

func checkParserErrors(t *testing.T, p *Parser, input string) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("input: %s\nparser has %d errors", input, len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

// ##############################################################
// Commenting this out as it's now obsolete in favor of PG's array syntax
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
