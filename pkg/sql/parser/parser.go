package parser

import (
	"fmt"
	"strconv"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

const (
	_ int = iota
	STATEMENT
	UNION          // UNION
	LOWEST         // Most of the time, this is the lowest we go, except for when we're joining statements together such as with UNIONS
	USING          // USING in a JOIN or an index
	FILTER         // FILTER i.e. COUNT(*) FILTER (WHERE i < 5)
	AGGREGATE      // ORDER BY in a function call
	BETWEEN        // BETWEEN
	NOT            // NOT
	OR             // OR
	AND            // AND
	IS             // IS, IS NULL, IS NOT NULL, IS DISTINCT FROM, IS NOT DISTINCT FROM
	FROM           // FROM i.e. substring('foobar' from 1 for 3)
	EQUALS         // ==
	LESSGREATER    // > or <
	COMPARE        // IN, LIKE, ILIKE, SIMILAR, REGEX
	WINDOW         // OVER
	SUM            // +
	PRODUCT        // *
	AT_TIME_ZONE   // AT TIME ZONE. Needs to be higher than normal operators, but lower than function calls
	EXPONENTIATION // ^
	PREFIX         // -X or !X
	JSON           // ->, ->>, #>, #>>, @>, <@, ?, ?&, ?|
	CALL           // myFunction(X)
	INDEX          // array[index]
	ARRAYRANGE     // array[:index], array[index:], array[start:end]
	CAST
)

var precedences = map[token.TokenType]int{
	token.UNION:             UNION,
	token.INTERSECT:         UNION,
	token.EXCEPT:            UNION,
	token.USING:             USING,
	token.EQ:                EQUALS,
	token.NOT_EQ:            EQUALS,
	token.ASSIGN:            EQUALS,
	token.TO:                EQUALS,
	token.LT:                LESSGREATER,
	token.GT:                LESSGREATER,
	token.LTE:               LESSGREATER,
	token.GTE:               LESSGREATER,
	token.PLUS:              SUM,
	token.MINUS:             SUM,
	token.SLASH:             PRODUCT,
	token.ASTERISK:          PRODUCT,
	token.MODULO:            PRODUCT,
	token.LPAREN:            CALL,
	token.LBRACKET:          INDEX,
	token.COLON:             ARRAYRANGE,
	token.AND:               AND,
	token.OR:                OR,
	token.NOT:               NOT,
	token.IS:                IS,
	token.ISNULL:            IS,
	token.NOTNULL:           IS,
	token.FROM:              FROM,
	token.OVER:              WINDOW,
	token.BETWEEN:           BETWEEN,
	token.IN:                COMPARE,
	token.LIKE:              COMPARE,
	token.ILIKE:             COMPARE,
	token.SIMILAR:           COMPARE,
	token.OVERLAP:           COMPARE,
	token.REGEXMATCH:        COMPARE,
	token.REGEXIMATCH:       COMPARE,
	token.REGEXNOTMATCH:     COMPARE,
	token.REGEXNOTIMATCH:    COMPARE,
	token.EXPONENTIATION:    EXPONENTIATION,
	token.JSONGETBYKEY:      JSON,
	token.JSONGETBYTEXT:     JSON,
	token.JSONGETBYPATH:     JSON,
	token.JSONGETBYPATHTEXT: JSON,
	token.JSONCONTAINS:      JSON,
	token.JSONCONTAINED:     JSON,
	token.JSONHASKEY:        JSON,
	token.JSONHASALLKEYS:    JSON,
	token.JSONHASANYKEYS:    JSON,
	token.JSONDELETE:        JSON,
	token.JSONCONCAT:        JSON,
	token.ORDER:             AGGREGATE,
	token.FILTER:            FILTER,
	token.AT_TIME_ZONE:      AT_TIME_ZONE,
	token.TIMESTAMP:         AT_TIME_ZONE,
}

// https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-PRECEDENCE
// operators/elements, from highest to lowest precedence:
// . table/column name separator
// :: typecast
// [ ] array element selection
// + - unary plus, unary minus
// ^ exponentiation
// * / % multiplication, division, modulo
// + - addition, subtraction
// BETWEEN IN LIKE ILIKE SIMILAR range containment, set membership, string matching
// < > = <= >= <> comparison operators
// NOT logical negation
// AND logical conjunction
// OR logical disjunction
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
	parseContext  int
)

