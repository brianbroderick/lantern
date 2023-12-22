package repl

import (
	"io"
)

// REPL stands for Read Eval Print Loop. i.e., console

func StartEval(in io.Reader, out io.Writer) {
	// scanner := bufio.NewScanner(in)

	// for {
	// 	fmt.Fprintf(out, PROMPT)

	// 	scanned := scanner.Scan()
	// 	if !scanned {
	// 		return
	// 	}
	// 	line := scanner.Text()

	// 	if line == "exit" {
	// 		fmt.Fprintln(out, "Bye!")
	// 		os.Exit(0)
	// 	}

	// 	l := lexer.New(line)
	// 	p := parser.New(l)

	// 	program := p.ParseProgram()
	// 	if len(p.Errors()) != 0 {
	// 		printParserErrors(out, p.Errors())
	// 		continue
	// 	}
	// 	evaluated := evaluator.Eval(program, nil)
	// 	if evaluated != nil {
	// 		io.WriteString(out, evaluated.Inspect())
	// 		io.WriteString(out, "\n")
	// 	}
	// }
}
