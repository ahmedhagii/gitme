package main

import (
	"context"
	"flag"
	// "fmt"
	"os"

	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&logCmd{}, "")
	subcommands.Register(&setupCmd{}, "")
	flag.Parse()

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
