package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/google/logger"
	"golang.org/x/oauth2"
)

const (
	username   = "binboum"
	repository = "argo-test"
	prefix     = "feature-"
	label      = "preview"
	randLength = 10
)

func RandStringRunes(n int) string {
	const letterRunes = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func commitFileToBranch(client *github.Client, owner, repo, branch, filePath, message string, content []byte) error {

	ctx := context.Background()
	fileContent := []byte("Fake file")

	opts := &github.RepositoryContentFileOptions{
		Message:   github.String("Add fakefile"),
		Content:   fileContent,
		Branch:    github.String(branch),
		Committer: &github.CommitAuthor{Name: github.String("Fake file"), Email: github.String("maxence@laude.pro")},
	}
	_, _, err := client.Repositories.CreateFile(ctx, owner, repo, "fakefile", opts)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func createPullRequestWithLabel(client *github.Client, owner, repo, title, body, head, base, label string) (*github.PullRequest, error) {
	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Body:  github.String(body),
		Head:  github.String(head),
		Base:  github.String(base),
	}

	pr, _, err := client.PullRequests.Create(context.Background(), owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("Error creating pull request: %v", err)
	}

	labelRequest := []string{label}
	_, _, err = client.Issues.AddLabelsToIssue(context.Background(), owner, repo, *pr.Number, labelRequest)
	if err != nil {
		return nil, fmt.Errorf("Error adding label to pull request: %v", err)
	}

	return pr, nil
}

func createBranches(accessToken string, typelabel string, numberBranch int) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	rand.Seed(time.Now().UnixNano())

	for i := 1; i <= numberBranch; i++ {
		randStr := RandStringRunes(randLength)
		branchName := fmt.Sprintf("%s%s", prefix, randStr)
		println(branchName)

		ref := &github.Reference{
			Ref:    github.String("refs/heads/" + branchName),
			Object: &github.GitObject{SHA: github.String("master")},
		}

		ref, _, err := client.Git.GetRef(context.Background(), username, repository, "refs/heads/main")
		if err != nil {
			fmt.Printf("Error fetching reference: %v\n", err)
		}

		ref.Ref = github.String("refs/heads/" + branchName)

		resp, _, err := client.Git.CreateRef(ctx, username, repository, ref)
		if err != nil {
			logger.Errorf("Erreur lors de la création de la branche : %v\nRéponse HTTP : %+v", err, resp)
		}

		content := []byte("Hello, World!")

		err = commitFileToBranch(client, username, repository, branchName, ".feature-branch", "Fake file", content)
		if err != nil {
			fmt.Printf("Error committing file: %v\n", err)
		}

		fmt.Printf("File %s committed to branch %s\n", "fake file", branchName)

		pr, err := createPullRequestWithLabel(client, username, repository, branchName, "Feature Branch", branchName, "main", "preview-"+typelabel)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		fmt.Printf("Pull request created: %s\n", *pr.HTMLURL)
	}
	return nil
}

func getAllBranches(ctx context.Context, client *github.Client, owner, repo string) ([]*github.Branch, error) {
	var allBranches []*github.Branch
	opts := &github.ListOptions{PerPage: 100}

	for {
		branches, resp, err := client.Repositories.ListBranches(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allBranches, nil
}

func deleteBranchesWithPrefix(accessToken string) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	branches, err := getAllBranches(ctx, client, username, repository)
	if err != nil {
		return err
	}

	for _, branch := range branches {
		if branch.GetName() == "master" || branch.GetName() == "main" {
			continue // skip main/master branch
		}
		if !branch.GetProtected() && strings.HasPrefix(branch.GetName(), prefix) {
			_, err := client.Git.DeleteRef(ctx, username, repository, "heads/"+branch.GetName())
			if err != nil {
				return err
			}
			fmt.Printf("Branch %s deleted successfully\n", branch.GetName())
		}
	}

	return nil
}

func main() {

	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken == "" {
		fmt.Println("GITHUB_TOKEN environment variable is not set")
		return
	}

	createFlag := flag.Bool("create", false, "Create branch")
	deleteFlag := flag.Bool("delete", false, "Delete branches with prefix")
	typePrFlag := flag.String("label", "", "Type of pr")
	numberPrFlag := flag.Int("number", 1, "Number of pr")

	flag.Parse()

	if *typePrFlag == "" {
		fmt.Println("label args required")
		return
	}

	if *createFlag {
		err := createBranches(accessToken, *typePrFlag, *numberPrFlag)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *deleteFlag {
		err := deleteBranchesWithPrefix(accessToken)
		if err != nil {
			fmt.Printf("Error deleting branches: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All branches with prefix", prefix, "deleted successfully.")
	}

	if !*createFlag && !*deleteFlag {
		flag.Usage()
		os.Exit(1)
	}
}
