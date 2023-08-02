package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/internal/postgresql/lexer"
	"github.com/stretchr/testify/assert"
)

// func TestParser(t *testing.T) {
// 	s := "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo where bar = $1"
// 	l := lexer.New(s)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	assert.Equal(t, 1, len(program.Statements))
// 	assert.Equal(t, "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo where bar = $1", program.Statements[0].String())
// }

func TestParserStatements(t *testing.T) {
	var tests = []struct {
		str           string
		result        string
		lenStatements int
	}{
		// Single line log entry
		{
			str:           "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from users where id = $1",
			result:        "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from users where id = $1",
			lenStatements: 1,
		},
		// Multiple line log entry
		{
			str:           "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo\n where bar = $1",
			result:        "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo where bar = $1",
			lenStatements: 1,
		},
		// Multiple log entries
		{
			str:           "2023-01-02 01:02:03 MDT:127.0.0.1(12345):postgres@sampledb:[23456]:LOG:  duration: 0.123 ms  execute <unnamed>: select * from foo where bar = $1\n2023-07-10 10:11:12 MDT:127.0.0.1(56789):postgres@sampledb:[98765]:LOG:  duration: 0.159 ms  execute <unnamed>: select * from bar where baz = $1",
			result:        "2023-01-02 01:02:03 MDT:127.0.0.1(12345):postgres@sampledb:[23456]:LOG:  duration: 0.123 ms  execute <unnamed>: select * from foo where bar = $1",
			lenStatements: 2,
		},
		// Multiple log entries on multiple lines
		{
			str:           "2023-01-02 01:02:03 MDT:127.0.0.1(12345):postgres@sampledb:[23456]:LOG:  duration: 0.123 ms  execute <unnamed>: select * \n  from multi\n  where \nline = $1\n2023-07-10 10:11:12 MDT:127.0.0.1(56789):postgres@sampledb:[98765]:LOG:  duration: 0.159 ms  execute <unnamed>: select * from\n second where id = $1",
			result:        "2023-01-02 01:02:03 MDT:127.0.0.1(12345):postgres@sampledb:[23456]:LOG:  duration: 0.123 ms  execute <unnamed>: select * from multi where line = $1",
			lenStatements: 2,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.str)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		assert.Equal(t, tt.lenStatements, len(program.Statements))
		assert.Equal(t, tt.result, program.Statements[0].String())
	}

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
