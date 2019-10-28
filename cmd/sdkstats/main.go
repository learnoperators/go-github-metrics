package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/go-github-metrics-1/pkg/sdkstats"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	var token string
	flag.StringVar(&token, "token", "", "GitHub API token")
	flag.Parse()
	args := flag.Args()
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: args[0]},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var queries []string

	queries = append(queries, "filename:Dockerfile quay.io/operator-framework/helm-operator")
	queries = append(queries, "filename:Dockerfile quay.io/operator-framework/ansible-operator")
	queries = append(queries, "filename:go.mod github.com/operator-framework/operator-sdk")
	queries = append(queries, "filename:Gopkg.toml github.com/operator-framework/operator-sdk")

	for _, r := range queries {
		stats, err := sdkstats.GetStats(client, r)
		if len(stats) == 0 {
			log.Fatal("GITHUB API is down at time, Please try after few minutes")
		} else {
			fileName := stats[0].ProjectType + ".json"
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Total count: ", len(stats))
			file, _ := json.MarshalIndent(stats, "", " ")
			_ = ioutil.WriteFile(fileName, file, 0644)
			fmt.Println("Results are written in ", fileName)
		}
	}

}