// These are different contexts that the parser can be in. This is because some
// expressions can be used in multiple contexts, but have different meanings.
const (
	XNIL parseContext = iota
	XCALL
	XLISTITEM
	XARRAY
	LITARRAY
	XGROUPED
	XDISTINCT
	XLOCK
	XNOT
	XIN
	XCREATE
	XUPDATE
	// XSELECTCOLUMN
)

// We may want to add the caller to the parser, to allow for context in conditions
// For example, an ORDER BY can show up in a select, but also in function calls
// For now, we're just passing the caller as a string in certain functions

type Parser struct {
	l            *lexer.Lexer
	errors       []string
	paramOffset  int
	traceLevel   int
	pos          lexer.Pos
	posPeek      lexer.Pos
	posPeekTwo   lexer.Pos
	posPeekThree lexer.Pos
	posPeekFour  lexer.Pos

	curToken       token.Token
	peekToken      token.Token
	peekTwoToken   token.Token
	peekThreeToken token.Token
	peekFourToken  token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	clause  token.TokenType // This is used to determine the clause that the current expression is in
	command token.TokenType // This is used to determine the command that the current expression is in
	context parseContext    // This is more specific than the clause. It's used to determine the context of the current expression
	not     bool            // This is used to determine if the NOT keyword is being used in a condition
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.ESCAPESTRING, p.parseEscapeStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.NULL, p.parseNull)
	p.registerPrefix(token.PARAM, p.parseParam)
	p.registerPrefix(token.UNKNOWN, p.parseUnknown)
	p.registerPrefix(token.TIMESTAMP, p.parseTimestampExpression)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.PLUS, p.parsePrefixExpression)
	p.registerPrefix(token.EXPONENTIATION, p.parsePrefixExpression)

	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.PARTITION, p.parseWindowExpression)
	p.registerPrefix(token.ORDER, p.parseWindowExpression)
	p.registerPrefix(token.SELECT, p.parseSelectExpression)
	p.registerPrefix(token.INSERT, p.parseInsertExpression)
	p.registerPrefix(token.UPDATE, p.parseUpdateExpression)
	p.registerPrefix(token.DELETE, p.parseDeleteExpression)
	p.registerPrefix(token.DISTINCT, p.parseDistinct)
	p.registerPrefix(token.ALL, p.parseDistinct)
	p.registerPrefix(token.CASE, p.parseCaseExpression)
	p.registerPrefix(token.LIKE, p.parseLikeExpression)
	p.registerPrefix(token.CAST, p.parseCastExpression)
	p.registerPrefix(token.INTERVAL, p.parseIntervalExpression)
	p.registerPrefix(token.WHERE, p.parseWhereExpression)
	p.registerPrefix(token.WITH, p.parseCTEExpression)
	p.registerPrefix(token.BOTH, p.parseTrimExpression)
	p.registerPrefix(token.LEADING, p.parseTrimExpression)
	p.registerPrefix(token.TRAILING, p.parseTrimExpression)
	p.registerPrefix(token.COLON, p.parsePrefixArrayRangeExpression)
	p.registerPrefix(token.VALUES, p.parseValuesExpression)
	p.registerPrefix(token.SEMICOLON, p.parseSemicolonExpression)
	p.registerPrefix(token.COMMIT, p.parseCommitExpression)
	p.registerPrefix(token.ROLLBACK, p.parseRollbackExpression)

	// Some tokens don't need special parse rules and can function as an identifier
	// If this becomes a problem, we can create a generic struct for these cases
	p.registerPrefix(token.LOCAL, p.parseIdentifier)
	p.registerPrefix(token.DEFAULT, p.parseIdentifier)
	p.registerPrefix(token.ANY, p.parseIdentifier)
	p.registerPrefix(token.USER, p.parseIdentifier)
	p.registerPrefix(token.SESSION_USER, p.parseIdentifier)
	p.registerPrefix(token.SYSTEM_USER, p.parseIdentifier)
	p.registerPrefix(token.CURRENT_CATALOG, p.parseIdentifier)
	p.registerPrefix(token.CURRENT_DATE, p.parseIdentifier)
	p.registerPrefix(token.CURRENT_SCHEMA, p.parseIdentifier)
	p.registerPrefix(token.CURRENT_ROLE, p.parseIdentifier)
	p.registerPrefix(token.CURRENT_TIME, p.parseIdentifier)
	p.registerPrefix(token.CURRENT_TIMESTAMP, p.parseIdentifier)
	p.registerPrefix(token.CURRENT_USER, p.parseIdentifier)
	p.registerPrefix(token.LEFT, p.parseIdentifier)
	p.registerPrefix(token.RIGHT, p.parseIdentifier)
	p.registerPrefix(token.ROW, p.parseIdentifier)
	p.registerPrefix(token.SET, p.parseIdentifier)
	p.registerPrefix(token.LAST, p.parseIdentifier)

	// This might be doing the same thing as parseIdentifier. TODO: check this out
	p.registerPrefix(token.ASTERISK, p.parseWildcardLiteral)
	p.registerPrefix(token.ALL, p.parseKeywordExpression)
	p.registerPrefix(token.NOT, p.parsePrefixKeywordExpression) // for things like NOT EXISTS (select...)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.ISNULL, p.parseInfixExpression)  // this might actually be a postfix operator
	p.registerInfix(token.NOTNULL, p.parseInfixExpression) // this might actually be a postfix operator
	p.registerInfix(token.LIKE, p.parseInfixExpression)
	p.registerInfix(token.ILIKE, p.parseInfixExpression)
	p.registerInfix(token.SIMILAR, p.parseSimilarToInfixExpression)
	p.registerInfix(token.BETWEEN, p.parseInfixExpression)
	p.registerInfix(token.REGEXMATCH, p.parseInfixExpression)
	p.registerInfix(token.REGEXIMATCH, p.parseInfixExpression)
	p.registerInfix(token.REGEXNOTMATCH, p.parseInfixExpression)
	p.registerInfix(token.REGEXNOTIMATCH, p.parseInfixExpression)
	p.registerInfix(token.OVER, p.parseInfixExpression)
	p.registerInfix(token.JSONGETBYKEY, p.parseInfixExpression)
	p.registerInfix(token.JSONGETBYTEXT, p.parseInfixExpression)
	p.registerInfix(token.JSONGETBYPATH, p.parseInfixExpression)
	p.registerInfix(token.JSONGETBYPATHTEXT, p.parseInfixExpression)
	p.registerInfix(token.JSONCONTAINS, p.parseInfixExpression)
	p.registerInfix(token.JSONCONTAINED, p.parseInfixExpression)
	p.registerInfix(token.JSONHASKEY, p.parseInfixExpression)
	p.registerInfix(token.JSONHASALLKEYS, p.parseInfixExpression)
	p.registerInfix(token.JSONHASANYKEYS, p.parseInfixExpression)
	p.registerInfix(token.JSONDELETE, p.parseInfixExpression)
	p.registerInfix(token.JSONCONCAT, p.parseInfixExpression)
	p.registerInfix(token.OVERLAP, p.parseInfixExpression)
	p.registerInfix(token.TO, p.parseInfixExpression)
	p.registerInfix(token.FILTER, p.parseInfixExpression)
	p.registerInfix(token.USING, p.parseInfixExpression)
	p.registerInfix(token.AT_TIME_ZONE, p.parseInfixExpression)
	p.registerInfix(token.UNION, p.parseUnionExpression)
	p.registerInfix(token.EXCEPT, p.parseUnionExpression)
	p.registerInfix(token.INTERSECT, p.parseUnionExpression)

	p.registerInfix(token.NOT, p.parseNotExpression)
	p.registerInfix(token.IS, p.parseIsExpression)

	p.registerInfix(token.IN, p.parseInExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseArrayExpression)
	p.registerInfix(token.COLON, p.parseArrayRangeExpression)
	p.registerInfix(token.ORDER, p.parseAggregateExpression)
	p.registerInfix(token.FROM, p.parseStringFunctionExpression)

	// Read three tokens, so curToken, peekToken, peekTwoToken, peekThreeToken, peekFourToken are all set
	p.nextToken()
	p.nextToken()
	p.nextToken()
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Input() string {
	return p.l.Input
}

