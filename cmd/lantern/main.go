package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/brianbroderick/lantern/pkg/sql/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is Lantern, a SQL analyzer.\n",
		user.Username)
	fmt.Println("Try typing a query or exit to quit.")
	fmt.Println("To allow multi-line queries, end your query with a semicolon.")
	fmt.Println("")
	repl.StartParser(os.Stdin, os.Stdout)
}
