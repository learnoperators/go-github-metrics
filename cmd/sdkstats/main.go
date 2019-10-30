package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
		log.Fatal("GITHUB API TOKEN MUST BE ENTERED \n Usage: './main. --token=YOURTOKEN'")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	queries := []sdkstats.RepoMetadataQuery{
		sdkstats.RepoMetadataQuery{
			ProjectType:   "ansible",
			Queries:       []string{"filename:Dockerfile quay.io/operator-framework/ansible-operator"},
			VersionParser: &projectVersionParser{},
		},
		sdkstats.RepoMetadataQuery{
			ProjectType:   "helm",
			Queries:       []string{"filename:Dockerfile quay.io/operator-framework/helm-operator"},
			VersionParser: &projectVersionParser{},
		},
		sdkstats.RepoMetadataQuery{
			ProjectType: "go",
			Queries: []string{
				"filename:go.mod github.com/operator-framework/operator-sdk",
				"filename:Gopkg.toml github.com/operator-framework/operator-sdk",
			},
			VersionParser: &projectVersionParser{},
		},
	}
	//Call GetStats function for wah Query String from 'queries'.
	//These Strings are specific to Operator-SDK patterns.
	for _, r := range queries {
		stats, err := sdkstats.GetStats(client, r)
		if err != nil {
			fmt.Printf("failed to get stats for query %q: %w", r, err)
			os.Exit(1)
		}
		fileName := r.ProjectType + ".json"
		fmt.Println("Total count: ", len(stats))
		file, _ := json.MarshalIndent(stats, "", " ")
		_ = ioutil.WriteFile(fileName, file, 0644)
		fmt.Println("Results are written in ", fileName)
	}
}

func (p projectVersionParser) ParseVersion(codeResults github.CodeResult, projectType string) (string, error) {
	// parse the version, assuming a helm result. if a version isn't found, return an error
	var searchQ, sdkVersion, searchLatest string
	sdkLatest := "11"

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

		} else if (projectType == "helm" || projectType == "ansible") && strings.Contains(r.GetFragment(), searchLatest) {
			sdkVersion = sdkLatest
		}
	}
	return sdkVersion, nil
}
