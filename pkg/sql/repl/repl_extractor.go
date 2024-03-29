package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/brianbroderick/lantern/pkg/sql/ast"
	"github.com/brianbroderick/lantern/pkg/sql/extractor"
	"github.com/brianbroderick/lantern/pkg/sql/lexer"
	"github.com/brianbroderick/lantern/pkg/sql/object"
	"github.com/brianbroderick/lantern/pkg/sql/parser"
	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func StartParser(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	var inspect *ast.Program
	maskParams := true

	lines := []string{}

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		cmd := scanner.Text()

		if cmd == "exit" {
			fmt.Fprintln(out, "Bye!")
			os.Exit(0)
		}

		if cmd == "inspect" {
			io.WriteString(out, inspect.Inspect(maskParams))
			io.WriteString(out, "\n")
			continue
		}

		lines = append(lines, cmd)

		if len(cmd) == 0 || cmd[len(cmd)-1:] != ";" {
			continue
		}

		line := strings.Join(lines, " ")

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		lines = []string{}
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())

			continue
		}

		inspect = program

		for i, s := range program.Statements {
			r := extractor.NewExtractor(&s)
			env := object.NewEnvironment()
			r.Extract(s, env)
			// Check errors with Extractor
			if len(r.Errors()) != 0 {
				printParserErrors(out, r.Errors())
				continue
			}

			fmt.Printf("\nStatement: %d\n\n", i+1)

			// Print out the columns
			if len(r.Columns) > 0 {
				fmt.Println("Selected Columns:")
				for fqcn, column := range r.Columns {
					if column.Clause == token.COLUMN {
						io.WriteString(out, fmt.Sprintf("  %s\n", fqcn))
					}
				}
				fmt.Println("")
			}

			// Print out the tables
			if len(r.Tables) > 0 {
				fmt.Println("Tables:")
				for _, table := range r.Tables {
					io.WriteString(out, fmt.Sprintf("  %s\n", table.Name))
				}
				fmt.Println("")
			}

			// Print out the joins
			if len(r.TableJoins) > 0 {
				fmt.Println("Joins:")
				for _, join := range r.TableJoins {
					io.WriteString(out, fmt.Sprintf("  %s TO %s\n", join.TableA, join.TableB))
				}
				fmt.Println("")
			}

			// Print out the columns
			if len(r.Columns) > 0 {
				fmt.Println("Where Columns:")
				for fqcn, column := range r.Columns {
					if column.Clause == token.WHERE {
						io.WriteString(out, fmt.Sprintf("  %s\n", fqcn))
					}
				}
				fmt.Println("")
			}

			// Print out the functions
			if len(r.Functions) > 0 {
				fmt.Println("Functions:")
				for _, function := range r.Functions {
					io.WriteString(out, fmt.Sprintf("  %s\n", function.Name))
				}
				fmt.Println("")
			}
		}

		io.WriteString(out, program.String(maskParams))
		io.WriteString(out, "\n\n")
	}
}
