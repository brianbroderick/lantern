package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/brianbroderick/lantern/internal/sql/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is Lantern, a SQL parser.\n",
		user.Username)
	fmt.Printf("Try typing a select statement or exit to quit.\n")
	repl.Start(os.Stdin, os.Stdout)
}
