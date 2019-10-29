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
	//Search function return CodeResults for the search string, including Basic repository data.
	searchOp, err := Search(client, q)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if strings.Contains(q, "helm") {
		projectType = "helm"
	} else if strings.Contains(q, "ansible") {
		projectType = "ansible"
	} else if strings.Contains(q, "go.mod") {
		projectType = "go.mod"
	} else if strings.Contains(q, "Gopkg.toml") {
		projectType = "Gopkg.toml"
	} else {
		projectType = ""
	}

	for j := 0; j < len(searchOp); j++ {
		for i := 0; i < len(searchOp[j].CodeResults); i++ {
			owner := searchOp[j].CodeResults[i].GetRepository().GetOwner().GetLogin()
			name := searchOp[j].CodeResults[i].GetRepository().GetName()
			fullName := searchOp[j].CodeResults[i].GetRepository().GetFullName()

			//getVersion returns SDK version by doing Fragment search
			if projectType == "Gopkg.toml" || projectType == "" {
				sdkVersion = 0
			} else {
				sdkVersion, err = strconv.Atoi(getversion(searchOp[j].CodeResults[i], projectType))
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
			}

			//GetRepoDetails returns Repository specififc details
			repoMap, err := GetRepoDetails(client, owner, name)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			if projectType == "" {
				projectType = repoMap["Language"]
			}
			TotalCommits, err := strconv.Atoi(repoMap["TotalCommits"])
			Stars, err := strconv.Atoi(repoMap["Stars"])
			if err != nil {
				fmt.Println(err)
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
	//Below Logic enables to collate Search results from all pages returned from API call.
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

//GetRepoDetails Returns map[string][string] with Stars,CreatedAt,PushedAt, and TotalCommits details for a given owner, and repo
func GetRepoDetails(client *github.Client, owner string, name string) (map[string]string, error) {

	repoMap := map[string]string{}

	ctx := context.Background()

	repos, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		fmt.Println(err)
		return nil, err
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
	repoMap["Language"] = repos.GetLanguage()
	return repoMap, nil

}

//getVersion takes each Code result from a search output, and searches for Fragments comtaining VERSION, and returns the same.
//The Fragments Text Macthes used are specific to Operator-SDK Version Pattern.
func getversion(codeResults github.CodeResult, projectType string) string {

	var searchQ, SDKversion, searchLatest string
	SDKLatest := "11"

	if projectType == "helm" {
		searchQ = "quay.io/operator-framework/helm-operator:v0."
		searchLatest = "quay.io/operator-framework/helm-operator:latest"
	}
	if projectType == "ansible" {
		searchQ = "quay.io/operator-framework/ansible-operator:v0."
		searchLatest = "quay.io/operator-framework/ansible-operator:latest"

	}
	if projectType == "go.mod" {
		searchQ = "github.com/operator-framework/operator-sdk v0."
	}

	for _, r := range codeResults.TextMatches {
		if strings.Contains(r.GetFragment(), searchQ) {
			posFirst := strings.Index(r.GetFragment(), searchQ)
			if posFirst == -1 {
				SDKversion = "N/A"
			}
			posFirstAdjusted := posFirst + len(searchQ)
			runes := []rune(r.GetFragment())
			c := posFirstAdjusted + 2
			if c == -1 {
				SDKversion = "N/A"
			}
			fmt.Println("HERE", string(runes[posFirstAdjusted:c]))
			SDKversion = strings.Trim(string(runes[posFirstAdjusted:c]), ".")
			fmt.Println("THERE", SDKversion)
		} else if (projectType == "helm" || projectType == "ansible") && strings.Contains(r.GetFragment(), searchLatest) {
			SDKversion = SDKLatest
		}
	}
	if SDKversion == "" {
		SDKversion = "0"
	}
	return SDKversion
}
