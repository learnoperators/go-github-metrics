package sdkstats

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

//RepoMetadata structure for storing Repository details
type RepoMetadata struct {
	URL          string
	Name         string
	Owner        string
	ProjectType  string
	Stars        int
	SDKVersion   string
	CreatedAt    string
	PushedAt     string
	TotalCommits int
}

//RepoMetadataQuery ...
type RepoMetadataQuery struct {
	ProjectType   string
	Queries       []string
	VersionParser VersionParser
}

//VersionParser ...
type VersionParser interface {
	ParseVersion(codeResults github.CodeResult, projectType string) (string, error)
}

//GetStats returns []RepoMetaData populated with Code results from Github Search Code API, for a given search string
func GetStats(client *github.Client, rq RepoMetadataQuery) ([]RepoMetadata, error) {
	var repoList []RepoMetadata
	var sdkVersion string
	for _, q := range rq.Queries {
		searchOp, err := Search(client, q)
		fmt.Println("HERE ", len(searchOp))
		if err != nil {
			return nil, err
		}
		if strings.Contains(q, "Gopkg.toml") {
			sdkVersion = "N/A"
		}
		for j := 0; j < len(searchOp); j++ {
			for i := 0; i < len(searchOp[j].CodeResults); i++ {
				if sdkVersion == "" {
					sdkVersion, err = rq.VersionParser.ParseVersion(searchOp[j].CodeResults[i], rq.ProjectType)
					if err != nil {
						return nil, err
					}
				}
				repoDetails := &RepoMetadata{
					URL:         searchOp[j].CodeResults[i].GetRepository().GetURL(),
					Name:        searchOp[j].CodeResults[i].GetRepository().GetName(),
					Owner:       searchOp[j].CodeResults[i].GetRepository().GetOwner().GetLogin(),
					ProjectType: rq.ProjectType,
					SDKVersion:  sdkVersion,
				}
				repoDetails, err = GetRepoDetails(client, repoDetails)
				repoList = append(repoList, *repoDetails)
			}
		}
	}
	return repoList, nil
}

//Search returns Github API results for Code Search, for a given Query string
func Search(client *github.Client, q string) ([]*github.CodeSearchResult, error) {
	var searchop []*github.CodeSearchResult
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
		searchop = append(searchop, op)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return searchop, nil
}

//GetRepoDetails Returns []RepoMetaData with Stars,CreatedAt,PushedAt, and TotalCommits details for a given owner, and repo
func GetRepoDetails(client *github.Client, repoDetails *RepoMetadata) (*RepoMetadata, error) {
	ctx := context.Background()
	repos, _, err := client.Repositories.Get(ctx, repoDetails.Owner, repoDetails.Name)
	if err != nil {
		return nil, err
	}
	commits, _, err := client.Repositories.ListContributorsStats(ctx, repoDetails.Owner, repoDetails.Name)
	var totalCommits int
	for _, r := range commits {
		totalCommits = totalCommits + r.GetTotal()
	}
	repoDetails.Stars = repos.GetStargazersCount()
	repoDetails.CreatedAt = repos.GetCreatedAt().String()
	repoDetails.PushedAt = repos.GetPushedAt().String()
	repoDetails.TotalCommits = totalCommits
	return repoDetails, nil
}
