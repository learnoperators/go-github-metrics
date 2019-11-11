package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/learnoperators/go-github-metrics/pkg/sdkstats"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type helmVersionParser struct{}
type baseVersionParser struct {
	searchQ      string
	searchLatest string
}
type gomodVersionParser struct {
	searchQ string
}
type unknownVersionParser struct{}

func main() {
	var token string
	flag.StringVar(&token, "token", "", "GitHub API token")
	flag.Parse()
	if len(token) == 0 {
		log.Fatal("GITHUB API TOKEN MUST BE ENTERED \n Usage: './main --token=YOURTOKEN'")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := sdkstats.Client{Client: github.NewClient(tc)}

	queries := []sdkstats.RepoMetadataQuery{
		sdkstats.RepoMetadataQuery{
			ProjectType: "helm",
			Queries:     []string{"filename:Dockerfile quay.io/operator-framework/helm-operator"},
			VersionParser: &baseVersionParser{
				searchQ: "quay.io/operator-framework/helm-operator",
			},
		},
		sdkstats.RepoMetadataQuery{
			ProjectType: "ansible",
			Queries:     []string{"filename:Dockerfile quay.io/operator-framework/ansible-operator"},
			VersionParser: &baseVersionParser{
				searchQ: "quay.io/operator-framework/ansible-operator",
			},
		},
		/*sdkstats.RepoMetadataQuery{
			ProjectType: "go.mod",
			Queries: []string{
				"filename:go.mod github.com/operator-framework/operator-sdk",
			},
			VersionParser: &gomodVersionParser{
				searchQ: "replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v",
			},
		},
		/*sdkstats.RepoMetadataQuery{
			ProjectType:   "gopkg.toml",
			Queries:       []string{"filename:Gopkg.toml github.com/operator-framework/operator-sdk"},
			VersionParser: &unknownVersionParser{},
		},*/
	}
	// GetStats function for Query String from 'queries', These Strings are specific to Operator-SDK patterns.
	collectStats := [][]sdkstats.RepoMetadata{}

	for _, r := range queries {
		stats, err := client.GetStats(ctx, r)

		fmt.Println("Total count for ", r.ProjectType, ":", len(stats))
		if err != nil {
			fmt.Printf("Failed to get stats for queries %v: %v\n", r.Queries, err)
		}
		collectStats = append(collectStats, stats)
	}
	fileName := "Search_Results.json"
	file, err := json.MarshalIndent(collectStats, "", " ")
	err = ioutil.WriteFile(fileName, file, 0644)
	if err != nil {
		fmt.Printf("Error encountered while writing results to file %v\n", err)
		fmt.Println(collectStats)
		os.Exit(1)
	}
	fmt.Println("Results are written in Search_Results.json")
}

// Parse the given Code result to search Text Matches for Version number.
func (p baseVersionParser) ParseVersion(codeResults github.CodeResult) (string, error) {
	baseImageRegex := regexp.QuoteMeta(p.searchQ)
	versionRegex := strings.Trim(`(:([^s]+\n))?`, "\n")
	re := regexp.MustCompile(baseImageRegex + versionRegex)
	for _, r := range codeResults.TextMatches {
		matches := re.FindStringSubmatch(r.GetFragment())
		if matches != nil {
			if matches[1] == "" {
				return "latest", nil
			} else {
				return strings.Trim(matches[2], "\n"), nil
			}
		}

	}
	return "unknown", nil
}

// Parse the given Code result to search Text Matches for Version number.
/*func (p baseVersionParser) ParseVersion(codeResults github.CodeResult) (string, error) {
	var version string
	var s, v []string

	for _, r := range codeResults.TextMatches {
		if strings.Contains(r.GetFragment(), p.searchQ) {
			value := r.GetFragment()
			s = strings.Split(value, "\n")
		stLoop:
			for _, st := range s {
				if strings.Contains(st, p.searchQ) {
					v = strings.Split(st, ":")
					if len(v) == 0 {
						version = "unknown"
					} else if len(v) == 1 {
						version = "latest"
					} else {
						version = v[1]
					}
					break stLoop
				}
			}
		}
	}
	if version == "" {
		version = "unknown"
	}
	return version, nil
}*/

// Parse the given Code result to search Text Matches for Version number.
func (p gomodVersionParser) ParseVersion(codeResults github.CodeResult) (string, error) {
	var version string
	var s, v []string

	for _, r := range codeResults.TextMatches {
		if strings.Contains(r.GetFragment(), p.searchQ) {
			value := r.GetFragment()
			s = strings.Split(value, "\n")
		stLoop:
			for _, st := range s {
				if strings.Contains(st, p.searchQ) {
					v = strings.Split(st, " v")
					if len(v) == 0 {
						version = "unknown"
					} else {
						version = v[1]
					}
					break stLoop
				}
			}
		}
	}
	return version, nil
}

// Parse the given Code result to search Text Matches for Version number.
func (p unknownVersionParser) ParseVersion(codeResults github.CodeResult) (string, error) {
	return "unknown", nil
}
