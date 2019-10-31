package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/go-github-metrics-1/pkg/sdkstats"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type projectVersionParser struct{}

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
	client := github.NewClient(tc)

	queries := []sdkstats.RepoMetadataQuery{
		/*sdkstats.RepoMetadataQuery{
			ProjectType:   "ansible",
			Queries:       []string{"filename:Dockerfile quay.io/operator-framework/ansible-operator"},
			VersionParser: &projectVersionParser{},
		},
		sdkstats.RepoMetadataQuery{
			ProjectType:   "helm",
			Queries:       []string{"filename:Dockerfile quay.io/operator-framework/helm-operator"},
			VersionParser: &projectVersionParser{},
		},*/
		sdkstats.RepoMetadataQuery{
			ProjectType: "go",
			Queries: []string{
				"filename:go.mod github.com/operator-framework/operator-sdk",
				"filename:Gopkg.toml github.com/operator-framework/operator-sdk",
			},
			VersionParser: &projectVersionParser{},
		},
	}
	//GetStats function for Query String from 'queries', These Strings are specific to Operator-SDK patterns.
	for _, r := range queries {
		stats, err := sdkstats.GetStats(client, r)
		if _, ok := err.(*github.AcceptedError); ok {
			log.Println("Job is scheduled on GitHub side")
		} else if _, ok := err.(*github.RateLimitError); ok {
			log.Println("Rate Limit has reached.")
		} else if err != nil {
			fmt.Println("Failed to get Stats for Query", r, err)
		}
		//Write results into JSON , named with ProjectType.
		fileName := r.ProjectType + ".json"
		fmt.Println("Total count: ", len(stats))
		file, _ := json.MarshalIndent(stats, "", " ")
		_ = ioutil.WriteFile(fileName, file, 0644)
		fmt.Println("Results are written in ", fileName)
	}

}

//Parse the given Code result to search Text Matches for Version number.
func (p projectVersionParser) ParseVersion(codeResults github.CodeResult, projectType string) (string, error) {
	var searchQ, sdkVersion, searchLatest string
	sdkLatest := "11"
	var posFirstAdjusted, c int
	//Text match strings for Helm/Ansible are always from the First line of Docker file, Hence fixing the Version Indices.
	if projectType == "helm" {
		searchQ = "quay.io/operator-framework/helm-operator:v0."
		searchLatest = "quay.io/operator-framework/helm-operator:latest"
		posFirstAdjusted = 49
		c = 51
	} else if projectType == "ansible" {
		searchQ = "quay.io/operator-framework/ansible-operator:v0."
		searchLatest = "quay.io/operator-framework/ansible-operator:master"
		posFirstAdjusted = 52
		c = 54
	}
	if projectType != "go" {
		for _, r := range codeResults.TextMatches {
			if strings.Contains(r.GetFragment(), searchQ) {
				runes := []rune(r.GetFragment())
				sdkVersion = strings.Trim(string(runes[posFirstAdjusted:c]), ".")
				if _, err := strconv.Atoi(sdkVersion); err != nil {
					sdkVersion = "N/A"
				}
			} else if strings.Contains(r.GetFragment(), searchLatest) {
				sdkVersion = sdkLatest
			}
		}
	}
	if projectType == "go" {
		searchQ = "github.com/operator-framework/operator-sdk v0."
		for _, r := range codeResults.TextMatches {
			if strings.Contains(r.GetFragment(), searchQ) {
				posFirst := strings.Index(r.GetFragment(), searchQ)
				if posFirst == -1 {
					sdkVersion = "N/A"
				}
				posFirstAdjusted := posFirst + len(searchQ)
				runes := []rune(r.GetFragment())
				c := posFirstAdjusted + 2
				if c == -1 {
					sdkVersion = "N/A"
				}
				sdkVersion = strings.Trim(string(runes[posFirstAdjusted:c]), ".")
				if _, err := strconv.Atoi(sdkVersion); err != nil {
					sdkVersion = "N/A"
				}
			}
		}
	}
	return sdkVersion, nil
}
