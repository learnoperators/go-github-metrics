package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

//FetchStarGazers ccc
func FetchStarGazers(client *github.Client, op []*github.CodeSearchResult) map[int][]string {
	context := context.Background()
	var starGazer map[string]int
	starGazer = make(map[string]int)
	for i := 0; i < len(op); i++ {
		for _, r := range op[i].CodeResults {
			owner := r.GetRepository().GetOwner().GetLogin()
			reponame := r.GetRepository().GetName()
			fullname := r.GetRepository().GetFullName()
			repoDetails, _, err := client.Repositories.Get(context, owner, reponame)
			if err != nil {
				fmt.Println(err)
			}
			starGazer[fullname] = repoDetails.GetStargazersCount()
		}
	}
	sortedStarGazers := SortMyStars(starGazer)
	return sortedStarGazers
}

//SortMyStars kk
func SortMyStars(starGazer map[string]int) map[int][]string {
	reverseStarGazer := map[int][]string{}
	for k, v := range starGazer {
		reverseStarGazer[v] = append(reverseStarGazer[v], k)
	}
	return reverseStarGazer
}

func main() {

	context := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "...Your Token....."},
	)
	tc := oauth2.NewClient(context, ts)
	client := github.NewClient(tc)

	opts := &github.SearchOptions{TextMatch: true,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var helmOp []*github.CodeSearchResult
	for {
		op, resp, err := client.Search.Code(context, "filename:Dockerfile quay.io/operator-framework/helm-operator", opts)
		if err != nil {
			fmt.Println(err)
		}
		helmOp = append(helmOp, op)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	var ansibleOp []*github.CodeSearchResult
	for {
		op, resp, err := client.Search.Code(context, "filename:Dockerfile quay.io/operator-framework/ansible-operator", opts)
		if err != nil {
			fmt.Println(err)
		}
		ansibleOp = append(ansibleOp, op)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	var goOp []*github.CodeSearchResult
	for {
		op, resp, err := client.Search.Code(context, "github.com/operator-framework/operator-sdk filename:go.mod filename:Gopkg.toml", opts)
		if err != nil {
			fmt.Println(err)
		}
		goOp = append(goOp, op)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	fmt.Println("***********HELM BASED REPOS*********************************")
	var helmStarGazers = FetchStarGazers(client, helmOp)
	fmt.Println(helmStarGazers)

	fmt.Println("***********ANSIBLE BASED REPOS*********************************")
	var ansibleStarGazers = FetchStarGazers(client, ansibleOp)
	fmt.Println(ansibleStarGazers)

	fmt.Println("***********GO BASED REPOS*********************************")
	var goStargazers = FetchStarGazers(client, goOp)
	fmt.Println(goStargazers)

}
