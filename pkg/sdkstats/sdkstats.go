package sdkstats

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

//RepoMetadata structure for storing Repository details
type RepoMetadata struct {
	FullName     string
	ProjectType  string
	Stars        int
	SDKVersion   int
	CreatedAt    string
	PushedAt     string
	TotalCommits int
}

//GetStats returns []RepoMetaData populated with Code results from Github Search Code API, for a given search string
func GetStats(client *github.Client, q string) ([]RepoMetadata, error) {
	var repoList []RepoMetadata
	var projectType string
	var sdkVersion int

	//Search function return CodeResults for the q queery, and also Basic repository data.
	searchOp, err := Search(client, q)

	if err != nil {
		return nil, err
	}

	if strings.Contains(q, "helm") {
		projectType = "helm"
	}
	if strings.Contains(q, "ansible") {
		projectType = "ansible"
	}
	if strings.Contains(q, "go.mod") {
		projectType = "go.mod"
	}
	if strings.Contains(q, "Gopkg.toml") {
		projectType = "Gopkg.toml"
	}
	for j := 0; j < len(searchOp); j++ {
		for i := 0; i < len(searchOp[j].CodeResults); i++ {
			owner := searchOp[j].CodeResults[i].GetRepository().GetOwner().GetLogin()
			name := searchOp[j].CodeResults[i].GetRepository().GetName()
			fullName := searchOp[j].CodeResults[i].GetRepository().GetFullName()

			//getVersion returns SDK version by doing Fragment search
			if projectType == "Gopkg.toml" {
				sdkVersion = 0
			} else {
				sdkVersion = getversion(searchOp[j].CodeResults[i], projectType)
			}

			//GetRepoDetails returns Repository specififc details
			repoMap := GetRepoDetails(client, owner, name)

			TotalCommits, err := strconv.Atoi(repoMap["TotalCommits"])
			Stars, err := strconv.Atoi(repoMap["Stars"])
			if err != nil {
				return nil, err
			}

			repoDetails := &RepoMetadata{
				FullName:     fullName,
				ProjectType:  projectType,
				SDKVersion:   sdkVersion,
				Stars:        Stars,
				CreatedAt:    repoMap["CreatedAt"],
				PushedAt:     repoMap["PushedAt"],
				TotalCommits: TotalCommits,
			}
			repoList = append(repoList, *repoDetails)

		}
	}

	return repoList, nil
}

//Search returns Github API results for Code Search, for a given Query string
func Search(client *github.Client, q string) ([]*github.CodeSearchResult, error) {

	var searchOp []*github.CodeSearchResult
	ctx := context.Background()

	opts := &github.SearchOptions{TextMatch: true,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	for {
		op, resp, err := client.Search.Code(ctx, q, opts)
		if err != nil {
			return nil, err
		}
		searchOp = append(searchOp, op)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return searchOp, nil

}

//GetRepoDetails Returns map[string][string] with Stars,CreatedAt,PushedAt, and TotalCommits details for a given onwer, and repo
func GetRepoDetails(client *github.Client, owner string, name string) map[string]string {

	repoMap := map[string]string{}

	ctx := context.Background()

	repos, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		fmt.Println(err)
	}
	commits, _, err := client.Repositories.ListContributorsStats(ctx, owner, name)
	var total int
	for _, r := range commits {
		total = total + r.GetTotal()
	}
	repoMap["Stars"] = strconv.Itoa(repos.GetStargazersCount())
	repoMap["CreatedAt"] = repos.GetCreatedAt().String()
	repoMap["PushedAt"] = repos.GetPushedAt().String()
	repoMap["TotalCommits"] = strconv.Itoa(total)
	return repoMap

}

func getversion(codeResults github.CodeResult, projectType string) int {

	var searchQ string
	var SDKversion int

	if projectType == "helm" {
		searchQ = "quay.io/operator-framework/helm-operator:v0."
	}
	if projectType == "ansible" {
		searchQ = "quay.io/operator-framework/ansible-operator:v0."
	}
	if projectType == "go.mod" {
		searchQ = "github.com/operator-framework/operator-sdk v0."
	}

loop:
	for j := 1; j <= 11; j++ {
		for _, r := range codeResults.TextMatches {
			if strings.Contains(r.GetFragment(), searchQ+strconv.Itoa(j)+".") {
				SDKversion = j
				break loop
			}
		}
	}
	return SDKversion
}