// TODO: This is a hack to get around compound infix expressions
// We need to figure out a better way to do this, perhaps by implementing a real queue
// Perhaps something like: https://github.com/eapache/queue/blob/main/v2/queue.go
// Multi word keywords is getting annoying... we need to think of a better way to do this.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.pos = p.posPeek
	p.peekToken = p.peekTwoToken
	p.pos = p.posPeekTwo
	p.peekTwoToken = p.peekThreeToken
	p.posPeekTwo = p.posPeekThree
	p.peekThreeToken = p.peekFourToken
	p.posPeekThree = p.posPeekFour
	newToken, pos := p.advanceToken()
	p.peekFourToken = newToken
	p.posPeekFour = pos

	// Read into the future for AT TIME ZONE since AT isn't a reserved word (it's often used as a column or table alias)
	if p.peekTwoToken.Type == token.IDENT {

		switch p.peekTwoToken.Upper {
		case "AT":
			if p.peekThreeToken.Type == token.IDENT && p.peekThreeToken.Upper == "TIME" {
				if p.peekFourToken.Type == token.IDENT && p.peekFourToken.Upper == "ZONE" {
					p.peekTwoToken = token.Token{Type: token.AT_TIME_ZONE, Lit: "AT TIME ZONE"}
					newToken, pos = p.advanceToken()
					p.posPeekThree = pos
					p.peekThreeToken = newToken
					newToken, pos = p.advanceToken()
					p.peekFourToken = newToken
					p.posPeekFour = pos
				}
			}
		case "GROUP":
			if p.peekThreeToken.Type == token.BY {
				p.peekTwoToken = token.Token{Type: token.GROUP_BY, Lit: "GROUP BY"}
				p.peekThreeToken = p.peekFourToken
				p.posPeekThree = p.posPeekFour
				newToken, pos = p.advanceToken()
				p.peekFourToken = newToken
				p.posPeekFour = pos
			}
		}
	}

}

