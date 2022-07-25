// Package main provides an utility for list your github followers
// and the ones recently unfollowed you.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	gh "github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

var user = flag.String("user", "hitzhangjie", "github user (specify shell var GITHUB_ACCESS_TOKEN before running")

func init() {
	flag.Parse()
}

func main() {
	client := gh.NewClient(authenticatedClient())

	var followers []*gh.User

	var pageIdx int = 1
	var pageSize int = 10

	for {
		users, resp, err := client.Users.ListFollowers(context.TODO(), *user, &gh.ListOptions{
			Page:    pageIdx,
			PerPage: pageSize,
		})
		if err != nil {
			if _, ok := err.(*gh.RateLimitError); ok {
				fmt.Println("hit rate limit...wait 1s")
				time.Sleep(time.Second)
				continue
			}
			fmt.Println("list followers fail: %v", err)
			os.Exit(1)
		}
		followers = append(followers, users...)
		if resp.NextPage == 0 {
			break
		}
		pageIdx = resp.NextPage
	}

	fmt.Printf("%s has %d followers:\n", *user, len(followers))
	fmt.Printf("---------------------------------------------")
	prettyprint(followers)
}

func authenticatedClient() *http.Client {
	token := os.Getenv("GITHUB_ACCESS_TOKEN")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.TODO(), ts)

	return gh.NewClient(tc).Client()
}

func prettyprint(users []*gh.User) {
	for i, u := range users {
		fmt.Printf("%03d. %s\n", i+1, u.GetLogin())
	}
}
