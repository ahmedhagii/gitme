package main

import (
	"context"
	"flag"

	"github.com/ahmedhagii/gitme"
	"github.com/google/subcommands"
)

type logCmd struct {
	since  string
	until  string
	author string
	path   string
}

func (*logCmd) Name() string     { return "log" }
func (*logCmd) Synopsis() string { return "Shows a log of this repo's commit history" }
func (*logCmd) Usage() string {
	return `log [--author <github_name> ] [--since <DD-MM-YYYY>] [--until <DD-MM-YYYY>]
	\t[--parth <file_path>:
  Shows a history of commits.
`
}

func (c *logCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.author, "author", "", "Limit commits output to ones with specified author")
	f.StringVar(&c.since, "since", "", "Show commits after specified date")
	f.StringVar(&c.until, "until", "", "Show commits before specified date")
	f.StringVar(&c.since, "path", "", "Show commits only affecting the specified path whether it's a file or a directory")
}

func (c *logCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	_ = gitme.Log{}
	return subcommands.ExitSuccess
}