func (p *Parser) advanceToken() (newToken token.Token, pos lexer.Pos) {
	newToken, pos = p.l.Scan()
	// Skip comments
	iter := 0
	for newToken.Type == token.SQLCOMMENT {
		newToken, pos = p.l.Scan()
		iter++
		if iter > 50000 {
			return token.Token{Type: token.ILLEGAL, Lit: "Infinite loop in COMMENT block"}, pos
		}
	}
	// if newToken.Type == token.SQLCOMMENT {
	// 	newToken, pos = p.l.Scan()
	// }
	return newToken, pos
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// Uncomment this if you want to use it
// func (p *Parser) peekTwoTokenIs(t token.TokenType) bool {
// 	return p.peekTwoToken.Type == t
// }

// This is essentially an OR on curTokenIs
func (p *Parser) curTokenIsOne(tokens []token.TokenType) bool {
	for _, t := range tokens {
		if p.curTokenIs(t) {
			return true
		}
	}
	return false
}

// This is essentially !curTokenIsOne, but it can't match any token in the list
// Uncomment this if you want to use it
// func (p *Parser) curTokenIsNot(tokens []token.TokenType) bool {
// 	for _, t := range tokens {
// 		if p.curTokenIs(t) { // The token matches one of the tokens in the list
// 			return false
// 		}
// 	}

// 	return true
// }

func (p *Parser) peekTokenIsOne(tokens []token.TokenType) bool {
	for _, t := range tokens {
		if p.peekTokenIs(t) {
			return true
		}
	}
	return false
}

// Uncomment this if you want to use it
// func (p *Parser) peekTwoTokenIsOne(tokens []token.TokenType) bool {
// 	for _, t := range tokens {
// 		if p.peekTwoTokenIs(t) {
// 			return true
// 		}
// 	}
// 	return false
// }

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// Uncomment this if you want to use it
// func (p *Parser) expectPeekIsOne(tokens []token.TokenType) bool {
// 	found := false
// 	for _, t := range tokens {
// 		if p.peekTokenIs(t) {
// 			found = true
// 		}
// 	}
// 	if found {
// 		p.nextToken()
// 		return true
// 	} else {
// 		p.peekErrorIsOne(tokens)
// 		return false
// 	}
// }

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) PrintErrors() {
	if len(p.errors) == 0 {
		return
	}
	for _, msg := range p.errors {
		fmt.Printf("parser error: %s\n", msg)
	}
}

