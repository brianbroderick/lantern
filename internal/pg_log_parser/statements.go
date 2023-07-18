package pgLogParser

import (
	"bytes"
	"time"
)

//*******
// LOG
//*******

// LogStatement represents a log entry
type LogStatement struct {
	eventDate    time.Time
	remoteHost   string
	remotePort   int
	userName     string
	databaseName string
	processID    int
	duration     time.Duration
	preparedStep string
	preparedName string
	statement    string
}

func (s *LogStatement) String() string {
	var buf bytes.Buffer
	_, _ = buf.WriteString("LOG")

	return buf.String()
}

func (s *LogStatement) KeyTok() Token {
	return TIMESTAMP
}

// parseLogStatement parses a log entry a Statement AST object.
func (p *Parser) parseLogStatement() (*LogStatement, error) {
	p.Unscan()
	stmt := &LogStatement{}

	return stmt, nil
}
