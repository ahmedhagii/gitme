package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"

	"github.com/ahmedhagii/gitme/misc"
	"github.com/google/go-github/github"
	"github.com/google/subcommands"
)

type contribCmd struct {
	owner string
	repo  string
	token string
}

func (*contribCmd) Name() string     { return "contrib" }
func (*contribCmd) Synopsis() string { return "Get a list of contributors" }
func (*contribCmd) Usage() string {
	return `
	Shows a list of contributors to the specified repository. It uses the info
	provided to 'gitme setup' by default, but can be overridden by setting the flags
	below. However, getting a list of contributors requires authentication,
	so either provide a new access token or the old saved one will be used.

usage:	contrib --owner <repo_owner_name> --repo <repo_name> --token <github_token>

`
}

func (c *contribCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.owner, "owner", "", "Name of the repo's owner 'github/<owner>/<repo>'")
	f.StringVar(&c.repo, "repo", "", "Name of the repo 'github/<owner>/<repo>'")
	f.StringVar(&c.token, "token", "", "Generated github access token")
}

func (c *contribCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if (c.owner == "" && c.repo != "") || (c.owner != "" && c.repo == "") {
		fmt.Println("if you want to override the repo info, you must provide both --owner and --repo")
		return subcommands.ExitFailure
	}
	configData, err := ioutil.ReadFile("/tmp/gitme-config")
	if err != nil {
		fmt.Println(`couldn't read config file at "/tmp/gtime-config"
			run gitme setup <args>`)
	}
	config := setupCmd{}
	_ = json.Unmarshal(configData, &config)

	if c.owner != "" {
		config.Owner = c.owner
		config.Repo = c.repo
		if c.token != "" {
			config.Token = c.token
		}
	}

	// move to util function
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	oauthClient := oauth2.NewClient(ctx, ts)
	client := github.NewClient(oauthClient)
	//

	contribs, err := misc.ListContributors(config.Owner, config.Repo, client)
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	misc.OutputToPager(contribs)
	return subcommands.ExitSuccess
}
