package repl

// import (
// 	"bufio"
// 	"fmt"
// 	"io"
// 	"os"

// 	"github.com/brianbroderick/lantern/pkg/sql/ast"
// 	"github.com/brianbroderick/lantern/pkg/sql/lexer"
// 	"github.com/brianbroderick/lantern/pkg/sql/parser"
// )

// func StartParser(in io.Reader, out io.Writer) {
// 	scanner := bufio.NewScanner(in)

// 	var inspect *ast.Program
// 	maskParams := true

// 	for {
// 		fmt.Fprint(out, PROMPT)
// 		scanned := scanner.Scan()
// 		if !scanned {
// 			return
// 		}

// 		line := scanner.Text()

// 		if line == "exit" {
// 			fmt.Fprintln(out, "Bye!")
// 			os.Exit(0)
// 		}

// 		if line == "inspect" {
// 			io.WriteString(out, inspect.Inspect(maskParams))
// 			io.WriteString(out, "\n")
// 			continue
// 		}

// 		l := lexer.New(line)
// 		p := parser.New(l)

// 		program := p.ParseProgram()
// 		if len(p.Errors()) != 0 {
// 			printParserErrors(out, p.Errors())
// 			continue
// 		}

// 		inspect = program
// 		io.WriteString(out, program.String(maskParams))
// 		io.WriteString(out, "\n\n")
// 		io.WriteString(out, program.Inspect(maskParams))
// 		io.WriteString(out, "\n")
// 	}
// }