// Uncomment this if you want to use it
// func (p *Parser) peekErrorIsOne(tokens []token.TokenType) {
// 	toks := []string{}
// 	for _, t := range tokens {
// 		toks = append(toks, t.String())
// 	}

// 	msg := fmt.Sprintf("expected next token to be one of %s, got %s: %s instead. current token is: %s: %s",
// 		strings.Join(toks, ", "), p.peekToken.Type, p.peekToken.Lit, p.curToken.Type, p.curToken.Lit)
// 	p.errors = append(p.errors, msg)
// }

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s: %s instead. current token is: %s: %s",
		t, p.peekToken.Type, p.peekToken.Lit, p.curToken.Type, p.curToken.Lit)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found at line %d char %d", t, p.pos.Line, p.pos.Char)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	defer p.untrace(p.trace("ParseProgram"))

	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)
		p.paramOffset = 0 // reset this to zero so the next query will start at $1

		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	defer p.untrace(p.trace("parseStatement"))

	p.clause = p.curToken.Type
	p.command = p.curToken.Type

	switch p.curToken.Type {
	case token.SELECT:
		return p.parseSelectStatement()
	case token.WITH:
		return p.parseCTEStatement()
	case token.SET:
		return p.parseSetStatement()
	case token.DROP:
		return p.parseDropStatement()
	case token.CREATE:
		return p.parseCreateStatement()
	case token.ANALYZE:
		return p.parseAnalyzeStatement()
	case token.INSERT:
		return p.parseInsertStatement()
	case token.UPDATE:
		return p.parseUpdateStatement()
	case token.DELETE:
		return p.parseDeleteStatement()
	case token.SEMICOLON:
		return p.parseSemicolonStatement()
	case token.COMMIT:
		return p.parseCommitStatement()
	case token.ROLLBACK:
		return p.parseRollbackStatement()
	case token.IDENT:
		switch p.curToken.Upper {
		case "BEGIN":
			return p.parseBeginStatement()
		case "SHOW":
			return p.parseShowStatement()
		case "SAVEPOINT":
			return p.parseSavepointStatement()
		default:
			return p.parseExpressionStatement()
		}
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer p.untrace(p.trace("parseExpressionStatement"))

	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIsOne([]token.TokenType{token.SEMICOLON}) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer p.untrace(p.trace("parseExpression"))

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	leftExp = p.determineInfix(precedence, leftExp)

	// This is why all expressions must have a SetCast method
	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		leftExp.SetCast(p.parseDoubleColonExpression())
	}

	return leftExp
}

func (p *Parser) determineInfix(precedence int, leftExp ast.Expression) ast.Expression {
	var end []token.TokenType
	switch p.context {
	case XCALL: // Allow order by to denote an aggregate function
		end = []token.TokenType{token.COMMA}
	case XCREATE:
		end = []token.TokenType{token.INCLUDING, token.EXCLUDING}
	// case XSELECTCOLUMN:
	// 	end = []token.TokenType{token.FROM, token.WHERE, token.GROUP_BY, token.HAVING, token.ORDER, token.LIMIT, token.OFFSET, token.FETCH, token.FOR, token.SEMICOLON, token.EOF}
	case XUPDATE:
		end = []token.TokenType{token.SET, token.FROM, token.WHERE, token.RETURNING, token.SEMICOLON, token.EOF}
	default:
		end = []token.TokenType{token.COMMA, token.WHERE, token.GROUP_BY, token.HAVING, token.ORDER, token.LIMIT, token.OFFSET, token.FETCH, token.FOR, token.SEMICOLON}
	}

	for !p.peekTokenIsOne(end) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]

		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseColumnSection() ast.Expression {
	defer p.untrace(p.trace("parseColumnSection"))

	if p.peekTokenIs(token.DOT) {
		p.nextToken()
		p.nextToken()
		if p.curTokenIs(token.ASTERISK) {
			return &ast.WildcardLiteral{Token: p.curToken, Value: p.curToken.Lit, Branch: p.clause, CommandTag: p.command}
		} else if p.curTokenIs(token.IDENT) {
			return &ast.SimpleIdentifier{Token: p.curToken, Value: p.curToken.Lit, Branch: p.clause, CommandTag: p.command}
		}
	}
	return nil
}

