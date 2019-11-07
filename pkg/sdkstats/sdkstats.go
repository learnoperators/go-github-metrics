package sdkstats

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
)

// RepoMetadata struct is for storing individual epository details.
type RepoMetadata struct {
	URL          string
	Name         string
	Owner        string
	ProjectType  string
	Stars        int
	Version      string
	CreatedAt    string
	PushedAt     string
	TotalCommits int
}

// RepoMetadataQuery struct provides search queries, projecttype, and VersionParser declaration
type RepoMetadataQuery struct {
	ProjectType   string
	Queries       []string
	VersionParser VersionParser
}

// VersionParser Interface implements ParseVersion(), to parse version number from a given Text Match result.
type VersionParser interface {
	ParseVersion(codeResults github.CodeResult) (string, error)
}

// GetStats returns List of Repositories satisfyng the Search criteria,
// populated with Stars, TotalCommits, Initial and Last Commit details.
// Also included are basic details of repository such as URL, Owner, and name.

func GetStats(ctx context.Context, tc *http.Client, rq RepoMetadataQuery) ([]RepoMetadata, error) {
	var repoList []RepoMetadata
	for _, q := range rq.Queries {
		searchOp, err := search(ctx, tc, q)
		if err != nil {
			return nil, err
		}
		for _, j := range searchOp {
			for _, i := range j.CodeResults {
				repoDetails, err := getRepoDetails(ctx, tc, i, rq)
				if _, ok := err.(*github.AcceptedError); ok {
					continue
				} else if err != nil {
					return nil, err
				}
				repoList = append(repoList, *repoDetails)
			}
		}
	}
	return repoList, nil
}

// search returns Github API results for Code Search, for a given Query string
func search(ctx context.Context, tc *http.Client, q string) ([]*github.CodeSearchResult, error) {
	var searchop []*github.CodeSearchResult
	client := github.NewClient(tc)

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

// getRepoDetails Returns []RepoMetaData with Stars,CreatedAt,PushedAt, and TotalCommits details for a given owner, and repo.
func getRepoDetails(ctx context.Context, tc *http.Client, codeResults github.CodeResult, rq RepoMetadataQuery) (*RepoMetadata, error) {
	client := github.NewClient(tc)

	repoOwner := codeResults.GetRepository().GetOwner().GetLogin()
	repoName := codeResults.GetRepository().GetName()

	version, err := rq.VersionParser.ParseVersion(codeResults)
	if err != nil {
		version = "N/A"
	}

	repo, _, err := client.Repositories.Get(ctx, repoOwner, repoName)
	if err != nil {
		return nil, err
	}

	totalCommits := 0
	commits, _, err := client.Repositories.ListContributorsStats(ctx, repoOwner, repoName)
	if err != nil {
		totalCommits = -1
	}
	for _, r := range commits {
		totalCommits = totalCommits + r.GetTotal()
	}

	repoDetails := &RepoMetadata{
		URL:          codeResults.GetRepository().GetURL(),
		Name:         repoName,
		Owner:        repoOwner,
		ProjectType:  rq.ProjectType,
		Version:      version,
		Stars:        repo.GetStargazersCount(),
		CreatedAt:    repo.GetCreatedAt().String(),
		PushedAt:     repo.GetPushedAt().String(),
		TotalCommits: totalCommits,
	}
	return repoDetails, nil
}
