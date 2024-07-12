package logs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/brianbroderick/lantern/internal/postgresql/lexer"
	"github.com/brianbroderick/lantern/internal/postgresql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/projectpath"
)

func Logs() {
	f := "postgresql.log.2024-07-10-1748"
	// f := "postgresql-2024-07-09_000000.log"
	// f := "postgresql-simple.log"
	path := filepath.Join(projectpath.Root, "logs", f)

	file, err := readPayload(path)
	if HasErr("processFile", err) {
		return
	}
	l := lexer.New(string(file))
	p := parser.New(l)
	program := p.ParseProgram()

	fmt.Println("Number of statements", len(program.Statements))
}

func HasErr(msg string, err error) bool {
	if err != nil {
		fmt.Printf("Message: %s\nHasErr: %s\n\n", msg, err.Error())
		return true
	}
	return false
}

func readPayload(file string) ([]byte, error) {
	if len(file) == 0 {
		return []byte{}, errors.New("file is empty")
	}

	data, err := os.ReadFile(string(file))
	if HasErr("readPayload", err) {
		return []byte{}, err
	}
	return data, nil
}

// func checkParserErrors(p *parser.Parser) {
// 	errors := p.Errors()
// 	if len(errors) == 0 {
// 		return
// 	}

// 	fmt.Printf("parser has %d errors", len(errors))
// 	for _, msg := range errors {
// 		fmt.Printf("parser error: %q", msg)
// 	}
// }
