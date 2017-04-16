package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/ahmedhagii/gitme"
	"github.com/google/go-github/github"
	"github.com/google/subcommands"
	"golang.org/x/oauth2"
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
	return `
	Shows a history of commits.

usage:	log [--author <github_name> ] [--since <DD-MM-YYYY>] [--until <DD-MM-YYYY>]
	    [--parth <file_path>]:

`
}

func (c *logCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.author, "author", "", "Limit commits output to ones with specified author")
	f.StringVar(&c.since, "since", "", "Show commits after specified date")
	f.StringVar(&c.until, "until", "", "Show commits before specified date")
	f.StringVar(&c.path, "path", "", "Show commits only affecting the specified path whether it's a file or a directory")
}

func (c *logCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	configData, err := ioutil.ReadFile("/tmp/gitme-config")
	if err != nil {
		fmt.Println(`couldn't read config file at "/tmp/gtime-config"
			run gitme setup`)
	}
	config := setupCmd{}
	_ = json.Unmarshal(configData, &config)

	// move to util function
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	oauthClient := oauth2.NewClient(ctx, ts)
	client := github.NewClient(oauthClient)
	//

	opt := github.CommitsListOptions{}
	if c.author != "" {
		opt.Author = c.author
	}
	if c.path != "" {
		opt.Path = c.path
	}
	if c.since != "" {
		sinceDate, err := time.Parse("02-01-2006", c.since)
		if err != nil {
			fmt.Println("pass date to --since in the correct format DD-MM-YYYY", err)
			return subcommands.ExitFailure
		}
		opt.Since = sinceDate
	}
	if c.until != "" {
		untilDate, err := time.Parse("02-01-2006", c.until)
		if err != nil {
			fmt.Println("pass date to --until in the correct format DD:MM:YYY", err)
			return subcommands.ExitFailure
		}
		opt.Until = untilDate
	}
	opt.ListOptions = github.ListOptions{PerPage: 1000}

	gitme.ListCommits(config.Owner, config.Repo, &opt, client)

	return subcommands.ExitSuccess
}