func (p *Parser) parseIdentifier() ast.Expression {
	defer p.untrace(p.trace("parseIdentifier"))

	x := &ast.Identifier{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
	x.Value = append(x.Value, &ast.SimpleIdentifier{Token: p.curToken, Value: p.curToken.Lit, Branch: p.clause, CommandTag: p.command})

	for p.peekTokenIs(token.DOT) {
		x.Value = append(x.Value, p.parseColumnSection())
	}

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		x.SetCast(p.parseDoubleColonExpression())
	}
	return x
}

func (p *Parser) parseWildcardLiteral() ast.Expression {
	defer p.untrace(p.trace("parseWildcardLiteral"))

	return &ast.WildcardLiteral{Token: p.curToken, Value: p.curToken.Lit, Branch: p.clause, CommandTag: p.command}
}

func (p *Parser) parseKeywordExpression() ast.Expression {
	defer p.untrace(p.trace("parseKeywordExpression"))

	return &ast.KeywordExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer p.untrace(p.trace("parseIntegerLiteral"))

	// Incrementing the offset is to help when masking parameters in the AST
	p.paramOffset++
	lit := &ast.IntegerLiteral{Token: p.curToken, ParamOffset: p.paramOffset, Branch: p.clause, CommandTag: p.command}

	// TODO: sometimes numbers are larger than int64
	// value := new(big.Int)
	// value.SetString(p.curToken.Lit, 10)

	value, err := strconv.ParseInt(p.curToken.Lit, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Lit)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		lit.SetCast(p.parseDoubleColonExpression())
	}

	return lit
}

func (p *Parser) parseParam() ast.Expression {
	defer p.untrace(p.trace("parseParam"))

	p.paramOffset++
	return &ast.ParamLiteral{Token: p.curToken, ParamOffset: p.paramOffset, Branch: p.clause, CommandTag: p.command}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	defer p.untrace(p.trace("parseFloatLiteral"))

	// Incrementing the offset is to help when masking parameters in the AST
	p.paramOffset++
	lit := &ast.FloatLiteral{Token: p.curToken, ParamOffset: p.paramOffset, Branch: p.clause, CommandTag: p.command}

	value, err := strconv.ParseFloat(p.curToken.Lit, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Lit)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		lit.SetCast(p.parseDoubleColonExpression())
	}

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	defer p.untrace(p.trace("parseStringLiteral"))

	p.paramOffset++
	str := &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Lit, ParamOffset: p.paramOffset, Branch: p.clause, CommandTag: p.command}
	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		str.SetCast(p.parseDoubleColonExpression())
	}
	return str
}

func (p *Parser) parseEscapeStringLiteral() ast.Expression {
	defer p.untrace(p.trace("parseEscapeStringLiteral"))

	p.paramOffset++
	str := &ast.EscapeStringLiteral{Token: p.curToken, Value: p.curToken.Lit, ParamOffset: p.paramOffset, Branch: p.clause, CommandTag: p.command}
	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		str.SetCast(p.parseDoubleColonExpression())
	}
	return str
}

func (p *Parser) parseSemicolonStatement() *ast.SemicolonStatement {
	defer p.untrace(p.trace("parseSemicolonStatement"))

	s := &ast.SemicolonStatement{Token: p.curToken}
	s.Expression = p.parseExpression(STATEMENT)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return s
}

