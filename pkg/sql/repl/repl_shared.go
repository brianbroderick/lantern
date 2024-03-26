package repl

import "io"

const PROMPT = ">> "

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
	io.WriteString(out, "\nWhoops! Did someone turn the light off!\n")
	io.WriteString(out, "\n parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
