package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

func main() {
	usageMsg := `usage: gitme <command> [<args>]

The following is a list of available gitme commands

   log        Shows a log of this repo's commit history
   setup      Setup the data required for this tool
   contrib    Get a list of contributors

See 'gitme <command> --help' to read about a specific subcommand.
`
	flag.CommandLine.Usage = func() {
		fmt.Println(usageMsg)
	}
	subcommands.Register(&logCmd{}, "")
	subcommands.Register(&setupCmd{}, "")
	subcommands.Register(&contribCmd{}, "")
	flag.Parse()

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
