package gitme

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"

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

func printCommit(ctx context.Context, rep *github.RepositoryCommit, owner string, repo string, exclude []string, client *github.Client) string {
	var output bytes.Buffer
	output.WriteString(fmt.Sprintf("%-8s  %s\n", colorize(color.FgYellow, "commit:"), colorize(color.FgYellow, rep.GetSHA())))
	output.WriteString(fmt.Sprintf("%-8s  %s\n", colorize(color.FgWhite, "author:"), colorize(color.FgWhite, rep.Author.GetLogin())))
	output.WriteString(fmt.Sprintf("message: %-s\n", rep.Commit.GetMessage()))
	output.WriteString(fmt.Sprintf("%-8s %v\n", "date:", rep.Commit.Author.GetDate()))

	r, _, _ := client.Repositories.GetCommit(ctx, owner, repo, rep.GetSHA())
	additions := 0
	deletions := 0
	fileChanges := []string{}

	for _, rr := range r.Files {
		pass := false
		for _, path := range exclude {
			if strings.Contains(rr.GetFilename(), path) {
				pass = true
				break
			}
		}
		if !pass {
			deletions += rr.GetDeletions()
			additions += rr.GetAdditions()
			fileChanges = append(fileChanges, rr.GetFilename()+" "+color.GreenString("+"+strconv.Itoa(rr.GetAdditions()))+
				" "+color.RedString("-"+strconv.Itoa(rr.GetDeletions())))
		}
		// fmt.Println(printDiffs(rr.GetPatch()))
	}
	output.WriteString(fmt.Sprintf("%s, %s\n",
		color.GreenString("Additions: "+strconv.Itoa(additions)),
		color.RedString("Deletions: "+strconv.Itoa(deletions))))
	output.WriteString(fmt.Sprintln(color.BlueString(r.GetHTMLURL())))
	output.WriteString(fmt.Sprintln())

	for _, file := range fileChanges {
		output.WriteString(fmt.Sprintf("\t%s\n", file))
	}
	output.WriteString(fmt.Sprintln("\n"))
	return output.String()
}

func ListCommits(owner string, repo string, exclude []string, opt *github.CommitsListOptions, client *github.Client) error {
	ctx := context.Background()

	// declare your pager
	cmd := exec.Command("less")
	// create a pipe (blocking)
	pipeReader, pipeWriter := io.Pipe()
	// defer pipeWriter.Close()
	// Set your i/o's
	cmd.Stdin = pipeReader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()

	go func() error {
		for page := 1; true; page++ {
			opt.ListOptions = github.ListOptions{Page: page, PerPage: 50}
			commitList, _, err := client.Repositories.ListCommits(ctx, owner, repo, opt)

			if err != nil {
				fmt.Println("errror getting list of commits", err)
				return err
			}
			// just check where to stop requesting more pages
			if len(commitList) == 0 {
				pipeWriter.Close()
				break
			}
			out1 := make(chan CommitResult, len(commitList))
			var wg sync.WaitGroup
			for ind, rep := range commitList {
				go func(ind int, rep *github.RepositoryCommit) {
					wg.Add(1)
					defer wg.Done()
					str := printCommit(ctx, rep, owner, repo, exclude, client)
					out1 <- CommitResult{index: ind, output: str}
				}(ind, rep)
			}

			commitResults := CommitResults{}
			for i := 0; i < len(commitList); i++ {
				commitResults = append(commitResults, <-out1)
			}
			close(out1)
			wg.Wait()

			sort.Sort(commitResults)
			for _, val := range commitResults {
				fmt.Fprintf(pipeWriter, val.output)
			}
		}
		return nil
	}()

	cmd.Wait()
	// fmt.Println(">>>>>>>>>>> out", pagerErr)
	// if pagerErr == nil {
	// 	pipeWriter.Close()
	// 	pipeReader.Close()
	// 	fmt.Println("breaking")
	// 	return nil
	// }
	pipeWriter.Close()
	pipeReader.Close()
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

type Config struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

func main() {
	color.NoColor = false
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Gmark is a tool for grading projects residing on github.\n\n")
		flag.PrintDefaults()
	}
	// configj, _ := json.Marshal(&Config{Owner: "jfkdf", Repo: "kdjfd"})
	// ioutil.WriteFile("/tmp/gitme-config", configj, 0644)
	// data, _ := ioutil.ReadFile("/tmp/gitme-config")
	// config := Config{}
	// _ = json.Unmarshal(data, &config)
	// fmt.Printf("%+v", config)
	// return
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
	_ = tc
	client := github.NewClient(nil)

	// list all repositories for the authenticated user
	// repos, _, _ := client.Repositories.List(ctx, "", nil)
	// // for _, repo := range repos {
	// // fmt.Println(repo.GetName())
	// // }

	// logins, err := listContributors("secourse2016", "404notfound", client)
	// if err != nil {
	// }
	_ = client
	// _ = ListCommits("secourse2016", "404notfound", client)

	// log.Fatal(err)
	// fmt.Println(logins)
}
