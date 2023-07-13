package pgLogParser

import "bytes"

//*******
// LOG
//*******

// LogStatement represents a log entry
type LogStatement struct{}

func (s *LogStatement) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("LOG")

	return buf.String()
}

func (s *LogStatement) KeyTok() Token {
	return TIMESTAMP
}

// parseLogStatement parses a log entry a Statement AST object.
// This function assumes the TIMESTAMP token has already been consumed.
func (p *Parser) parseLogStatement() (*LogStatement, error) {
	return &LogStatement{}, nil
}
