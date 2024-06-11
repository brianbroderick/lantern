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
			r := extractor.NewExtractor(&s, true, extractor.DefaultUUID)
			env := object.NewEnvironment()
			r.Extract(s, env)
			// Check errors with Extractor
			if len(r.Errors()) != 0 {
				printParserErrors(out, r.Errors())
				continue
			}

			io.WriteString(out, fmt.Sprintf("\nStatement: %d\n\n", i+1))

			// Print out the columns
			if len(r.ColumnsInQueries) > 0 {
				io.WriteString(out, "Columns in Query:\n")
				io.WriteString(out, fmt.Sprintf("  %-8s %s\n", "Clause", "Name"))
				io.WriteString(out, fmt.Sprintf("  %-8s %s\n", "--------", "--------"))
				for _, column := range r.ColumnsInQueries {
					io.WriteString(out, fmt.Sprintf("  %-8s %s\n", column.Clause, column.Name))
				}
				io.WriteString(out, "\n")
			}

			// Print out the tables
			if len(r.TablesInQueries) > 0 {
				io.WriteString(out, "Tables:\n")

				for _, table := range r.TablesInQueries {
					io.WriteString(out, fmt.Sprintf("  %s\n", table.Name))
				}
				io.WriteString(out, "\n")
			}

			// Print out the joins
			if len(r.TableJoinsInQueries) > 0 {
				io.WriteString(out, "Joins:\n")
				for _, join := range r.TableJoinsInQueries {
					io.WriteString(out, fmt.Sprintf("  %s TO %s\n", join.TableA, join.TableB))
				}
				io.WriteString(out, "\n")
			}

			// Print out the functions
			if len(r.FunctionsInQueries) > 0 {
				io.WriteString(out, "Functions:\n")

				for _, function := range r.FunctionsInQueries {
					io.WriteString(out, fmt.Sprintf("  %s\n", function.Name))
				}
				io.WriteString(out, "\n")
			}
		}

		io.WriteString(out, program.String(maskParams))
		io.WriteString(out, "\n\n")
	}
}