func (p *Parser) parseSemicolonExpression() ast.Expression {
	defer p.untrace(p.trace("parseSemicolonExpression"))

	return &ast.SemicolonExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	defer p.untrace(p.trace("parsePrefixExpression"))

	expression := &ast.PrefixExpression{
		Token:      p.curToken,
		Operator:   p.curToken.Lit,
		Branch:     p.clause,
		CommandTag: p.command,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parsePrefixKeywordExpression() ast.Expression {
	defer p.untrace(p.trace("parsePrefixKeywordExpression"))

	expression := &ast.PrefixKeywordExpression{
		Token:      p.curToken,
		Operator:   p.curToken.Lit,
		Branch:     p.clause,
		CommandTag: p.command,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseInfixExpression"))

	x := &ast.InfixExpression{
		Token:      p.curToken,
		Operator:   p.curToken.Lit,
		Left:       left,
		Branch:     p.clause,
		CommandTag: p.command,
	}

	if p.not {
		x.Not = true
		p.not = false
	}

	precedence := p.curPrecedence()
	p.nextToken()

	x.Right = p.parseExpression(precedence)

	return x
}

func (p *Parser) parseSimilarToInfixExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseInfixExpression"))

	precedence := p.curPrecedence()

	x := &ast.InfixExpression{
		Token:      p.curToken,
		Operator:   p.curToken.Lit,
		Left:       left,
		Branch:     p.clause,
		CommandTag: p.command,
	}

	if p.not {
		x.Not = true
		p.not = false
	}

	if p.peekTokenIs(token.TO) {
		x.Operator = "SIMILAR TO"
		p.nextToken()
	}

	p.nextToken()

	x.Right = p.parseExpression(precedence)

	return x
}

func (p *Parser) parseUnionExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseUnionExpression"))

	expression := &ast.UnionExpression{
		Token:      p.curToken,
		Operator:   p.curToken.Lit,
		Left:       left,
		Branch:     p.clause,
		CommandTag: p.command,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	if p.curTokenIs(token.ALL) {
		expression.All = true
		p.nextToken()
	}

	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	defer p.untrace(p.trace("parseBoolean"))

	p.paramOffset++

	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE), ParamOffset: p.paramOffset, Branch: p.clause, CommandTag: p.command}
}

func (p *Parser) parseNull() ast.Expression {
	defer p.untrace(p.trace("parseNull"))

	p.paramOffset++
	x := &ast.Null{Token: p.curToken, ParamOffset: p.paramOffset, Branch: p.clause, CommandTag: p.command}

	// This is why all expressions must have a SetCast method
	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		x.SetCast(p.parseDoubleColonExpression())
	}
	return x
}

func (p *Parser) parseUnknown() ast.Expression {
	defer p.untrace(p.trace("parseUnknown"))

	p.paramOffset++
	x := &ast.Unknown{Token: p.curToken, ParamOffset: p.paramOffset, Branch: p.clause, CommandTag: p.command}

	// This is why all expressions must have a SetCast method
	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		x.SetCast(p.parseDoubleColonExpression())
	}
	return x
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	defer p.untrace(p.trace("parseGroupedExpression"))
	context := p.context
	p.setContext(XGROUPED)        // sets the context for the parseExpressionListItem function
	defer p.resetContext(context) // reset to prior context

	p.nextToken()

	x := &ast.GroupedExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
	x.Elements = p.parseExpressionList([]token.TokenType{token.RPAREN})

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		x.SetCast(p.parseDoubleColonExpression())
	}

	return x
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseCallExpression"))

	// When it's a function, set the clause to FUNCTION_CALL so on evaluation, we know if an IDENT is a function, column, etc.
	function.SetClause(token.FUNCTION_CALL)

	x := &ast.CallExpression{Token: p.curToken, Function: function, Branch: p.clause, CommandTag: p.command}

	p.nextToken()

	// DISTINCT CLAUSE
	if p.curTokenIsOne([]token.TokenType{token.ALL, token.DISTINCT}) {
		x.Distinct = p.parseDistinct()
	}
	context := p.context
	p.setContext(XCALL)           // sets the context for the parseExpressionListItem function
	defer p.resetContext(context) // reset to prior context

	x.Arguments = p.parseExpressionList([]token.TokenType{token.RPAREN})

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		x.SetCast(p.parseDoubleColonExpression())
	}

	return x
}

