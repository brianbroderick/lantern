package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/brianbroderick/lantern/internal/sql/lexer"
	"github.com/brianbroderick/lantern/internal/sql/parser"
)

// REPL stands for Read Eval Print Loop. i.e., console

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		if line == "exit" {
			fmt.Fprintln(out, "Bye!")
			os.Exit(0)
		}

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

const LANTERN = `     ___!___
    /   |   \
   /    |    \
  /     |     \
 /______|______\
  \    )|(    / 
   \  ( | )  /
    \  =|=  /
     \__|__/   
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, LANTERN)
	io.WriteString(out, "\nWoops! Did someone turn the light off!\n")
	io.WriteString(out, "\n parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
