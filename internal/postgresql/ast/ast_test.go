package ast

// func TestString(t *testing.T) {
// 	program := &Program{
// 		Statements: []Statement{
// 			&LogStatement{
// 				Token: token.Token{Type: token.Date, Lit: "date"},
// 				Name: &Identifier{
// 					Token: token.Token{Type: token.IDENT, Lit: "myVar"},
// 					Value: "myVar",
// 				},
// 				Value: &Identifier{
// 					Token: token.Token{Type: token.IDENT, Lit: "anotherVar"},
// 					Value: "anotherVar",
// 				},
// 			},
// 		},
// 	}

// 	if program.String() != "let myVar = anotherVar;" {
// 		t.Errorf("program.String() wrong. got=%q", program.String())
// 	}
// }
