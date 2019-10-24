package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

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
		op, resp, err := client.Search.Code(context, "github.com/operator-framework/operator-sdk filename:go.mod", opts)
		if err != nil {
			fmt.Println(err)
		}
		goOp = append(goOp, op)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	golangOp, _, err := client.Search.Code(context, "github.com/operator-framework/operator-sdk filename:Gopkg.toml", opts)
	if err != nil {
		fmt.Println(err)
	}

	//HELM based operator-sdk project breakdown count, per version between 0.1 - 0.11
	helmVersion := map[int]int{}
	var helmCnt int
	for i := 0; i < len(helmOp); i++ {
		for _, r := range helmOp[i].CodeResults {
			helmCnt++
		helmloop:
			for j := 1; j <= 11; j++ {
				for _, r := range r.TextMatches {
					if strings.Contains(r.GetFragment(), "quay.io/operator-framework/helm-operator:v0."+strconv.Itoa(j)+".") {
						helmVersion[j] = helmVersion[j] + 1
						break helmloop
					}
				}
			}
		}
	}

	//ANSIBLE based operator-sdk project breakdown count, per version between 0.1 - 0.11

	ansibleVersion := map[int]int{}
	var ansibleCnt int
	for k := 0; k < len(ansibleOp); k++ {
		for i := 0; i < len(ansibleOp[k].CodeResults); i++ {
			ansibleCnt++
		ansibleLoop:
			for j := 1; j <= 11; j++ {
				for _, r := range ansibleOp[k].CodeResults[i].TextMatches {
					if strings.Contains(r.GetFragment(), "quay.io/operator-framework/ansible-operator:v0."+strconv.Itoa(j)+".") {
						ansibleVersion[j] = ansibleVersion[j] + 1
						break ansibleLoop
					}
				}
			}
		}
	}

	//GO based operator-sdk(go.mod) project breakdown count, per version between 0.1 - 0.11
	goVersion := map[int]int{}
	var goCnt int
	for k := 0; k < len(goOp); k++ {
		for i := 0; i < len(goOp[k].CodeResults); i++ {
			goCnt++
		goLoop:
			for j := 1; j <= 11; j++ {
				for _, r := range goOp[k].CodeResults[i].TextMatches {
					if strings.Contains(r.GetFragment(), "github.com/operator-framework/operator-sdk v0."+strconv.Itoa(j)+".") {
						goVersion[j] = goVersion[j] + 1
						break goLoop
					}
				}
			}
		}
	}

	fmt.Println("HELM based project Breakdown count by version between 0.1 - 0.11 \n", helmVersion)
	fmt.Println("ANSIBLE based project Breakdown count by version between 0.1 - 0.11 \n", ansibleVersion)
	fmt.Println("GO based project(go.mod) Breakdown count by version between 0.1 - 0.11 \n", goVersion)

	//Breakdown of Code Count for Project Type GO/HELM/ANSIBLE
	fmt.Println(" Ansible based Code Count: ", ansibleCnt, "\n Helm based Code Count: ", helmCnt, "\n GO based Code Count: ", golangOp.GetTotal()+goCnt)

}
