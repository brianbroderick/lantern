package pgLogParser

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

func TestScanner(t *testing.T) {
	var tests = []struct {
		s   string
		tok Token
		lit string
	}{
		{
			s:   "2023-07-10",
			tok: DATE,
			lit: "2023-07-10",
		},
		{
			s:   "09:52:46",
			tok: TIME,
			lit: "09:52:46",
		},
		{
			s:   "4321",
			tok: INTEGER,
			lit: "4321",
		},
		{
			s:   "14.42",
			tok: NUMBER,
			lit: "14.42",
		},
		{
			s:   "127.0.0.1",
			tok: IPADDR,
			lit: "127.0.0.1",
		},
		{
			s:   "0.42",
			tok: NUMBER,
			lit: "0.42",
		},
		{
			s:   "broderick",
			tok: IDENT,
			lit: "broderick",
		},
		{
			s:   "brian42",
			tok: IDENT,
			lit: "brian42",
		},
		{
			s:   "brian_42",
			tok: IDENT,
			lit: "brian_42",
		},
		{
			s:   "\"quotes\"",
			tok: STRING,
			lit: "quotes",
		},
		{
			s:   "\"quotes with \\\" in them.\"",
			tok: STRING,
			lit: "quotes with \" in them.",
		},
		{
			s:   "\"\\q\"", // an invalid escape surrounded by quotes
			tok: BADESCAPE,
			lit: "\\q",
		},
		{
			s:   "\"quotes without an ending quote",
			tok: BADSTRING,
			lit: "quotes without an ending quote",
		},
	}

	for _, tt := range tests {
		tok, _, lit := NewScanner(strings.NewReader(tt.s)).Scan()
		assert.Equal(t, Tokens[tt.tok], Tokens[tok])
		assert.Equal(t, tt.lit, lit)
	}
}
