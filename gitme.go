package gitme

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	authorName = flag.String("author", "", "the author to show commits for")
	until      = flag.String("until", "", "show commits until this date")
	since      = flag.String("since", "", "show commits since this date")
	listRepos  = flag.String("list", "", "list all repos on the system")
	addRepo    = flag.String("add-repo", "", "add this repo to the system")
	printAll   = flag.Bool("print-all", false, "print all output at once, useful for piping output to a file")
)

type CommitResult struct {
	index  int
	output string
}
type CommitResults []CommitResult

func (slice CommitResults) Len() int {
	return len(slice)
}
func (slice CommitResults) Less(i, j int) bool {
	return slice[i].index < slice[j].index
}
func (slice CommitResults) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func ListContributors(owner string, repo string, client *github.Client) ([]string, error) {
	ctx := context.Background()
	repos2, _, err := client.Repositories.ListCollaborators(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}
	loginsArray := []string{}
	for _, repo := range repos2 {
		loginsArray = append(loginsArray, repo.GetLogin())
		// fmt.Println(repo.GetLogin())
	}
	return loginsArray, nil
}

// func printCommit(ctx context.Context, rep *github.RepositoryCommit, owner string, repo string, client *github.Client) {
// 	fmt.Printf("%-8s %s\n", colorize(color.FgYellow, "commit:"), colorize(color.FgYellow, rep.GetSHA()))
// 	fmt.Printf("message: %-s\n", rep.Commit.GetMessage())
// 	fmt.Printf("%-8s %v\n", "date:", rep.Commit.Author.GetDate())

// 	r, _, _ := client.Repositories.GetCommit(ctx, owner, repo, rep.GetSHA())
// 	fmt.Printf("%s, %s\n",
// 		color.GreenString("Additions: "+strconv.Itoa(r.Stats.GetAdditions())),
// 		color.RedString("Deletions: "+strconv.Itoa(r.Stats.GetDeletions())))
// 	fmt.Println(color.BlueString(r.GetHTMLURL()))
// 	fmt.Println()

// 	files := r.Files
// 	for _, rr := range files {
// 		fmt.Printf("   %s\n", rr.GetFilename())
// 		// fmt.Println(printDiffs(rr.GetPatch()))
// 	}
// 	fmt.Println("\n")
// }

func printCommit(ctx context.Context, rep *github.RepositoryCommit, owner string, repo string, client *github.Client) string {
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("%-8s %s\n", colorize(color.FgYellow, "commit:"), colorize(color.FgYellow, rep.GetSHA())))
	output.WriteString(fmt.Sprintf("message: %-s\n", rep.Commit.GetMessage()))
	output.WriteString(fmt.Sprintf("%-8s %v\n", "date:", rep.Commit.Author.GetDate()))

	r, _, _ := client.Repositories.GetCommit(ctx, owner, repo, rep.GetSHA())
	output.WriteString(fmt.Sprintf("%s, %s\n",
		color.GreenString("Additions: "+strconv.Itoa(r.Stats.GetAdditions())),
		color.RedString("Deletions: "+strconv.Itoa(r.Stats.GetDeletions()))))
	output.WriteString(fmt.Sprintln(color.BlueString(r.GetHTMLURL())))
	output.WriteString(fmt.Sprintln())

	files := r.Files
	for _, rr := range files {
		output.WriteString(fmt.Sprintf("\t%s\n", rr.GetFilename()))
		// fmt.Println(printDiffs(rr.GetPatch()))
	}
	output.WriteString(fmt.Sprintln("\n"))
	return output.String()
}

func listCommits(owner string, repo string, client *github.Client) error {
	ctx := context.Background()

	opt := github.CommitsListOptions{Author: "YasmeenWafa"}
	if *since != "" {
		sinceDate, err := time.Parse("02-01-2006", *since)
		if err != nil {
			log.Fatal("pass date to --since in the correct format DD-MM-YYYY", err)
		}
		opt.Since = sinceDate
	}
	if *until != "" {
		untilDate, err := time.Parse("02-01-2006", *until)
		if err != nil {
			log.Fatal("pass date to --until in the correct format DD:MM:YYY", err)
		}
		opt.Until = untilDate
	}
	opt.ListOptions = github.ListOptions{PerPage: 1000}
	repos, _, err := client.Repositories.ListCommits(ctx, owner, repo, &opt)
	if err != nil {
		return nil
	}

	var wg sync.WaitGroup
	out1 := make(chan CommitResult, len(repos))
	for ind, rep := range repos {
		go func(ind int, rep *github.RepositoryCommit) {
			wg.Add(1)
			defer wg.Done()
			// fmt.Println("done ", ind, rep.GetSHA())
			str := printCommit(ctx, rep, owner, repo, client)
			out1 <- CommitResult{index: ind, output: str}
			// fmt.Println(str)
		}(ind, rep)
	}

	commitResults := CommitResults{}
	for i := 0; i < len(repos); i++ {
		commitResults = append(commitResults, <-out1)
	}
	close(out1)
	wg.Wait()
	fmt.Printf("%s\n\n", colorize(color.FgYellow, "Total commits:", strconv.Itoa(len(repos))))

	sort.Sort(commitResults)
	if *printAll {
		for _, val := range commitResults {
			fmt.Println(val.output)
		}
	} else {
		printOnDemand(commitResults)
	}
	return nil
}

func printOnDemand(output CommitResults) {
	lines := []string{}
	for _, val := range output {
		splitted := strings.Split(val.output, "\n")
		for _, line := range splitted {
			lines = append(lines, line)
		}
	}
	fmt.Println("yal")
	for i := 0; i < len(lines); i++ {
		if i < 30 {
			fmt.Printf("%-70s\n", lines[i])
		} else if i == 30 {
			fmt.Printf("%-70s", lines[i])
		} else {

			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			if text[0] == 'q' {
				break
			}
			fmt.Printf("%-70s", lines[i])
		}
	}
}

func printDiffs(content string) string {
	lines := strings.Split(content, "\n")
	var output bytes.Buffer

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if line[0] == '+' {
			c := color.New(color.FgGreen, color.Bold).SprintFunc()
			output.WriteString(c(line))
		} else if line[0] == '-' {
			c := color.New(color.FgRed, color.Bold).SprintFunc()
			output.WriteString(c(line))
		} else {
			c := color.New(color.FgWhite, color.Bold).SprintFunc()
			output.WriteString(c(line))
		}
		output.WriteString("\n")
	}
	return output.String()
}

func colorize(colorVar color.Attribute, str ...string) string {
	colorAgent := color.New(colorVar, color.Bold).SprintFunc()
	return colorAgent(strings.Join(str, " "))
}

func main() {
	color.NoColor = false
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Gmark is a tool for grading projects residing on github.\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	// client := github.NewClient(nil)

	// _ = os.Chdir("/Users/Ahmed/go/src/gmark")
	// out, _ := exec.Command("more", "gmark.go").Output()
	// fmt.Println(out)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "c57d151c767b927601458fdae8b88b23a7529788"},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all repositories for the authenticated user
	// repos, _, _ := client.Repositories.List(ctx, "", nil)
	// // for _, repo := range repos {
	// // fmt.Println(repo.GetName())
	// // }

	// logins, err := listContributors("secourse2016", "404notfound", client)
	// if err != nil {
	// }
	_ = listCommits("secourse2016", "404notfound", client)
	// log.Fatal(err)
	// fmt.Println(logins)
}
