package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/brianbroderick/lantern/internal/postgresql/logs"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "help":
		helpCmd := flag.NewFlagSet("help", flag.ExitOnError)
		helpCmd.Parse(os.Args[2:])
		printHelp()
		os.Exit(0)
	case "process":
		strArgs, _, boolArgs := processCli(os.Args)
		if boolArgs["rebuildJson"] != nil && *boolArgs["rebuildJson"] {
			// f := "postgresql.log.2024-07-10-1748.cp"

			fmt.Println("Processing log file", *strArgs["file"])
			log := logs.LoadLogFile(*strArgs["file"])
			logs.AggregateLogs(*strArgs["file"], log, "queries.json", "databases.json")
		}

		logs.UpsertQueries()
		logs.UpsertDatabases()
		logs.ExtractAndUpsertQueryMetadata()
	default:
		printHelp()
		os.Exit(1)
	}

}

func processCli(args []string) (map[string]*string, map[string]*int, map[string]*bool) {
	processCmd := flag.NewFlagSet("process", flag.ExitOnError)

	strArgs := make(map[string]*string)
	intArgs := make(map[string]*int)
	boolArgs := make(map[string]*bool)

	boolArgs["rebuildJson"] = processCmd.Bool("rebuild_json", true, "Rebuild the json files from the logs")
	strArgs["file"] = processCmd.String("file", "", "File to be processed")

	processCmd.Parse(args[2:])

	return strArgs, intArgs, boolArgs
}

func printHelp() {
	helpText := `
  Usage: lantern-logs [command] [arguments]

  lantern-logs help             - Print this help message
	lantern-logs process          - Process a log file
		--rebuild_json=false        - Rebuild the json files from the logs
		--file=                     - File to be processed
	`

	fmt.Println(helpText)
}
