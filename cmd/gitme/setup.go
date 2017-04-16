package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	// "github.com/ahmedhagii/gitme"
	// "github.com/google/go-github/github"
	"github.com/google/subcommands"
)

type setupCmd struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Token string `json:"token"`
}

func (*setupCmd) Name() string     { return "setup" }
func (*setupCmd) Synopsis() string { return "setup the data required for this tool" }
func (*setupCmd) Usage() string {
	return `
Usage:	gitme setup --owner <repo_owner_name> --repo <repo_name>
		--token <github_token>

`
}

func (c *setupCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.Owner, "owner", "", "Name of the repo's owner \"github/<owner>/<repo>\"")
	f.StringVar(&c.Repo, "repo", "", "Name of the repo \"github/<owner>/<repo>\"")
	f.StringVar(&c.Token, "token", "", "Generated github access token")
}

func (c *setupCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if c.Owner == "" {
		printMissing("repo owner name must be provided", c)
		return subcommands.ExitUsageError
	}
	if c.Repo == "" {
		printMissing("repo's name must be provided", c)
		return subcommands.ExitUsageError
	}
	if c.Token == "" {
		printMissing("you must provide your generated github access token", c)
		return subcommands.ExitUsageError
	}

	configj, err := json.Marshal(c)
	if err != nil {
		fmt.Println("failed to marshal setup data into json")
		return subcommands.ExitFailure
	}
	err = ioutil.WriteFile("/tmp/gitme-config", configj, 0644)
	if err != nil {
		fmt.Println("couldn't write to \"tmp/gitme-config\"")
	}
	return subcommands.ExitSuccess
}

func printMissing(msg string, c *setupCmd) {
	fmt.Printf("\n\t%s\n\n", msg)
	fmt.Println(c.Usage())
}
