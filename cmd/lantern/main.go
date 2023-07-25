package main

import (
	"fmt"
	"os"

	"github.com/brianbroderick/lantern/internal/postgresql/repl"
)

func main() {
	fmt.Printf("This is Lantern, the PostgreSQL log analyzer.\n")
	fmt.Printf("Enter a log line to analyze:\n")
	repl.Start(os.Stdin, os.Stdout)
}
