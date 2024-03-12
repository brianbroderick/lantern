package repo

import (
	"fmt"
	"testing"
	"time"

	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/stretchr/testify/assert"
)

func TestShaQuery(t *testing.T) {
	maskParams := true
	t1 := time.Now()

	tests := []struct {
		input  string
		output string
		sha    string
	}{
		{"select * from users where id = 42", "(SELECT * FROM users WHERE (id = ?));", "b90be791fd20b7fc4925c087456e73493496364d"},
		{"select * from users where id = 74", "(SELECT * FROM users WHERE (id = ?));", "b90be791fd20b7fc4925c087456e73493496364d"},
		{"DROP TABLE IF EXISTS listing;", "(DROP TABLE IF EXISTS listing);", "da66725b4ceb14a317abd3ccb5beec320c482f5d"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)

		sha := ShaQuery(output)
		assert.Equal(t, tt.sha, sha, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.sha, sha)
	}
	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	avg := timeDiff / time.Duration(len(tests))
	fmt.Printf("TestShaQuery, Elapsed Time: %s, Avg per query: %s\n", timeDiff, avg)
}

func TestUUIDv5(t *testing.T) {
	maskParams := true
	t1 := time.Now()

	tests := []struct {
		input  string
		output string
		uid    string
	}{
		{"select * from users where id = 42", "(SELECT * FROM users WHERE (id = ?));", "a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"},
		{"select * from users where id = 74", "(SELECT * FROM users WHERE (id = ?));", "a2497c7b-dd5d-5be9-99b7-637eb8bacc4b"},
		{"DROP TABLE IF EXISTS listing;", "(DROP TABLE IF EXISTS listing);", "5dc5bf9b-f243-5bab-bf47-c306ed00d9ea"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()

		output := program.String(maskParams)
		assert.Equal(t, tt.output, output, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.output, output)

		uid := UuidV5(output).String()
		assert.Equal(t, tt.uid, uid, "input: %s\nprogram.String() not '%s'. got=%s", tt.input, tt.uid, uid)
	}
	t2 := time.Now()
	timeDiff := t2.Sub(t1)
	avg := timeDiff / time.Duration(len(tests))
	fmt.Printf("TestUUIDv5, Elapsed Time: %s, Avg per query: %s\n", timeDiff, avg)
}