func (p *Parser) parseExpressionList(end []token.TokenType) []ast.Expression {
	defer p.untrace(p.trace("parseExpressionList"))

	list := []ast.Expression{}

	if p.curTokenIsOne(end) {
		return list
	}

	list = append(list, p.parseExpression(STATEMENT))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(STATEMENT))
	}

	if !p.peekTokenIsOne(end) {
		return nil
	}

	p.nextToken()

	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	defer p.untrace(p.trace("parseArrayLiteral"))

	context := p.context
	p.setContext(LITARRAY)        // sets the context for the parseExpressionListItem function
	defer p.resetContext(context) // reset to prior context

	array := &ast.ArrayLiteral{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
	p.nextToken()
	array.Elements = p.parseExpressionList([]token.TokenType{token.RBRACKET})

	return array
}

func (p *Parser) parseArrayExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseArrayExpression"))

	context := p.context
	p.setContext(XARRAY)          // sets the context for the parseExpressionListItem function
	defer p.resetContext(context) // reset to prior context

	array := &ast.ArrayLiteral{Token: p.curToken, Left: left, Branch: p.clause, CommandTag: p.command}
	p.nextToken()
	array.Elements = p.parseExpressionList([]token.TokenType{token.RBRACKET})

	if p.peekTokenIs(token.DOUBLECOLON) {
		p.nextToken()
		p.nextToken()
		array.SetCast(p.parseDoubleColonExpression())
	}

	return array
}

func (p *Parser) parsePrefixArrayRangeExpression() ast.Expression {
	defer p.untrace(p.trace("parseArrayRangeExpression"))

	expression := &ast.InfixExpression{
		Token:      p.curToken,
		Operator:   p.curToken.Lit,
		Left:       &ast.Infinity{Token: p.curToken, Branch: p.clause, CommandTag: p.command},
		Branch:     p.clause,
		CommandTag: p.command,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseArrayRangeExpression(left ast.Expression) ast.Expression {
	defer p.untrace(p.trace("parseArrayRangeExpression"))

	expression := &ast.InfixExpression{
		Token:      p.curToken,
		Operator:   p.curToken.Lit,
		Left:       left,
		Branch:     p.clause,
		CommandTag: p.command,
	}

	if p.peekTokenIs(token.RBRACKET) {
		expression.Right = &ast.Infinity{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
	} else {
		precedence := p.curPrecedence()
		p.nextToken()
		expression.Right = p.parseExpression(precedence)
	}

	return expression
}

func (p *Parser) parseIntervalExpression() ast.Expression {
	defer p.untrace(p.trace("parseIntervalExpression"))

	interval := &ast.IntervalExpression{Token: p.curToken, Branch: p.clause, CommandTag: p.command}
	p.nextToken()
	interval.Value = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Lit, Branch: p.clause, CommandTag: p.command}

	intervalOption := map[string]bool{"YEAR": true, "MONTH": true, "DAY": true, "HOUR": true, "MINUTE": true, "SECOND": true}

	if p.peekTokenIs(token.IDENT) {
		if _, ok := intervalOption[p.peekToken.Upper]; ok {
			p.nextToken()
			opt := p.curToken.Upper
			if p.peekTokenIs(token.TO) {
				p.nextToken()
				p.nextToken()
				opt += " TO " + p.curToken.Upper
			}

			interval.Unit = &ast.SimpleIdentifier{Token: p.curToken, Value: opt, Branch: p.clause, CommandTag: p.command}
		}
	}

	return interval
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// This allows the parser to know what context it's currently in.
// These two functions fire before and after the parseExpression function via a defer
func (p *Parser) resetContext(context parseContext) {
	p.context = context
}

func (p *Parser) setContext(context parseContext) parseContext {
	p.context = context
	return p.context
}
