package pgLogParser

// type Log struct {
// 	Statements []Statement
// }

// func (l *Log) ParseLog() *Log {
// 	program := &Log{}
// 	program.Statements = []Statement{}

// 	for !l.curTokenIs(EOF) {
// 		stmt := l.parseStatement()
// 		if stmt != nil {
// 			program.Statements = append(program.Statements, stmt)
// 		}
// 		l.nextToken()
// 	}

// 	return program
// }

// func (l *Log) parseStatement() Statement {
// 	switch p.curToken.Type {
// 	case LET:
// 		return p.parseLetStatement()
// 	case RETURN:
// 		return p.parseReturnStatement()
// 	default:
// 		return p.parseExpressionStatement()
// 	}
// }
